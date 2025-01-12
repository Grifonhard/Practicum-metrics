package psql

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"time"
)

// StorDB интерфейс для предоставления возможности вышестоящим сервисам мокировать DB
type StorDB interface {
	Begin() (*sql.Tx, error)
	BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error)
	Close() error
	Conn(ctx context.Context) (*sql.Conn, error)
	CreateMetricsTable() error
	Driver() driver.Driver
	Exec(query string, args ...any) (sql.Result, error)
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	GetArrayValues(metric string, metricName string) (values []float64, err error)
	GetOneValue(metric string, metricName string) (float64, error)
	List(metricOneValue string, metricArrayValues string) (map[string]float64, map[string][]float64, error)
	Ping() error
	PingContext(ctx context.Context) error
	PingDB() error
	Prepare(query string) (*sql.Stmt, error)
	PrepareContext(ctx context.Context, query string) (*sql.Stmt, error)
	PushAdd(metric string, metricName string, value float64) error
	PushReplace(metric string, metricName string, value float64) error
	Query(query string, args ...any) (*sql.Rows, error)
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRow(query string, args ...any) *sql.Row
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
	SetConnMaxIdleTime(d time.Duration)
	SetConnMaxLifetime(d time.Duration)
	SetMaxIdleConns(n int)
	SetMaxOpenConns(n int)
	Stats() sql.DBStats
}
