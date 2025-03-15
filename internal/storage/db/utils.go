package db

import (
	"context"
	"errors"
	"net"
	"os"
	"time"
)

var (
	MaxRetries           = 3
	ExpBackoffMultiplier = 2 * time.Second
)

// tries to detect if there is a point to retry
func isErrorRetriable(err error) bool {

	if err == nil {
		return false
	}

	// checking if error is network-related
	var netErr *net.OpError
	if errors.As(err, &netErr) {
		return true
	}

	// checking if syscall error (connect refused)
	var sysErr *os.SyscallError
	if errors.As(err, &sysErr) {
		return true
	}

	// connection timeout
	if errors.Is(err, os.ErrDeadlineExceeded) {
		return true
	}

	return false

}

func calculateBackoff(i int) time.Duration {
	return 1*time.Second + time.Duration(i)*ExpBackoffMultiplier
}

// function that makes retries
func retryWithResult[T any](ctx context.Context, request func() (T, error)) (T, error) {
	var result T
	var err error

	for i := 0; i < 1+MaxRetries; i++ {

		//fmt.Println(time.Now(), "try", i)
		result, err = request()

		if err == nil {
			return result, nil
		}

		if !isErrorRetriable(err) {
			return result, err
		}

		backoff := calculateBackoff(i)

		select {
		case <-time.After(backoff):
		case <-ctx.Done():
			return result, ctx.Err() // Context timeout/cancellation
		}

	}

	return result, err
}
