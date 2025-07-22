# web-service-gin

A simple RESTful web service for managing albums, built with [Gin](https://github.com/gin-gonic/gin) in Go.

## Features

- List all albums
- Retrieve a single album by ID
- Add a new album

## Prerequisites

- Go 1.13 or newer

## Installation

Clone the repository and install dependencies:

```bash
git clone https://github.com/brentmzey/web-service-gin.git
cd web-service-gin
go mod tidy
```

## Running the Service

Start the server:

```bash
go run main.go
```

The service will be available at [http://localhost:8080](http://localhost:8080).

## API Endpoints

### Get all albums

- **Endpoint:** `GET /albums`
- **Response:** JSON array of all albums

```bash
curl http://localhost:8080/albums
```

### Get album by ID

- **Endpoint:** `GET /albums/:id`
- **Response:** JSON object of the album, or 404 if not found

```bash
curl http://localhost:8080/albums/1
```

### Add a new album

- **Endpoint:** `POST /albums`
- **Request Body:** JSON object with `id`, `title`, `artist`, and `price`
- **Response:** JSON object of the created album

```bash
curl -X POST -H "Content-Type: application/json" \
  -d '{"id": "4", "title": "The Beatles", "artist": "The Beatles", "price": 12.99}' \
  http://localhost:8080/albums
```

## Example Album Object

```json
{
  "id": "1",
  "title": "Blue Train",
  "artist": "John Coltrane",
  "price": 56.99
}
```

## Project Structure

- [`main.go`](main.go): Main application source code
- [`go.mod`](go.mod): Go module definition

## License

MIT License - see the [LICENSE](LICENSE) file for details.