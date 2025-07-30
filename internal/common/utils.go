// Package common provides shared utilities, error types, logging helpers,
// cryptographic functions, and retry mechanisms used across the application.
package common

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"time"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
)

// tries to detect if there is a point to retry
func isErrorRetriable(err error) bool {

	if err == nil {
		return false
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

	// Проверка PostgreSQL спец. ошибок
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		// Class 08 - Connection Exception
		if pgerrcode.IsConnectionException(pgErr.Code) {
			return true
		}
	}

	// checking if error is network-related
	var netErr *net.OpError
	if errors.As(err, &netErr) && netErr.Temporary() {
		return true
	}

	return false

}

var (
	MaxRetries int           = 3
	ExpBackoff time.Duration = 2 * time.Second
)

// RetryWithResult retries the provided request function up to MaxRetries times with exponential backoff.
// It stops retrying if the error is not considered retriable or if the context is canceled.
// Returns the result of the request or the last encountered error.
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

func WriteToConsole(msg string) {
	fmt.Printf("%v %s \n", time.Now(), msg)
}
