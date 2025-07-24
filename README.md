# web-service-go

A simple RESTful web service for managing albums, built with Go's standard `net/http` package (no Gin dependency) and using [`github.com/google/uuid`](https://github.com/google/uuid) for unique album IDs.

---

## Features

- List all albums
- Retrieve a single album by ID (UUID)
- Add a new album with unique UUID generation
- Pretty/indented JSON responses
- Request logging with method, path, status, and duration
- **Rate limiting with exponential backoff** for repeated requests

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

## Rate Limiting & Exponential Backoff

This service includes a **rate limiting middleware**. If a client (by IP address) makes more than 5 requests within 15 seconds, further requests are rejected with HTTP 429 ("Too Many Requests"). The rejection is logged, and the suggested wait time increases exponentially (using a backoff algorithm: `waitTime = 2^requestCount` seconds).

**Example log output:**

```
⏳ Rate limit exceeded for 127.0.0.1:54321, waiting 64s
```

**Example response:**

```json
{
  "message": "Too many requests, please wait a bit"
}
```

---

## Using `jq` for Pretty JSON Output

Many examples in this README use [`jq`](https://jqlang.org/) to pretty-print JSON responses.

### Install `jq` on macOS (Homebrew):

```bash
brew install jq
```

### Install `jq` on Debian/Ubuntu:

```bash
sudo apt-get update
sudo apt-get install jq
```

### Other Platforms

You can download binaries for your platform from the official site:  
https://jqlang.org/download/

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
curl http://localhost:8080/albums | jq

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
- **Implements rate limiting with exponential backoff**: If a client exceeds 5 requests in 15 seconds, further requests are blocked and the suggested wait time increases exponentially.
- Uses only Go's standard library (`net/http`, `encoding/json`, etc).

---

## License

MIT License - see the LICENSE file for details.