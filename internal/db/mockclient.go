package db

import (
	"context"

	"github.com/stretchr/testify/mock"
)

type MockDBClient struct {
	mock.Mock
}

func (m *MockDBClient) Ping(ctx context.Context) error {
	return nil
}
