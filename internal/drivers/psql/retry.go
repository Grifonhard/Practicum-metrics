package psql

import (
	"database/sql"
	"errors"
	"time"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
)

// Настройки повторных обращений в случае неудачных попыток
const (
	MAXRETRIES            = 3               // Максимальное количество попыток
	RETRYINTERVALINCREASE = 2 * time.Second // на столько растёт интервал между попытками, начиная с 1 секунды
)

// openRetry повторение попыток открытия данных
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

// execRetry повторение попыток отправить директивы в базу 
func (db *DB) execRetry(query string, args ...any) (result sql.Result, err error) {
	if db == nil {
		return nil, ErrNotInit
	}
	if db.DB == nil {
		return nil, ErrNotInit
	}
	for i := 0; i < MAXRETRIES; i++ {
		result, err = db.Exec(query, args...)
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

// rowScanRetry повторение попыток сканировать строку данных, пришедших из базы
func (db *DB) rowScanRetry(row *sql.Row, dest ...any) (err error) {
	if db == nil {
		return ErrNotInit
	}
	if db.DB == nil {
		return ErrNotInit
	}
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

// queryRetry повторение попыток сделать запрос
func (db *DB) queryRetry(query string, args ...any) (rows *sql.Rows, err error) {
	if db == nil {
		return nil, ErrNotInit
	}
	if db.DB == nil {
		return nil, ErrNotInit
	}
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
