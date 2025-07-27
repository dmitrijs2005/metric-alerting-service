package db

import (
	"context"
	"testing"

	"github.com/dmitrijs2005/metric-alerting-service/internal/metric"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
)

func TestPostgresRepository(t *testing.T) {

	ctx := context.Background()

	dbname := "test"
	dbuser := "test"
	dbpassword := "test"

	// Start the postgres ctr and run any migrations on it
	ctr, err := postgres.Run(
		ctx,
		"postgres:16-alpine",
		postgres.WithDatabase(dbname),
		postgres.WithUsername(dbuser),
		postgres.WithPassword(dbpassword),
		postgres.BasicWaitStrategies(),
		postgres.WithSQLDriver("pgx"),
	)
	require.NoError(t, err)

	dbURI, err := ctr.ConnectionString(ctx)
	require.NoError(t, err)

	t.Log("Postgres container started")

	client, err := NewPostgresClient(dbURI)
	require.NoError(t, err)

	err = client.Ping(ctx)
	require.NoError(t, err)

	t.Logf("Ping successfull")

	err = client.RunMigrations(ctx)
	require.NoError(t, err)

	t.Logf("Migrations successfull")

	RunRepositoryTests(t, ctx, client)

	defer testcontainers.CleanupContainer(t, ctr)

}

func RunRepositoryTests(t *testing.T, ctx context.Context, client *PostgresClient) {

	metrics := []metric.Metric{&metric.Counter{Name: "counter1", Value: 1}, &metric.Gauge{Name: "gauge1", Value: 3.14}}

	t.Run("AddMetrics", func(t *testing.T) {

		for _, m := range metrics {
			err := client.Add(ctx, m)
			assert.NoError(t, err)
		}

	})

	t.Run("Retrieve metrics", func(t *testing.T) {

		for _, m := range metrics {
			got, err := client.Retrieve(ctx, m.GetType(), m.GetName())

			assert.NoError(t, err, "Expected no error for existing metric")
			assert.Equal(t, m, got, "Retrieved metric should match the stored value")

		}

		_, err := client.Retrieve(ctx, metric.MetricTypeCounter, "some_nonexisting_metric")
		assert.Error(t, err, "Expected an error for non-existing metric")

	})

	t.Run("Retrieve all metrics", func(t *testing.T) {

		got, err := client.RetrieveAll(ctx)
		assert.NoError(t, err, "Expected no error for existing metric")
		assert.Equal(t, len(got), 2)

		for _, m := range got {
			found := false
			for _, mSource := range metrics {
				if m.GetName() == mSource.GetName() && m.GetType() == mSource.GetType() {
					assert.Equal(t, m.GetValue(), mSource.GetValue())
					found = true
				}
			}
			assert.True(t, found)
		}

	})

	t.Run("Update", func(t *testing.T) {

		type args struct {
			metric metric.Metric
			value  interface{}
		}
		tests := []struct {
			name      string
			args      args
			wantValue interface{}
		}{
			{"Test Counter update", args{&metric.Counter{Name: "counter1"}, int64(1)}, int64(2)},
			{"Test Gauge update", args{&metric.Gauge{Name: "gauge1"}, float64(4.15)}, float64(4.15)},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				err := client.Update(ctx, tt.args.metric, tt.args.value)
				assert.NoError(t, err)

				got, err := client.Retrieve(ctx, tt.args.metric.GetType(), tt.args.metric.GetName())
				assert.NoError(t, err, "Expected no error for existing metric")
				assert.Equal(t, tt.wantValue, got.GetValue(), "Retrieved metric should match the stored value")

			})
		}

	})

	t.Run("UpdateBatch", func(t *testing.T) {

		type upd struct {
			m metric.Metric
			v interface{}
		}

		updates := []upd{upd{m: &metric.Counter{Name: "counter1", Value: int64(2)}, v: int64(4)}, upd{m: &metric.Gauge{Name: "gauge1", Value: float64(4.15)}, v: float64(4.15)}}

		var batch []metric.Metric
		for _, item := range updates {
			batch = append(batch, item.m)
		}

		err := client.UpdateBatch(ctx, &batch)
		assert.NoError(t, err)

		items, err := client.RetrieveAll(ctx)
		assert.NoError(t, err)

		for _, u := range updates {
			found := false
			for _, item := range items {
				if u.m.GetName() == item.GetName() && u.m.GetType() == item.GetType() {
					assert.Equal(t, u.v, item.GetValue())
					found = true
				}
			}
			assert.True(t, found)
		}

	})

}
