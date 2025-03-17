package process

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// ProcessHandler handles the /process endpoint with timeout management
func ProcessHandler(c *gin.Context) {
	// Get context values set by middleware
	requestID, _ := c.Get("requestID")
	timestamp, _ := c.Get("timestamp")

	// Parse timeout from query parameter, default to 10s if not provided
	timeoutStr := c.DefaultQuery("timeout", "10s")
	duration, err := time.ParseDuration(timeoutStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":     "Invalid timeout format",
			"requestID": requestID,
			"timestamp": timestamp,
			"message":   "Timeout should be in the format of 1s, 500ms, etc.",
		})
		return
	}

	// Create a context with the specified timeout
	ctx, cancel := context.WithTimeout(c.Request.Context(), duration)
	defer cancel()

	// Channel to signal completion of our "work"
	done := make(chan bool)

	// Simulate a long-running task
	go func() {
		// Simulate work that takes 5 seconds
		time.Sleep(5 * time.Second)
		done <- true
	}()

	// Wait for either the work to complete or the context to timeout
	select {
	case <-done:
		// Work completed successfully
		c.JSON(http.StatusOK, gin.H{
			"message":   "Processing completed successfully",
			"requestID": requestID,
			"timestamp": timestamp,
			"duration":  timeoutStr,
			"processed": true,
		})
	case <-ctx.Done():
		// Context cancelled or timed out
		err := ctx.Err()
		if err == context.DeadlineExceeded {
			c.JSON(http.StatusGatewayTimeout, gin.H{
				"error":     "Request timed out",
				"requestID": requestID,
				"timestamp": timestamp,
				"duration":  timeoutStr,
				"message":   "The operation took longer than the specified timeout",
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":     "Request cancelled",
				"requestID": requestID,
				"timestamp": timestamp,
				"message":   err.Error(),
			})
		}
	}
}
