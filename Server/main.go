package main

import (
	"fmt"

	controllers "Magic/controllers"

	"github.com/gin-gonic/gin"
)

func main() {
	router := gin.Default()

	// Route for movies
	router.GET("/Movies", controllers.GetMovies())
	router.GET("/Movie/:imdb_id", controllers.GetMovie())
	router.POST("/Add", controllers.Addmovies())
	router.POST("/registeruser", controllers.RegisterUser())

	// Test route
	router.GET("/hello", func(c *gin.Context) {
		c.String(200, "done")
	})

	// Start server on port 8000
	if err := router.Run(":8000"); err != nil {
		fmt.Println("could not start the server:", err)
	}
}
