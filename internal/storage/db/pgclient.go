package db

import (
	"context"
	"database/sql"
	"errors"

	"github.com/dmitrijs2005/metric-alerting-service/internal/common"
	"github.com/dmitrijs2005/metric-alerting-service/internal/metric"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
)

// PostgresClient provides a database-backed implementation of metric storage.
// It uses *sql.DB internally and supports transactional operations via DBExecutor.
type PostgresClient struct {
	db *sql.DB
}

// NewPostgresClient creates a new PostgresClient using the given DSN (Data Source Name).
// It opens a connection using the pgx driver and returns an initialized client.
func NewPostgresClient(dsn string) (*PostgresClient, error) {

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}

	return &PostgresClient{db}, nil
}

// Close closes the underlying database connection.
func (c *PostgresClient) Close() error {
	return c.db.Close()
}

// Ping checks whether the database connection is alive using PingContext.
func (c *PostgresClient) Ping(ctx context.Context) error {
	return c.db.PingContext(ctx)
}

// RunMigrations applies database schema migrations using goose.
func (c *PostgresClient) RunMigrations(ctx context.Context) error {
	goose.SetBaseFS(nil) // default is os.DirFS(".")

	if err := goose.UpContext(ctx, c.db, "./migrations"); err != nil {
		return err
	}

	return nil
}

// RetrieveAll fetches all stored metrics from the database and converts them
// into the metric.Metric interface representation.
func (c *PostgresClient) RetrieveAll(ctx context.Context) ([]metric.Metric, error) {

	var t metric.MetricType
	var n string
	var mvi sql.NullInt64
	var mvf sql.NullFloat64

	s := "select metric_type, metric_name, metric_value_int, metric_value_float from metrics"

	result := make([]metric.Metric, 0)

	rows, err := common.RetryWithResult(ctx, func() (*sql.Rows, error) {
		r, err := c.db.QueryContext(ctx, s)
		return r, err
	})

	if err != nil {
		return nil, err
	}

	for rows.Next() {

		err := rows.Scan(&t, &n, &mvi, &mvf)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return nil, common.ErrorMetricDoesNotExist
			} else {
				return nil, err
			}
		}

		m, err := metric.NewMetric(t, n)

		if err != nil {
			return nil, err
		}

		if gauge, ok := m.(*metric.Gauge); ok {
			err := gauge.Update(mvf.Float64)
			if err != nil {
				return nil, err
			}
		} else if counter, ok := m.(*metric.Counter); ok {
			err := counter.Update(mvi.Int64)
			if err != nil {
				return nil, err
			}
		} else {
			return nil, metric.ErrorInvalidMetricType
		}

		result = append(result, m)

	}

	// Check for iteration errors
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return result, nil

}

// ExecuteAdd inserts a metric using the provided DBExecutor (e.g. tx or db).
func (c *PostgresClient) ExecuteAdd(ctx context.Context, exec DBExecutor, m metric.Metric) error {
	var mvi sql.NullInt64
	var mvf sql.NullFloat64

	if gauge, ok := m.(*metric.Gauge); ok {
		mvi.Valid = false
		mvf.Float64 = gauge.Value
		mvf.Valid = true
	} else if counter, ok := m.(*metric.Counter); ok {
		mvi.Int64 = counter.Value
		mvi.Valid = true
		mvf.Valid = false
	} else {
		return metric.ErrorInvalidMetricType
	}

	s := "insert into metrics (metric_type, metric_name, metric_value_int, metric_value_float) values ($1, $2, $3, $4)"

	_, err := common.RetryWithResult(ctx, func() (sql.Result, error) {
		r, err := exec.ExecContext(ctx, s, m.GetType(), m.GetName(), mvi, mvf)
		return r, err
	})

	return err
}

// Add inserts a new metric into the database.
// If the metric already exists, this will result in a conflict.
func (c *PostgresClient) Add(ctx context.Context, m metric.Metric) error {
	return c.ExecuteAdd(ctx, c.db, m)
}

// ExecuteUpdate updates a metric using the provided DBExecutor.
func (c *PostgresClient) ExecuteUpdate(ctx context.Context, exec DBExecutor, m metric.Metric, v interface{}) error {

	s := "update metrics set "

	if _, ok := m.(*metric.Gauge); ok {
		s += "metric_value_float = $1 "
	} else if _, ok := m.(*metric.Counter); ok {
		s += "metric_value_int = metric_value_int + $1 "
	} else {
		return metric.ErrorInvalidMetricType
	}

	s += "where metric_type = $2 and metric_name = $3"

	_, err := common.RetryWithResult(ctx, func() (sql.Result, error) {
		r, err := exec.ExecContext(ctx, s, v, m.GetType(), m.GetName())
		return r, err
	})

	return err

}

// Update modifies the value of an existing metric.
// Counters are incremented; gauges are overwritten.
func (c *PostgresClient) Update(ctx context.Context, m metric.Metric, v interface{}) error {
	return c.ExecuteUpdate(ctx, c.db, m, v)
}

// ExecuteRetrieve fetches a single metric using the provided DBExecutor.
func (c *PostgresClient) ExecuteRetrieve(ctx context.Context, exec DBExecutor, t metric.MetricType, n string) (metric.Metric, error) {

	var mvi sql.NullInt64
	var mvf sql.NullFloat64

	s := "select metric_value_int, metric_value_float from metrics where metric_type=$1 and metric_name=$2"

	_, err := common.RetryWithResult(ctx, func() (*sql.Row, error) {
		r := exec.QueryRowContext(ctx, s, t, n)
		err := r.Scan(&mvi, &mvf)
		return r, err
	})

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, common.ErrorMetricDoesNotExist
		} else {
			return nil, err
		}
	}

	m, err := metric.NewMetric(t, n)
	if err != nil {
		return nil, err
	}

	if gauge, ok := m.(*metric.Gauge); ok {
		gauge.Value = mvf.Float64
	} else if counter, ok := m.(*metric.Counter); ok {
		counter.Value = mvi.Int64
	} else {
		return nil, metric.ErrorInvalidMetricType
	}

	return m, nil
}

// Retrieve fetches a single metric by type and name.
func (c *PostgresClient) Retrieve(ctx context.Context, t metric.MetricType, n string) (metric.Metric, error) {
	return c.ExecuteRetrieve(ctx, c.db, t, n)
}

// UpdateBatch updates a slice of metrics using a transaction.
// New metrics are inserted; existing ones are updated appropriately.
func (c *PostgresClient) UpdateBatch(ctx context.Context, metrics *[]metric.Metric) error {

	tx, err := c.db.Begin()
	if err != nil {
		return err
	}

	defer tx.Rollback()

	for _, metric := range *metrics {
		m, err := c.ExecuteRetrieve(ctx, tx, metric.GetType(), metric.GetName())

		if err != nil {
			if errors.Is(err, common.ErrorMetricDoesNotExist) {
				err := c.ExecuteAdd(ctx, tx, metric)
				if err != nil {
					return err
				}
			} else {
				return err
			}
		} else {
			err := c.ExecuteUpdate(ctx, tx, m, metric.GetValue())
			if err != nil {
				return err
			}
		}
	}

	return tx.Commit()

}
