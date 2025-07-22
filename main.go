# web-service-gin

A simple RESTful web service for managing albums, built with Go's standard `net/http` package.

---

## Features

- **List all albums**
- **Retrieve a single album by ID**
- **Add a new album**

---

## Prerequisites

- Go 1.13 or newer

---

## Installation

Clone the repository:

```bash
git clone https://github.com/brentmzey/web-service-gin.git
cd web-service-gin
```

---

## Running the Service

Start the server:

```bash
go run main.go
```

The service will be available at [http://localhost:8080](http://localhost:8080).

---

## API Endpoints

### Get all albums

- **Endpoint:** `GET /albums`
- **Response:** JSON array of all albums

**Example:**
```bash
curl http://localhost:8080/albums
```

---

### Get album by ID

- **Endpoint:** `GET /albums/:id`
- **Response:** JSON object of the album, or 404 if not found

**Example:**
```bash
curl http://localhost:8080/albums/1
```

---

### Add a new album

- **Endpoint:** `POST /albums`
- **Request Body:** JSON object with `id`, `title`, `artist`, and `price`
- **Response:** JSON object of the created album

**Example:**
```bash
curl -X POST -H "Content-Type: application/json" \
  -d '{"id": "4", "title": "The Beatles", "artist": "The Beatles", "price": 12.99}' \
  http://localhost:8080/albums
```

---

## Example Album Object

```json
{
  "id": "1",
  "title": "Blue Train",
  "artist": "John Coltrane",
  "price": 56.99
}
```

---

## Project Structure

- `main.go`: Main application source code
- `go.mod`: Go module definition

---

## License

MIT License - see the [LICENSE](LICENSE) file for details.

---

## Example Code

```go
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
```