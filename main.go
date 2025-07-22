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

// albums slice to seed record album data.
var albums = []album{
	{ID: uuid.New().String(), Title: "Blue Train", Artist: "John Coltrane", Price: 56.99},
	{ID: uuid.New().String(), Title: "Jeru", Artist: "Gerry Mulligan", Price: 17.99},
	{ID: uuid.New().String(), Title: "Sarah Vaughan and Clifford Brown", Artist: "Sarah Vaughan", Price: 39.99},
}

// writeJSON writes pretty JSON with status code and logs errors if any.
func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	js, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		http.Error(w, `{"message":"internal server error"}`, http.StatusInternalServerError)
		log.Printf("JSON marshal error: %v", err)
		return
	}
	w.Write(js)
}

// loggingMiddleware logs method, path, status, and duration for each request.
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		lrw := &loggingResponseWriter{ResponseWriter: w, statusCode: http.StatusOK}
		next.ServeHTTP(lrw, r)
		duration := time.Since(start)
		log.Printf("%s %s %d %s", r.Method, r.URL.Path, lrw.statusCode, duration)
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

// getAlbums responds with the list of all albums as pretty JSON.
func getAlbums(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, albums)
}

// getAlbumByID locates the album whose ID matches the id parameter in the request.
func getAlbumByID(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/albums/")
	for _, a := range albums {
		if a.ID == id {
			writeJSON(w, http.StatusOK, a)
			return
		}
	}
	writeJSON(w, http.StatusNotFound, map[string]string{"message": "album not found"})
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
}

func albumsHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		getAlbums(w, r)
	case http.MethodPost:
		postAlbums(w, r)
	default:
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"message": "Method not allowed"})
	}
}

func albumByIDHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		getAlbumByID(w, r)
	} else {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"message": "Method not allowed"})
	}
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/albums", albumsHandler)
	mux.HandleFunc("/albums/", albumByIDHandler)
	log.Println("Listening on http://localhost:8080")
	log.Fatal(http.ListenAndServe("localhost:8080", loggingMiddleware(mux)))
}
