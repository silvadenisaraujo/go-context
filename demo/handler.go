package demo

import (
	"context"
	"fmt"
	"math/rand"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// Global map to store cancellation functions
var activeRequests = sync.Map{}

// CancelHandler demonstrates manual cancellation of in-progress requests.
// It looks up a request by ID and cancels its context if found.
// Example: /cancel/req-123456
func CancelHandler(c *gin.Context) {
	requestID := c.Param("requestID")

	// Get and call the cancel function
	if cancelFunc, ok := activeRequests.Load(requestID); ok {
		cancelFunc.(context.CancelFunc)()
		c.JSON(http.StatusOK, gin.H{
			"message":   "Request canceled",
			"requestID": requestID,
		})
		return
	}

	c.JSON(http.StatusNotFound, gin.H{
		"error":     "Request not found",
		"requestID": requestID,
	})
}

// ProcessHandler demonstrates context timeout by processing a request with a configurable
// timeout. If processing takes longer than the specified timeout, the operation is canceled
// and an error is returned. The timeout is specified as a query parameter.
// Example: /process?timeout=5s
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

	// Create a new context with the timeout
	ctx, cancel := context.WithTimeout(c.Request.Context(), duration)
	defer cancel()

	// Simulate work in a goroutine
	resultChan := make(chan string)

	go func() {
		// Fixed sleep time of 5 seconds
		sleepTime := 5 * time.Second

		select {
		case <-ctx.Done():
			// No need to send anything, the main goroutine will handle this
			return
		case <-time.After(sleepTime):
			// Work completed successfully
			resultChan <- sleepTime.String()
			return
		}
	}()

	// Wait for result or timeout
	select {
	case sleepTime := <-resultChan:
		c.JSON(http.StatusOK, gin.H{
			"status":    "OK",
			"requestID": requestID,
			"timestamp": timestamp,
			"message":   "Process completed successfully",
			"sleepTime": sleepTime,
		})
	case <-ctx.Done():
		if ctx.Err() == context.DeadlineExceeded {
			c.JSON(http.StatusRequestTimeout, gin.H{
				"error":     "timeout",
				"requestID": requestID,
				"timestamp": timestamp,
				"message":   "Process timed out",
			})
		} else {
			c.JSON(http.StatusGatewayTimeout, gin.H{
				"error":     "canceled",
				"requestID": requestID,
				"timestamp": timestamp,
				"message":   "Request was canceled",
			})
		}
	}
}

// DemoParentCancellation demonstrates how cancellation propagates from parent to child contexts.
// It creates a parent context that's canceled after a random delay, and a child context
// with a longer timeout. When the parent is canceled, the child is also canceled.
func DemoParentCancellation(c *gin.Context) {
	requestID, _ := c.Get("requestID")
	timestamp, _ := c.Get("timestamp")

	// Create parent context with cancellation
	parentCtx, parentCancel := context.WithCancel(context.Background())

	// Simulate canceling the parent after some time (1-3 seconds)
	triggerTime := time.Duration(1+rand.Intn(3)) * time.Second
	go func() {
		time.Sleep(triggerTime)
		fmt.Printf("Request %v: Parent context canceled after %v\n", requestID, triggerTime)
		parentCancel()
	}()

	// Child context inherits from parent with a longer timeout
	childCtx, childCancel := context.WithTimeout(parentCtx, 10*time.Second)
	defer childCancel()

	// Simulate work in a goroutine
	resultChan := make(chan string)
	errChan := make(chan error)

	go func() {
		// Longer sleep time to ensure parent cancellation happens first
		sleepTime := 5 * time.Second

		select {
		case <-childCtx.Done():
			errChan <- childCtx.Err()
			return
		case <-time.After(sleepTime):
			resultChan <- sleepTime.String()
			return
		}
	}()

	// Wait for result or cancellation
	select {
	case sleepTime := <-resultChan:
		c.JSON(http.StatusOK, gin.H{
			"status":      "OK",
			"requestID":   requestID,
			"timestamp":   timestamp,
			"message":     "Process completed successfully",
			"sleepTime":   sleepTime,
			"parentDelay": triggerTime.String(),
		})
	case err := <-errChan:
		message := "Process failed"
		if err == context.Canceled {
			message = "Process was canceled by parent context"
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":       err.Error(),
			"requestID":   requestID,
			"timestamp":   timestamp,
			"message":     message,
			"parentDelay": triggerTime.String(),
		})
	}
}

