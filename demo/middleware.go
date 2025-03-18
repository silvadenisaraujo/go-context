package demo

import (
	"context"
	"time"

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

// ClientCancellationMiddleware handles client disconnection by canceling the request context
// when the client closes the connection. This allows long-running operations to terminate
// early if the client is no longer waiting for a response.
func ClientCancellationMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Create a context that will be canceled when the client disconnects
		ctx, cancel := context.WithCancel(c.Request.Context())
		defer cancel()

		// Replace the request context
		c.Request = c.Request.WithContext(ctx)

		// Monitor for client disconnection
		go func() {
			select {
			case <-c.Request.Context().Done():
				cancel()
			}
		}()

		c.Next()
	}
}
