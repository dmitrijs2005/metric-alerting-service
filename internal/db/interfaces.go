package db

import "context"

type DBClient interface {
	Ping(ctx context.Context) error
}
