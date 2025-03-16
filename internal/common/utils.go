package common

import (
	"context"
	"errors"
	"net"
	"os"
	"time"
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

var (
	MaxRetries int           = 3
	ExpBackoff time.Duration = 2 * time.Second
)

// function that processes retries
func RetryWithResult[T any](ctx context.Context, request func() (T, error)) (T, error) {
	var result T
	var err error

	for i := 0; i < 1+MaxRetries; i++ {

		result, err = request()

		if err == nil {
			return result, nil
		}

		if !isErrorRetriable(err) {
			return result, err
		}

		backoff := 1*time.Second + time.Duration(i)*ExpBackoff

		select {
		case <-time.After(backoff):
		case <-ctx.Done():
			return result, ctx.Err() // Context timeout/cancellation
		}

	}

	return result, err
}
