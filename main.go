package main

import (
	"net/http"
	"time"

	"go-context/process"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// RequestInfoMiddleware generates a request ID and timestamp and adds them to the context
func RequestInfoMiddleware(c *gin.Context) {
	// Generate unique request ID
	requestID := uuid.New().String()

	// Get current timestamp
	timestamp := time.Now()

	// Add values to the context
	c.Set("requestID", requestID)
	c.Set("timestamp", timestamp)

	// Continue processing
	c.Next()
}

func main() {
	// Force colored logging
	gin.ForceConsoleColor()

	// Create a default gin router
	r := gin.Default()

	// Apply the request info middleware to all routes
	r.Use(RequestInfoMiddleware)

	// Define a route that responds to a GET request at /ping
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "Hello, World!",
		})
	})

	// Add the new /process endpoint
	r.GET("/process", process.ProcessHandler)

	// Run the server on port 8080
	r.Run(":8080")
}
