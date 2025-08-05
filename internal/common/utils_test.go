package common

import (
	"context"
	"errors"
	"net"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fake retriable syscall error
func makeSyscallError() error {
	return &os.SyscallError{
		Err: errors.New("connection refused"),
	}
}

func makeNetTemporaryError() error {
	return &net.OpError{
		Err: &temporaryError{true},
	}
}

type temporaryError struct {
	temporary bool
}

func (e *temporaryError) Error() string   { return "temp error" }
func (e *temporaryError) Temporary() bool { return e.temporary }
func (e *temporaryError) Timeout() bool   { return false }

func makePgConnExceptionError() error {
	return &pgconn.PgError{
		Code: "08006", // Connection failure
	}
}

func makeNonRetriableError() error {
	return errors.New("permanent failure")
}

func TestIsErrorRetriable(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{"nil error", nil, false},
		{"syscall error", makeSyscallError(), true},
		{"deadline exceeded", os.ErrDeadlineExceeded, true},
		{"pg error", makePgConnExceptionError(), true},
		{"temporary net error", makeNetTemporaryError(), true},
		{"non-retriable error", makeNonRetriableError(), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, isErrorRetriable(tt.err))
		})
	}
}

func TestRetryWithResult_SuccessOnFirstTry(t *testing.T) {
	ctx := context.Background()
	callCount := 0

	result, err := RetryWithResult(ctx, func() (string, error) {
		callCount++
		return "ok", nil
	})

	require.NoError(t, err)
	assert.Equal(t, "ok", result)
	assert.Equal(t, 1, callCount)
}

func TestRetryWithResult_EventualSuccess(t *testing.T) {
	ctx := context.Background()
	callCount := 0

	result, err := RetryWithResult(ctx, func() (string, error) {
		callCount++
		if callCount < 3 {
			return "", makePgConnExceptionError()
		}
		return "done", nil
	})

	require.NoError(t, err)
	assert.Equal(t, "done", result)
	assert.Equal(t, 3, callCount)
}

func TestRetryWithResult_NonRetriableError(t *testing.T) {
	ctx := context.Background()
	callCount := 0

	result, err := RetryWithResult(ctx, func() (string, error) {
		callCount++
		return "", makeNonRetriableError()
	})

	require.Error(t, err)
	assert.Equal(t, "", result)
	assert.Equal(t, 1, callCount)
}

func TestRetryWithResult_ContextCancelled(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	callCount := 0

	result, err := RetryWithResult(ctx, func() (string, error) {
		callCount++
		return "", makePgConnExceptionError()
	})

	require.ErrorIs(t, err, context.DeadlineExceeded)
	assert.Equal(t, "", result)
	assert.GreaterOrEqual(t, callCount, 1)
}
