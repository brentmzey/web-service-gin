package main

import (
	"encoding/json"
	"log"
	"net/http"
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
	{ID: "1", Title: "Blue Train", Artist: "John Coltrane", Price: 56.99},
	{ID: "2", Title: "Jeru", Artist: "Gerry Mulligan", Price: 17.99},
	{ID: "3", Title: "Sarah Vaughan and Clifford Brown", Artist: "Sarah Vaughan", Price: 39.99},
}

// getAlbumByID locates the album whose ID matches the id parameter in the request.
func getAlbumByID(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Path[len("/albums/"):]

	for _, a := range albums {
		if a.ID == id {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(a)
			return
		}
	}

	http.Error(w, "album not found", http.StatusNotFound)
}

// getAlbums responds with the list of all albums as JSON.
func getAlbums(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(albums)
}

// postAlbums adds an album from JSON received in the request body.
func postAlbums(w http.ResponseWriter, r *http.Request) {
	var newAlbum album
	if err := json.NewDecoder(r.Body).Decode(&newAlbum); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	albums = append(albums, newAlbum)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(newAlbum)
}

func main() {
	http.HandleFunc("/albums", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			getAlbums(w, r)
		case http.MethodPost:
			postAlbums(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	http.HandleFunc("/albums/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			getAlbumByID(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	log.Fatal(http.ListenAndServe("localhost:8080", nil))
}
