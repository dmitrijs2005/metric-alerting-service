-- +goose Up
-- +goose StatementBegin
CREATE TABLE metrics (
    metric_name TEXT NOT NULL,
    metric_type TEXT NOT NULL,  -- e.g., "int64" or "float64"
    metric_value_int BIGINT,  -- For int64 metrics (NULL if float)
    metric_value_float DOUBLE PRECISION,  -- For float64 metrics (NULL if int)

    PRIMARY KEY (metric_name, metric_type)  -- Composite PK
);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE metrics
-- +goose StatementEnd
