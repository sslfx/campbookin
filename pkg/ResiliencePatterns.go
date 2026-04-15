package pkg

import (
	"context"
	"fmt"
	"math"
	"time"
)

func Retry(operation func() error, maxRetries int, baseDelay time.Duration) error {
	delay := baseDelay * time.Millisecond
	for i := 0; i < maxRetries; i++ {
		err := operation()
		if err == nil {
			return nil
		}
		if i == maxRetries-1 {
			return err
		}
		time.Sleep(delay)
		delay *= time.Duration(math.Pow(2, float64(i))) * time.Millisecond
	}
	return nil
}

func Timeout(operation func() error, timeout time.Duration) error {
	timeout = timeout * time.Millisecond
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	done := make(chan error, 1)
	go func() {
		done <- operation()
	}()
	select {
	case err := <-done:
		return err
	case <-ctx.Done():
		return fmt.Errorf("timed out")
	}
}

func ProcessWithDLQ(messages []string, operation func(string) error) error {
	dlq := make([]string, len(messages))
	for _, message := range messages {
		err := operation(message)
		if err != nil {
			fmt.Errorf("add to dlq: %v", err)
			dlq = append(dlq, message)
		}
	}
	return nil
}
