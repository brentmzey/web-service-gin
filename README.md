# web-service-gin

A simple RESTful web service for managing albums, built with Go's standard `net/http` package (no Gin dependency).

---

## Features

- List all albums
- Retrieve a single album by ID
- Add a new album
- Pretty/indented JSON responses
- Request logging with method, path, status, and duration

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

## API Endpoints & Example Calls

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

## Code Explanation

- **main.go** contains all logic:
  - Defines the `album` struct and a slice of seed albums.
  - Implements handlers for listing, retrieving, and adding albums.
  - Uses a custom `writeJSON` function for pretty JSON output.
  - Adds a logging middleware to log each request's method, path, status, and duration.
  - Uses only Go's standard library (`net/http`, `encoding/json`, etc).

---

## Example Code

```go
package main

import (
    "encoding/json"
    "log"
    "net/http"
    "strings"
    "time"
)

type album struct {
    ID     string  `json:"id"`
    Title  string  `json:"title"`
    Artist string  `json:"artist"`
    Price  float64 `json:"price"`
}

var albums = []album{
    {ID: "1", Title: "Blue Train", Artist: "John Coltrane", Price: 56.99},
    {ID: "2", Title: "Jeru", Artist: "Gerry Mulligan", Price: 17.99},
    {ID: "3", Title: "Sarah Vaughan and Clifford Brown", Artist: "Sarah Vaughan", Price: 39.99},
}

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

func loggingMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        start := time.Now()
        lrw := &loggingResponseWriter{ResponseWriter: w, statusCode: http.StatusOK}
        next.ServeHTTP(lrw, r)
        duration := time.Since(start)
        log.Printf("%s %s %d %s", r.Method, r.URL.Path, lrw.statusCode, duration)
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

func getAlbums(w http.ResponseWriter, r *http.Request) {
    writeJSON(w, http.StatusOK, albums)
}

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

func postAlbums(w http.ResponseWriter, r *http.Request) {
    var newAlbum album
    if err := json.NewDecoder(r.Body).Decode(&newAlbum); err != nil {
        writeJSON(w, http.StatusBadRequest, map[string]string{"message": err.Error()})
        return
    }
    albums = append(albums, newAlbum)
    writeJSON(w, http.StatusCreated, newAlbum)
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
```

---

## License

MIT License - see the [LICENSE](LICENSE) file for details.