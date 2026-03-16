package main

import (
	"Magic/routes"
	"fmt"
	"log"

	"github.com/joho/godotenv"

	"github.com/gin-gonic/gin"
)

func main() {
	router := gin.Default()
	err := godotenv.Load(".env")
	if err != nil {
		log.Println("warning could not load the .env file")
	}

	routes.SetPublicRoute(router)
	routes.SetProtectedRoute(router)

	// Test route
	router.GET("/hello", func(c *gin.Context) {
		c.String(200, "done")
	})

	// Start server on port 8000
	if err := router.Run(":8000"); err != nil {
		fmt.Println("could not start the server:", err)
	}
}
