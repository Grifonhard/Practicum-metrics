package psql

import (
	"database/sql"
	"time"
	"errors"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
)

// если неудачно
const (
	MAXRETRIES            = 3               // Максимальное количество попыток
	RETRYINTERVALINCREASE = 2 * time.Second // на столько растёт интервал между попытками, начиная с 1 секунды
)

func openRetry(driverName string, dataSourceName string) (db *sql.DB, err error) {
	for i := 0; i < MAXRETRIES; i++ {
		db, err = sql.Open(driverName, dataSourceName)
		if err != nil {
			if pgErr := new(pgconn.PgError); errors.As(err, &pgErr) && pgErr.Code == pgerrcode.ConnectionException {
				time.Sleep(time.Second + RETRYINTERVALINCREASE*time.Duration(i))
				continue
			}
			break
		} else {
			break
		}
	}
	return 
}

func (db *DB) execRetry(query string, args ...any) (result sql.Result, err error) {
	for i := 0; i < MAXRETRIES; i++ {
		result, err = db.Exec(query)
		if err != nil {
			if pgErr := new(pgconn.PgError); errors.As(err, &pgErr) && pgErr.Code == pgerrcode.ConnectionException {
				time.Sleep(time.Second + RETRYINTERVALINCREASE*time.Duration(i))
				continue
			}
			break
		} else {
			break
		}
	}
	return result, err
}

func (db *DB) rowScanRetry(row *sql.Row, dest ...any) (err error) {
	for i := 0; i < MAXRETRIES; i++ {
		err = row.Scan(dest...)
		if err != nil {
			if pgErr := new(pgconn.PgError); errors.As(err, &pgErr) && pgErr.Code == pgerrcode.ConnectionException {
				time.Sleep(time.Second + RETRYINTERVALINCREASE*time.Duration(i))
				continue
			}
			break
		} else {
			break
		}
	}
	return err
}

func (db *DB) queryRetry(query string, args ...any) (rows *sql.Rows, err error) {
	for i := 0; i < MAXRETRIES; i++ {
		rows, err = db.Query(query, args...)
		if err != nil {
			if pgErr := new(pgconn.PgError); errors.As(err, &pgErr) && pgErr.Code == pgerrcode.ConnectionException {
				time.Sleep(time.Second + RETRYINTERVALINCREASE*time.Duration(i))
				continue
			}
			break
		} else {
			break
		}
	}
	return rows, err
}