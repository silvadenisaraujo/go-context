package demo

import (
	"context"
	"fmt"
	"math/rand"
	"time"
)

// LockResponse demonstrates a long-running operation that has a locked method on defer
func LockResponse(ctx context.Context) {
	defer func() {
		onlyReturnWhenContextCancelled(ctx)
	}()

	fmt.Println("This line will be printed")
}

func SafeFireAndForget(ctx context.Context) {
	// Fire a long running operation in a goroutine
	go longRunningOperation(ctx)
}

// LongRunningOperation simulates a long-running operation that takes 2 minutes to complete.
func longRunningOperation(ctx context.Context) {
	fmt.Println("Starting long-running operation for request_id: ", ctx.Value("requestID"))

	// This takes can randomly fail
	if rand.Intn(10) < 2 {
		fmt.Println("Operation failed for request_id: ", ctx.Value("requestID"))
		return
	}

	select {
	case <-ctx.Done():
		fmt.Println("Operation canceled for request_id: ", ctx.Value("requestID"))
		return
	case <-time.After(2 * time.Minute):
		fmt.Println("Operation completed for request_id: ", ctx.Value("requestID"))
		return
	}
}

// Only return when the context is Done, to simulate a long-running operation that can be canceled.
func onlyReturnWhenContextCancelled(ctx context.Context) {
	select {
	case <-ctx.Done():
		return
	}
}
