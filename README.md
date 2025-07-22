# web-service-go

A simple RESTful web service for managing albums, built with Go's standard `net/http` package (no Gin dependency) and using [`github.com/google/uuid`](https://github.com/google/uuid) for unique album IDs.

---

## Features

- List all albums
- Retrieve a single album by ID (UUID)
- Add a new album with unique UUID generation
- Pretty/indented JSON responses
- Request logging with method, path, status, and duration

---

## Prerequisites

- Go 1.13 or newer

---

## Installation

Clone the repository and install dependencies:

```bash
git clone https://github.com/brentmzey/web-service-go.git
cd web-service-go
go get github.com/google/uuid
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

### Get album by ID (UUID)

- **Endpoint:** `GET /albums/:id`
- **Response:** JSON object of the album, or 404 if not found

**Example:**

```bash
# First, get all albums to find a UUID
curl http://localhost:8080/albums

# Suppose the output includes:
# [
#   {
#     "id": "b1e29e7a-1c2d-4c5e-8e7a-2f3b4c5d6e7f",
#     "title": "Blue Train",
#     "artist": "John Coltrane",
#     "price": 56.99
#   }
# ]

# Now fetch by UUID:
curl http://localhost:8080/albums/b1e29e7a-1c2d-4c5e-8e7a-2f3b4c5d6e7f
```

---

### Add a new album

- **Endpoint:** `POST /albums`
- **Request Body:** JSON object with `title`, `artist`, and `price`
- **Response:** JSON object of the created album with a unique UUID

**Examples:**

```bash
curl -X POST -H "Content-Type: application/json" \
  -d '{"title": "The Beatles", "artist": "The Beatles", "price": 12.99}' \
  http://localhost:8080/albums

curl -X POST -H "Content-Type: application/json" \
  -d '{"title": "Kind Of Blue", "artist": "Miles Davis", "price": 9.99}' \
  http://localhost:8080/albums

curl -X POST -H "Content-Type: application/json" \
  -d '{"title": "A Love Supreme", "artist": "John Coltrane", "price": 19.99}' \
  http://localhost:8080/albums
```

---

### More Example Usage

#### List all albums (pretty print with jq):

```bash
curl http://localhost:8080/albums | jq
```

#### Add multiple albums and capture their UUIDs:

```bash
curl -s -X POST -H "Content-Type: application/json" \
  -d '{"title": "In Rainbows", "artist": "Radiohead", "price": 14.99}' \
  http://localhost:8080/albums | jq '.id'

curl -s -X POST -H "Content-Type: application/json" \
  -d '{"title": "Abbey Road", "artist": "The Beatles", "price": 15.99}' \
  http://localhost:8080/albums | jq '.id'
```

#### Fetch an album by UUID (replace `<uuid>` with the actual UUID):

```bash
curl http://localhost:8080/albums/<uuid>
```

#### Try to fetch a non-existent album (should return 404):

```bash
curl http://localhost:8080/albums/00000000-0000-0000-0000-000000000000
```

#### Send invalid JSON (should return 400):

```bash
curl -X POST -H "Content-Type: application/json" \
  -d '{"title": "Bad JSON", "artist": "Nobody", "price": "not-a-number"}' \
  http://localhost:8080/albums
```

#### Send a request with a missing field (should return 201, but with zero value for missing):

```bash
curl -X POST -H "Content-Type: application/json" \
  -d '{"title": "Missing Artist", "price": 10.00}' \
  http://localhost:8080/albums
```

#### Use HTTPie for a more readable output:

```bash
http POST :8080/albums title="OK Computer" artist="Radiohead" price:=13.99
```

#### List all albums and extract all UUIDs:

```bash
curl -s http://localhost:8080/albums | jq '.[].id'
```

---

## Example Album Object

```json
{
  "id": "b1e29e7a-1c2d-4c5e-8e7a-2f3b4c5d6e7f",
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

`main.go` contains all logic:

- Defines the album struct and a slice of seed albums (with UUIDs).
- Implements handlers for listing, retrieving, and adding albums.
- Generates unique IDs for new albums using `github.com/google/uuid`.
- Uses a custom `writeJSON` function for pretty JSON output.
- Adds a logging middleware to log each request's method, path, status, and duration.
- Uses only Go's standard library (`net/http`, `encoding/json`, etc).

---

## License

MIT License - see the LICENSE file for details.

---

## `main.go`

```go
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
```