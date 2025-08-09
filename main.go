package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v4" // PostgreSQL
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"gorm.io/driver/sqlite" // SQLite
	"gorm.io/gorm"          // ORM for SQLite
)

// album represents data about a record album.
type album struct {
	ID     string  `json:"id"`
	Title  string  `json:"title"`
	Artist string  `json:"artist"`
	Price  float64 `json:"price"`
}

var albums = []album{
	{ID: uuid.New().String(), Title: "Blue Train", Artist: "John Coltrane", Price: 56.99},
	{ID: uuid.New().String(), Title: "Jeru", Artist: "Gerry Mulligan", Price: 17.99},
	{ID: uuid.New().String(), Title: "Sarah Vaughan and Clifford Brown", Artist: "Sarah Vaughan", Price: 39.99},
	// Add more albums...
}

type clientInfo struct {
	lastRequest  time.Time
	requestCount int
}

var clients = make(map[string]*clientInfo)

type Metrics struct {
	TotalRequests      int64
	TotalErrors        int64
	TotalAlbumsFetched int64
	TotalAlbumsAdded   int64
	TotalRateLimited   int64
	TotalLatencyMs     int64
}

var metrics = &Metrics{}

type MetricsStore interface {
	SaveMetrics(metrics Metrics) error
	LoadMetrics() (Metrics, error)
}

type InMemoryMetricsStore struct {
	metrics Metrics
}

func (store *InMemoryMetricsStore) SaveMetrics(metrics Metrics) error {
	store.metrics = metrics
	return nil
}

func (store *InMemoryMetricsStore) LoadMetrics() (Metrics, error) {
	return store.metrics, nil
}

// PostgresMetricsStore is a basic outline for future implementation
type PostgresMetricsStore struct {
	conn *pgx.Conn
}

func NewPostgresMetricsStore(conn *pgx.Conn) *PostgresMetricsStore {
	return &PostgresMetricsStore{conn: conn}
}

func (store *PostgresMetricsStore) SaveMetrics(metrics Metrics) error {
	// Implement PostgreSQL saving logic
	return nil
}

func (store *PostgresMetricsStore) LoadMetrics() (Metrics, error) {
	// Implement PostgreSQL loading logic
	return Metrics{}, nil
}

// Implement similar structures for SQLite, MongoDB, and DynamoDB

type SqliteMetricsStore struct {
	db *gorm.DB
}

func NewSqliteMetricsStore(db *gorm.DB) *SqliteMetricsStore {
	return &SqliteMetricsStore{db: db}
}

func (store *SqliteMetricsStore) SaveMetrics(metrics Metrics) error {
	// Implement SQLite saving logic
	return nil
}

func (store *SqliteMetricsStore) LoadMetrics() (Metrics, error) {
	// Implement SQLite loading logic
	return Metrics{}, nil
}

type MongoMetricsStore struct {
	collection *mongo.Collection
}

func NewMongoMetricsStore(collection *mongo.Collection) *MongoMetricsStore {
	return &MongoMetricsStore{collection: collection}
}

func (store *MongoMetricsStore) SaveMetrics(metrics Metrics) error {
	// Implement MongoDB saving logic
	return nil
}

func (store *MongoMetricsStore) LoadMetrics() (Metrics, error) {
	// Implement MongoDB loading logic
	return Metrics{}, nil
}

type DynamoMetricsStore struct {
	session *dynamodb.DynamoDB
}

func NewDynamoMetricsStore(sess *dynamodb.DynamoDB) *DynamoMetricsStore {
	return &DynamoMetricsStore{session: sess}
}

func (store *DynamoMetricsStore) SaveMetrics(metrics Metrics) error {
	// Implement DynamoDB saving logic
	return nil
}

func (store *DynamoMetricsStore) LoadMetrics() (Metrics, error) {
	// Implement DynamoDB loading logic
	return Metrics{}, nil
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	js, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		http.Error(w, `{"message":"internal server error"}`, http.StatusInternalServerError)
		log.Printf("🔥 JSON marshal error: %v", err)
		return
	}
	w.Write(js)
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		lrw := &loggingResponseWriter{ResponseWriter: w, statusCode: http.StatusOK}
		next.ServeHTTP(lrw, r)
		duration := time.Since(start)
		log.Printf("🚀 %s %s -> %d %s 🌟", r.Method, r.URL.Path, lrw.statusCode, duration)
	})
}

type loggingResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (lrw *loggingResponseWriter) WriteHeader(code int) {
	lrw.statusCode = code
	lrw.ResponseWriter.WriteHeader(code)
}

func metricsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		metrics.TotalRequests++
		lrw := &loggingResponseWriter{ResponseWriter: w, statusCode: http.StatusOK}
		next.ServeHTTP(lrw, r)
		latency := time.Since(start).Milliseconds()
		metrics.TotalLatencyMs += latency
		if lrw.statusCode >= 400 {
			metrics.TotalErrors++
		}
	})
}

