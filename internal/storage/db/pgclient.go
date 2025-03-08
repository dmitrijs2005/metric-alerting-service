package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/dmitrijs2005/metric-alerting-service/internal/metric"
	"github.com/dmitrijs2005/metric-alerting-service/internal/storage"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
)

type PostgresClient struct {
	db *sql.DB
}

func NewPostgresClient(dsn string) (*PostgresClient, error) {

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}

	return &PostgresClient{db}, nil
}

func (c *PostgresClient) Close() error {
	return c.db.Close()
}

func (c *PostgresClient) Ping(ctx context.Context) error {
	return c.db.PingContext(ctx)
}

func (c *PostgresClient) RunMigrations(ctx context.Context) error {
	goose.SetBaseFS(nil) // default is os.DirFS(".")

	if err := goose.UpContext(ctx, c.db, "./migrations"); err != nil {
		return err
	}

	return nil
}

func (c *PostgresClient) RetrieveAll(ctx context.Context) ([]metric.Metric, error) {

	var t metric.MetricType
	var n string
	var mvi sql.NullInt64
	var mvf sql.NullFloat64

	s := "select metric_type, metric_name, metric_value_int, metric_value_float from metrics"

	result := make([]metric.Metric, 0)

	rows, err := c.db.QueryContext(ctx, s)
	if err != nil {
		return nil, err
	}

	for rows.Next() {

		err := rows.Scan(&t, &n, &mvi, &mvf)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return nil, errors.New(storage.MetricDoesNotExist)
			} else {
				return nil, err
			}
		}

		fmt.Println(t, n, mvi, mvf)

		m, err := metric.NewMetric(t, n)

		if err != nil {
			return nil, err
		}

		if gauge, ok := m.(*metric.Gauge); ok {
			fmt.Println("111")
			err := gauge.Update(mvf.Float64)
			if err != nil {
				return nil, err
			}
			fmt.Println("111", gauge)
		} else if counter, ok := m.(*metric.Counter); ok {
			fmt.Println("222")
			err := counter.Update(mvi.Int64)
			if err != nil {
				return nil, err
			}
			fmt.Println("222", counter)
		} else {
			return nil, metric.ErrorInvalidMetricType
		}

		result = append(result, m)

	}

	return result, nil

}

func (c *PostgresClient) Add(ctx context.Context, m metric.Metric) error {

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
	_, err := c.db.ExecContext(ctx, s, m.GetType(), m.GetName(), mvi, mvf)
	return err
}

func (c *PostgresClient) Update(ctx context.Context, m metric.Metric, v interface{}) error {

	var mvi sql.NullInt64
	var mvf sql.NullFloat64

	if gauge, ok := m.(*metric.Gauge); ok {
		mvi.Valid = false
		err := gauge.Update(v)
		if err != nil {
			return err
		}
		mvf.Float64 = gauge.Value
		mvf.Valid = true
	} else if counter, ok := m.(*metric.Counter); ok {
		err := counter.Update(v)
		if err != nil {
			return err
		}
		mvi.Int64 = counter.Value
		mvi.Valid = true
		mvf.Valid = false
	} else {
		return metric.ErrorInvalidMetricType
	}

	s := "update metrics set metric_value_int = $1, metric_value_float = $2 where metric_type = $3 and metric_name = $4"
	_, err := c.db.ExecContext(ctx, s, mvi, mvf, m.GetType(), m.GetName())
	return err

}

func (c *PostgresClient) Retrieve(ctx context.Context, t metric.MetricType, n string) (metric.Metric, error) {

	var mvi sql.NullInt64
	var mvf sql.NullFloat64

	s := "select metric_value_int, metric_value_float from metrics where metric_type=$1 and metric_name=$2"

	row := c.db.QueryRowContext(ctx, s, t, n)
	err := row.Scan(&mvi, &mvf)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New(storage.MetricDoesNotExist)
		} else {
			return nil, err
		}
	}

	m, err := metric.NewMetric(t, n)

	if gauge, ok := m.(*metric.Gauge); ok {
		gauge.Update(mvf)
	} else if counter, ok := m.(*metric.Counter); ok {
		counter.Update(mvi)
	} else {
		return nil, metric.ErrorInvalidMetricType
	}

	return m, nil
}