// NeverRespondWithTimeoutContextHandler demonstrates a long-running operation that never responds.
// The handler creates a new context with a timeout and calls a function that locks the response.
func NeverRespondWithTimeoutContextHandler(c *gin.Context) {
	requestID, _ := c.Get("requestID")
	timestamp, _ := c.Get("timestamp")

	ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
	defer cancel()

	LockResponse(ctx)

	c.JSON(http.StatusOK, gin.H{
		"status":    "OK",
		"requestID": requestID,
		"timestamp": timestamp,
		"message":   "Process completed successfully",
	})
}

// NeverRespondWithGoContextHandler demonstrates a long-running operation that never responds.
// It uses a Go context to simulate a locked operation that never returns.
// This is a common mistake that can lead to goroutine leaks and resource exhaustion.
func NeverRespondWithGoContextHandler(c *gin.Context) {
	requestID, _ := c.Get("requestID")
	timestamp, _ := c.Get("timestamp")

	LockResponse(context.Background())

	c.JSON(http.StatusOK, gin.H{
		"status":    "OK",
		"requestID": requestID,
		"timestamp": timestamp,
		"message":   "Process completed successfully",
	})
}

// FireAndForgetHandler demonstrates a fire-and-forget operation that runs in the background.
// It starts a long-running operation in a goroutine and returns immediately without waiting for it.
// This is useful for tasks that don't need to block the main request flow.
func FireAndForgetHandler(c *gin.Context) {
	requestID, _ := c.Get("requestID")
	timestamp, _ := c.Get("timestamp")

	ctx := context.WithValue(context.Background(), "requestID", requestID)
	SafeFireAndForget(ctx)

	c.JSON(http.StatusOK, gin.H{
		"status":    "OK",
		"requestID": requestID,
		"timestamp": timestamp,
		"message":   "Process completed successfully",
	})
}

func BrokenFireAndForgetHandler(c *gin.Context) {
	requestID, _ := c.Get("requestID")
	timestamp, _ := c.Get("timestamp")

	// Create a context based on the gin context and we close it end of the function
	ctx, cancel := context.WithCancel(c.Request.Context())
	ctx = context.WithValue(ctx, "requestID", requestID)
	defer cancel()

	SafeFireAndForget(ctx)

	c.JSON(http.StatusOK, gin.H{
		"status":    "OK",
		"requestID": requestID,
		"timestamp": timestamp,
		"message":   "Process completed successfully",
	})
}

// SetupRoutes configures the Gin router with all the demonstration endpoints and middleware.
// It adds middleware for request ID generation and client cancellation detection,
// and registers all the handler functions.
func SetupRoutes(r *gin.Engine) {
	// Middleware to generate a request ID if not present
	r.Use(func(c *gin.Context) {
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = fmt.Sprintf("req-%d", rand.Int63())
		}
		c.Set("requestID", requestID)
		c.Set("timestamp", time.Now().Format(time.RFC3339))
		c.Next()
	})

	// Set up the routes
	r.GET("/process", ProcessHandler)
	r.GET("/cancel/:requestID", CancelHandler)
	r.GET("/demo-parent-cancel", DemoParentCancellation)
	r.GET("/never-respond", NeverRespondWithGoContextHandler)
	r.GET("/never-respond-timeout", NeverRespondWithTimeoutContextHandler)
	r.GET("/fire-and-forget", FireAndForgetHandler)
	r.GET("/broken-fire-and-forget", BrokenFireAndForgetHandler)
}