func metricsHandler(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"totalRequests":      metrics.TotalRequests,
		"totalErrors":        metrics.TotalErrors,
		"totalAlbumsFetched": metrics.TotalAlbumsFetched,
		"totalAlbumsAdded":   metrics.TotalAlbumsAdded,
		"totalRateLimited":   metrics.TotalRateLimited,
		"averageLatencyMs":   avgLatency(),
	})
}

func avgLatency() int64 {
	if metrics.TotalRequests == 0 {
		return 0
	}
	return metrics.TotalLatencyMs / metrics.TotalRequests
}

func rateLimitingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		clientIP := r.RemoteAddr
		if info, exists := clients[clientIP]; exists {
			if time.Since(info.lastRequest) < 15*time.Second {
				info.requestCount++
				if info.requestCount > 5 {
					waitTime := time.Duration(1<<info.requestCount) * time.Second
					metrics.TotalRateLimited++
					http.Error(w, "Too many requests, please wait a bit", http.StatusTooManyRequests)
					log.Printf("⏳ Rate limit exceeded for %s, waiting %v", clientIP, waitTime)
					return
				}
			} else {
				info.requestCount = 1
			}
			info.lastRequest = time.Now()
		} else {
			clients[clientIP] = &clientInfo{requestCount: 1, lastRequest: time.Now()}
		}
		next.ServeHTTP(w, r)
	})
}

func getAlbums(w http.ResponseWriter, r *http.Request) {
	metrics.TotalAlbumsFetched++
	writeJSON(w, http.StatusOK, albums)
	log.Println("🎶 Fetched all albums")
}

func getAlbumByID(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/albums/")
	for _, a := range albums {
		if a.ID == id {
			writeJSON(w, http.StatusOK, a)
			log.Printf("🔍 Album found: %s", a.Title)
			return
		}
	}
	writeJSON(w, http.StatusNotFound, map[string]string{"message": "album not found"})
	log.Println("❌ Album not found")
}

func postAlbums(w http.ResponseWriter, r *http.Request) {
	var newAlbum struct {
		Title  string  `json:"title"`
		Artist string  `json:"artist"`
		Price  float64 `json:"price"`
	}
	if err := json.NewDecoder(r.Body).Decode(&newAlbum); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"message": err.Error()})
		log.Println("📉 Bad request:", err)
		return
	}

	album := album{
		ID:     uuid.New().String(),
		Title:  newAlbum.Title,
		Artist: newAlbum.Artist,
		Price:  newAlbum.Price,
	}

	albums = append(albums, album)
	metrics.TotalAlbumsAdded++
	writeJSON(w, http.StatusCreated, album)
	log.Printf("✨ New album added: %s by %s", album.Title, album.Artist)
}

func albumsHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		getAlbums(w, r)
	case http.MethodPost:
		postAlbums(w, r)
	default:
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"message": "Method not allowed"})
		log.Println("🔒 Method not allowed")
	}
}

func albumByIDHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		getAlbumByID(w, r)
	} else {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"message": "Method not allowed"})
		log.Println("🔒 Method not allowed")
	}
}

func setupMetricsStore() MetricsStore {
	dbType := os.Getenv("DB_TYPE")
	switch dbType {
	case "postgres":
		conn, err := pgx.Connect(context.Background(), os.Getenv("DATABASE_URL"))
		if err != nil {
			log.Fatalf("Unable to connect to database: %v", err)
		}
		return NewPostgresMetricsStore(conn)

	case "sqlite":
		db, err := gorm.Open(sqlite.Open("file:metrics.db?cache=shared&_fk=1"), &gorm.Config{})
		if err != nil {
			log.Fatalf("Failed to connect to SQLite database: %v", err)
		}
		return NewSqliteMetricsStore(db)

	case "mongodb":
		client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI("mongodb://localhost:27017"))
		if err != nil {
			log.Fatalf("Failed to connect to MongoDB: %v", err)
		}
		return NewMongoMetricsStore(client.Database("metricsDb").Collection("metrics"))

	case "dynamodb":
		sess := session.Must(session.NewSession())
		svc := dynamodb.New(sess)
		return NewDynamoMetricsStore(svc)

	default:
		return &InMemoryMetricsStore{}
	}
}

var metricsStore MetricsStore

func main() {
	metricsStore = setupMetricsStore()
	mux := http.NewServeMux()
	mux.HandleFunc("/albums", albumsHandler)
	mux.HandleFunc("/albums/", albumByIDHandler)
	mux.HandleFunc("/metrics", metricsHandler)
	log.Println("🎧 Listening on http://localhost:8080")

	wrappedMux := metricsMiddleware(loggingMiddleware(rateLimitingMiddleware(mux)))
	log.Fatal(http.ListenAndServe("localhost:8080", wrappedMux))
}
