// Package common provides shared utilities, error types, logging helpers,
// cryptographic functions, and retry mechanisms used across the application.
package common

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"strings"
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

// FilterArgs returns a slice of command-line arguments that only contains
// the allowed flags (and their values) specified in allowedFlags.
//
// Supported formats:
//  1. Flag and value as separate arguments:  -c conf.json
//  2. Flag and value combined with '=':      --config=conf.json
//
// Parameters:
//
//	args         — the command-line arguments (usually os.Args[1:])
//	allowedFlags — list of allowed flag names (e.g. []string{"-c", "--config"})
//
// Returns:
//
//	A slice containing the allowed flags and their values (if provided separately).
func FilterArgs(args []string, allowedFlags []string) []string {
	// Convert the list of allowed flags into a map for O(1) lookup
	allowed := make(map[string]struct{}, len(allowedFlags))
	for _, f := range allowedFlags {
		allowed[f] = struct{}{}
	}

	// Initialize the result slice as empty (not nil) so it’s always safe to use
	filtered := make([]string, 0, len(args))

	// Iterate over the arguments
	for i := 0; i < len(args); i++ {
		arg := args[i]

		// Case 1: flag in the form "--flag=value" or "-f=value"
		if strings.HasPrefix(arg, "-") && strings.Contains(arg, "=") {
			// Extract the flag name (before the '=')
			name := strings.SplitN(arg, "=", 2)[0]
			// If this flag is allowed, keep the whole "flag=value" argument
			if _, ok := allowed[name]; ok {
				filtered = append(filtered, arg)
			}
			continue
		}

		// Case 2: flag as a separate argument (value might follow)
		if _, ok := allowed[arg]; ok {
			filtered = append(filtered, arg)
			// If the next argument exists and does not look like another flag,
			// treat it as this flag's value and include it
			if i+1 < len(args) && !strings.HasPrefix(args[i+1], "-") {
				filtered = append(filtered, args[i+1])
				i++ // skip the value in the next loop iteration
			}
		}
	}

	return filtered
}
