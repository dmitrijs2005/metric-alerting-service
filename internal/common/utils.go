package common

import (
	"errors"
	"net"
	"os"
	"syscall"
)

func isConnectionRefused(err error) bool {
	var opErr *net.OpError
	if errors.As(err, &opErr) { // Unwraps and checks if it's a network operation error
		var sysErr *os.SyscallError
		if errors.As(opErr.Err, &sysErr) { // Unwraps and checks for syscall error
			return sysErr.Err == syscall.ECONNREFUSED
		}
	}
	return false
}

func IsErrorRetriable(err error) bool {
	return isConnectionRefused(err)
}
