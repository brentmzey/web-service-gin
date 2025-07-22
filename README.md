# web-service-gin

## Getting Started

### Prerequisites

- Go 1.13+

### Installing

```bash
go get github.com/gin-gonic/gin
```

## Usage

```go
package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func main() {
	router := gin.Default()
	router.GET("/albums", func(c *gin.Context) {
		c.IndentedJSON(http.StatusOK, gin.H{
			"message": "Hello, World!",
		})
	})
	
	err := router.Run(":8080")
	if err != nil {
		panic(err)
	}
}
```

## Example API calls

### Get all albums

```bash
curl http://localhost:8080/albums
```

### Get album by ID

```bash
curl http://localhost:8080/albums/1
```

### Add album

```bash
curl -X POST -H "Content-Type: application/json" -d '{"id": "4", "title": "The Beatles", "artist": "The Beatles", "price": 12.99}' http://localhost:8080/albums
```

## License

MIT License - see the [LICENSE](LICENSE) file for details
