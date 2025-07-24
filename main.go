package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
)

// album represents data about a record album.
type album struct {
	ID     string  `json:"id"`
	Title  string  `json:"title"`
	Artist string  `json:"artist"`
	Price  float64 `json:"price"`
}

// albums slice to seed record album data with more entries.
var albums = []album{
	{ID: uuid.New().String(), Title: "Blue Train", Artist: "John Coltrane", Price: 56.99},
	{ID: uuid.New().String(), Title: "Jeru", Artist: "Gerry Mulligan", Price: 17.99},
	{ID: uuid.New().String(), Title: "Sarah Vaughan and Clifford Brown", Artist: "Sarah Vaughan", Price: 39.99},
	{ID: uuid.New().String(), Title: "The Dark Side of the Moon", Artist: "Pink Floyd", Price: 22.99},
	{ID: uuid.New().String(), Title: "Abbey Road", Artist: "The Beatles", Price: 29.99},
	{ID: uuid.New().String(), Title: "Kind of Blue", Artist: "Miles Davis", Price: 9.99},
	{ID: uuid.New().String(), Title: "Back to Black", Artist: "Amy Winehouse", Price: 19.99},
	{ID: uuid.New().String(), Title: "Nevermind", Artist: "Nirvana", Price: 15.99},
	{ID: uuid.New().String(), Title: "The Wall", Artist: "Pink Floyd", Price: 25.99},
	{ID: uuid.New().String(), Title: "Rumours", Artist: "Fleetwood Mac", Price: 18.99},
	{ID: uuid.New().String(), Title: "A Love Supreme", Artist: "John Coltrane", Price: 19.99},
	{ID: uuid.New().String(), Title: "In the Wee Small Hours", Artist: "Frank Sinatra", Price: 12.99},
}

// clientInfo holds request tracking information.
type clientInfo struct {
	lastRequest  time.Time
	requestCount int
}

// clients map maintains client request info.
var clients = make(map[string]*clientInfo)

// writeJSON writes pretty JSON with status code and logs errors if any.
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

// loggingMiddleware logs method, path, status, and duration for each request with emojis.
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		lrw := &loggingResponseWriter{ResponseWriter: w, statusCode: http.StatusOK}
		next.ServeHTTP(lrw, r)
		duration := time.Since(start)
		log.Printf("🚀 %s %s -> %d %s 🌟", r.Method, r.URL.Path, lrw.statusCode, duration)
	})
}

// loggingResponseWriter captures status code for logging.
type loggingResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (lrw *loggingResponseWriter) WriteHeader(code int) {
	lrw.statusCode = code
	lrw.ResponseWriter.WriteHeader(code)
}

// rateLimitingMiddleware enforces rate limits on requests.
func rateLimitingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		clientIP := r.RemoteAddr

		if info, exists := clients[clientIP]; exists {
			if time.Since(info.lastRequest) < 15*time.Second {
				info.requestCount++
				if info.requestCount > 5 {
					waitTime := time.Duration(1<<info.requestCount) * time.Second
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

// getAlbums responds with the list of all albums as pretty JSON.
func getAlbums(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, albums)
	log.Println("🎶 Fetched all albums")
}

// getAlbumByID locates the album whose ID matches the id parameter in the request.
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

// postAlbums adds an album from JSON received in the request body.
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

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/albums", albumsHandler)
	mux.HandleFunc("/albums/", albumByIDHandler)
	log.Println("🎧 Listening on http://localhost:8080")

	wrappedMux := loggingMiddleware(rateLimitingMiddleware(mux))
	log.Fatal(http.ListenAndServe("localhost:8080", wrappedMux))
}
