package common

import (
	"context"
	"errors"
	"net"
	"os"
	"reflect"
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

func TestFilterArgs(t *testing.T) {
	tests := []struct {
		name         string
		args         []string
		allowedFlags []string
		want         []string
	}{
		{
			name:         "short flag with separate value",
			args:         []string{"-c", "conf.json", "-a", "localhost"},
			allowedFlags: []string{"-c", "--config"},
			want:         []string{"-c", "conf.json"},
		},
		{
			name:         "long flag with equals",
			args:         []string{"--config=alt.json", "-a", "localhost"},
			allowedFlags: []string{"-c", "--config"},
			want:         []string{"--config=alt.json"},
		},
		{
			name:         "both short and long present, preserve order",
			args:         []string{"--config=first.json", "-c", "second.json", "-x", "1"},
			allowedFlags: []string{"-c", "--config"},
			want:         []string{"--config=first.json", "-c", "second.json"},
		},
		{
			name:         "unknown flags ignored",
			args:         []string{"-x", "1", "--y=2", "positional"},
			allowedFlags: []string{"-c", "--config"},
			want:         []string{},
		},
		{
			name:         "flag without value at end is kept as-is",
			args:         []string{"-c"},
			allowedFlags: []string{"-c", "--config"},
			want:         []string{"-c"},
		},
		{
			name:         "flag followed by another flag (no value)",
			args:         []string{"-c", "-notvalue"},
			allowedFlags: []string{"-c", "--config"},
			want:         []string{"-c"},
		},
		{
			name:         "value that looks like a flag but with equals form",
			args:         []string{"--config=--weird.json"},
			allowedFlags: []string{"--config"},
			want:         []string{"--config=--weird.json"},
		},
		{
			name:         "multiple allowed flags kept",
			args:         []string{"-a", "localhost:8080", "-c", "conf.json", "--other", "x"},
			allowedFlags: []string{"-c", "-a"},
			want:         []string{"-a", "localhost:8080", "-c", "conf.json"},
		},
		{
			name:         "empty args",
			args:         []string{},
			allowedFlags: []string{"-c", "--config"},
			want:         []string{},
		},
		{
			name:         "path with spaces remains single arg",
			args:         []string{"-c", "/home/user/conf.json"},
			allowedFlags: []string{"-c"},
			want:         []string{"-c", "/home/user/conf.json"},
		},
		{
			name:         "do not treat next dash-starting token as value",
			args:         []string{"-c", "--config=alt.json"},
			allowedFlags: []string{"-c", "--config"},
			want:         []string{"-c", "--config=alt.json"},
		},
		{
			name:         "repeated allowed flag is preserved in order",
			args:         []string{"-c", "one.json", "-c", "two.json"},
			allowedFlags: []string{"-c"},
			want:         []string{"-c", "one.json", "-c", "two.json"},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			got := FilterArgs(tt.args, tt.allowedFlags)
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("FilterArgs() = %#v, want %#v", got, tt.want)
			}
		})
	}
}
