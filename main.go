package main

import (
	"net/http"

	"go-context/demo"

	"github.com/gin-gonic/gin"
)

func main() {
	// Force colored logging
	gin.ForceConsoleColor()

	// Create a default gin router
	r := gin.Default()

	// Apply the request info middleware to all routes
	r.Use(demo.RequestInfoMiddleware)

	// Add the client cancellation middleware
	r.Use(demo.ClientCancellationMiddleware())

	// Define a route that responds to a GET request at /ping
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "Hello, World!",
		})
	})

	// Setup demo endpoints
	demo.SetupRoutes(r)

	// Run the server on port 8080
	r.Run(":8080")
}
