package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
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

func main() {
	router := gin.Default()
	router.GET("/albums", getAlbums)
	router.GET("/albums/:id", getAlbumByID)
	router.POST("/albums", postAlbums)

	router.Run(":8080")
}

// getAlbums responds with the list of all albums as JSON.
func getAlbums(c *gin.Context) {
	c.IndentedJSON(http.StatusOK, albums)
}

// postAlbums adds an album from JSON received in the request body.
func postAlbums(c *gin.Context) {
	var newAlbum album

	// Call BindJSON to bind the received JSON to
	// newAlbum.
	if err := c.BindJSON(&newAlbum); err != nil {
		return
	}

	// Add the new album to the slice.
	albums = append(albums, newAlbum)
	c.IndentedJSON(http.StatusCreated, newAlbum)
}

// getAlbumByID locates the album whose ID value matches the id
// parameter sent by the client, then returns that album as a response.
func getAlbumByID(c *gin.Context) {
	id := c.Param("id")

	// Loop over the list of albums, looking for
	// an album whose ID value matches the parameter.
	for _, a := range albums {
		if a.ID == id {
			c.IndentedJSON(http.StatusOK, a)
			return
		}
	}
	c.IndentedJSON(http.StatusNotFound, gin.H{"message": "album not found"})
}

// go.mod: module example.com/albums
// go 1.22

// package main

// import (
// 	"hash/fnv"
// 	"net/http"
// 	"sort"
// 	"sync"

// 	"github.com/gin-gonic/gin"
// )

// type album struct {
// 	ID     string  `json:"id"`
// 	Title  string  `json:"title"`
// 	Artist string  `json:"artist"`
// 	Price  float64 `json:"price"`
// }

// const shardsN = 16

// type shard struct {
// 	mu sync.RWMutex
// 	m  map[string]album
// }

// var table [shardsN]shard

// func init() {
// 	for i := 0; i < shardsN; i++ {
// 		table[i].m = make(map[string]album)
// 	}
// 	seed := []album{
// 		{ID: "1", Title: "Blue Train", Artist: "John Coltrane", Price: 56.99},
// 		{ID: "2", Title: "Jeru", Artist: "Gerry Mulligan", Price: 17.99},
// 		{ID: "3", Title: "Sarah Vaughan and Clifford Brown", Artist: "Sarah Vaughan", Price: 39.99},
// 	}
// 	for _, a := range seed {
// 		put(a)
// 	}
// }

// func hashID(id string) uint32 {
// 	h := fnv.New32a()
// 	_, _ = h.Write([]byte(id))
// 	return h.Sum32()
// }
// func pick(id string) *shard { return &table[hashID(id)%shardsN] }

// func get(id string) (album, bool) {
// 	s := pick(id)
// 	s.mu.RLock()
// 	defer s.mu.RUnlock()
// 	a, ok := s.m[id]
// 	return a, ok
// }
// func put(a album) {
// 	s := pick(a.ID)
// 	s.mu.Lock()
// 	s.m[a.ID] = a
// 	s.mu.Unlock()
// }
// func listAll() []album {
// 	// Read-lock all shards to build a consistent snapshot
// 	for i := 0; i < shardsN; i++ {
// 		table[i].mu.RLock()
// 	}
// 	defer func() {
// 		for i := 0; i < shardsN; i++ {
// 			table[i].mu.RUnlock()
// 		}
// 	}()
// 	var out []album
// 	for i := 0; i < shardsN; i++ {
// 		for _, a := range table[i].m {
// 			out = append(out, a)
// 		}
// 	}
// 	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
// 	return out
// }

// func main() {
// 	gin.SetMode(gin.ReleaseMode)
// 	r := gin.Default()
// 	r.GET("/albums", func(c *gin.Context) { c.IndentedJSON(http.StatusOK, listAll()) })
// 	r.GET("/albums/:id", func(c *gin.Context) {
// 		if a, ok := get(c.Param("id")); ok {
// 			c.IndentedJSON(http.StatusOK, a)
// 			return
// 		}
// 		c.JSON(http.StatusNotFound, gin.H{"message": "album not found"})
// 	})
// 	r.POST("/albums", func(c *gin.Context) {
// 		var a album
// 		if err := c.ShouldBindJSON(&a); err != nil || a.ID == "" {
// 			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid input"})
// 			return
// 		}
// 		put(a)
// 		c.Header("Location", "/albums/"+a.ID)
// 		c.IndentedJSON(http.StatusCreated, a)
// 	})
// 	_ = r.Run(":8080")
// }
