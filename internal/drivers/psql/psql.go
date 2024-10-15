package psql

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	_ "github.com/jackc/pgx/v5/stdlib"
)

const (
	TABLENAME             = "metrics"
	COLUMNMETRIC          = "metric"
	COLUMNMETRICTYPE      = "TEXT"
	COLUMNMETRICVALUE     = "value"
	COLUMNMETRICVALUETYPE = "DOUBLE PRECISION"
)

// если неудачное
const (
	MAXRETRIES            = 3               // Максимальное количество попыток
	RETRYINTERVALINCREASE = 2 * time.Second // на столько растёт интервал между попытками, начиная с 1 секунды
)

type DB struct {
	*sql.DB
}

func ConnectDB(dsn string) (*DB, error) {
	var db *sql.DB
	var err error
	for i := 0; i < MAXRETRIES; i++ {
		db, err = sql.Open("pgx", dsn)
		if err != nil {
			if pgErr := (*pgconn.PgError)(nil); errors.As(err, pgErr) && pgErr.Code == pgerrcode.ConnectionException {
				time.Sleep(time.Second + RETRYINTERVALINCREASE*time.Duration(i))
				continue
			}
			break
		} else {
			break
		}
	}
	if err == nil {
		return &DB{
			DB: db,
		}, nil
	} else {
		return nil, err
	}
}

func (db *DB) PingDB() error {
	return db.DB.Ping()
}

func (db *DB) Close() error {
	return db.DB.Close()
}

func (db *DB) CreateMetricsTable() error {
	query := fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s (
							%s %s,
							%s %s
						);`,
		TABLENAME,
		COLUMNMETRIC, COLUMNMETRICTYPE,
		COLUMNMETRICVALUE, COLUMNMETRICVALUETYPE)

	var err error
	for i := 0; i < MAXRETRIES; i++ {
		_, err := db.Exec(query)
		if err != nil {
			if pgErr := (*pgconn.PgError)(nil); errors.As(err, pgErr) && pgErr.Code == pgerrcode.ConnectionException {
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

func (db *DB) PushReplace(metric, metricName string, value float64) error {
	ms := MetricString{
		MetricType: metric,
		MetricName: metricName,
	}
	query := `UPDATE ` + TABLENAME + ` ` +
		`SET ` + COLUMNMETRICVALUE + ` = $1 ` +
		`WHERE ` + COLUMNMETRIC + ` = $2;`

	// pgx НЕ ПОДДЕРЖИВАЕТ Value()
	var result sql.Result
	var err error
	for i := 0; i < MAXRETRIES; i++ {
		result, err = db.Exec(query, value, ms.MetricType+METRICSEPARATOR+ms.MetricName)
		if err != nil {
			if pgErr := (*pgconn.PgError)(nil); errors.As(err, pgErr) && pgErr.Code == pgerrcode.ConnectionException {
				time.Sleep(time.Second + RETRYINTERVALINCREASE*time.Duration(i))
				continue
			}
			break
		} else {
			break
		}
	}
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		insertQuery := `INSERT INTO ` + TABLENAME + ` (` + COLUMNMETRIC + `, ` + COLUMNMETRICVALUE + `) VALUES ($1, $2);`
		for i := 0; i < MAXRETRIES; i++ {
			_, err = db.Exec(insertQuery, ms.MetricType+METRICSEPARATOR+ms.MetricName, value)
			if err != nil {
				if pgErr := (*pgconn.PgError)(nil); errors.As(err, pgErr) && pgErr.Code == pgerrcode.ConnectionException {
					time.Sleep(time.Second + RETRYINTERVALINCREASE*time.Duration(i))
					continue
				}
				break
			} else {
				break
			}
		}
	}
	return err
}

func (db *DB) PushAdd(metric, metricName string, value float64) error {
	ms := MetricString{
		MetricType: metric,
		MetricName: metricName,
	}
	query := `INSERT INTO ` + TABLENAME +
		`(` + COLUMNMETRIC + `, ` + COLUMNMETRICVALUE + `) ` +
		`VALUES ($1, $2);`

	var err error
	// pgx НЕ ПОДДЕРЖИВАЕТ Value()
	for i := 0; i < MAXRETRIES; i++ {
		_, err = db.Exec(query, ms.MetricType+METRICSEPARATOR+ms.MetricName, value)
		if err != nil {
			if pgErr := (*pgconn.PgError)(nil); errors.As(err, pgErr) && pgErr.Code == pgerrcode.ConnectionException {
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

func (db *DB) GetOneValue(metric, metricName string) (float64, error) {
	ms := MetricString{
		MetricType: metric,
		MetricName: metricName,
	}
	query := `SELECT ` + COLUMNMETRICVALUE + ` ` +
		`FROM ` + TABLENAME + ` ` +
		`WHERE ` + COLUMNMETRIC + `=$1;`

	// pgx НЕ ПОДДЕРЖИВАЕТ Value()
	var err error
	row := db.QueryRow(query, ms.MetricType+METRICSEPARATOR+ms.MetricName)
	var value sql.NullFloat64

	for i := 0; i < MAXRETRIES; i++ {
		err = row.Scan(&value)
		if err != nil {
			if pgErr := (*pgconn.PgError)(nil); errors.As(err, pgErr) && pgErr.Code == pgerrcode.ConnectionException {
				time.Sleep(time.Second + RETRYINTERVALINCREASE*time.Duration(i))
				continue
			}
			break
		} else {
			break
		}
	}
	if err != nil {
		return 0, err
	}
	if !value.Valid {
		return 0, ErrNoData
	}

	return value.Float64, nil
}

func (db *DB) GetArrayValues(metric, metricName string) (values []float64, err error) {
	ms := MetricString{
		MetricType: metric,
		MetricName: metricName,
	}
	query := `SELECT ` + COLUMNMETRICVALUE + ` ` +
		`FROM ` + TABLENAME + ` ` +
		`WHERE ` + COLUMNMETRIC + `=$1;`

	// pgx НЕ ПОДДЕРЖИВАЕТ Value()
	var rows *sql.Rows
	for i := 0; i < MAXRETRIES; i++ {
		rows, err = db.Query(query, ms.MetricType+METRICSEPARATOR+ms.MetricName)
		if err != nil {
			if pgErr := (*pgconn.PgError)(nil); errors.As(err, pgErr) && pgErr.Code == pgerrcode.ConnectionException {
				time.Sleep(time.Second + RETRYINTERVALINCREASE*time.Duration(i))
				continue
			}
			break
		} else {
			break
		}
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var value float64

		err = rows.Scan(&value)
		if err != nil {
			return nil, err
		}

		values = append(values, value)
	}

	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return values, nil
}

func (db *DB) List(metricOneValue, metricArrayValues string) (map[string]float64, map[string][]float64, error) {
	typeValue := make(map[string]float64)
	typeValues := make(map[string][]float64)

	query := `SELECT * FROM ` + TABLENAME + `;`

	var err error
	var rows *sql.Rows
	for i := 0; i < MAXRETRIES; i++ {
		rows, err = db.Query(query)
		if err != nil {
			if pgErr := (*pgconn.PgError)(nil); errors.As(err, pgErr) && pgErr.Code == pgerrcode.ConnectionException {
				time.Sleep(time.Second + RETRYINTERVALINCREASE*time.Duration(i))
				continue
			}
			break
		} else {
			break
		}
	}
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var metric Metric

		err = rows.Scan(&metric.MetricS, &metric.Value)
		if err != nil {
			return nil, nil, err
		}

		switch metric.MetricS.MetricType {
		case metricOneValue:
			typeValue[metric.MetricS.MetricName] = metric.Value
		case metricArrayValues:
			typeValues[metric.MetricS.MetricName] = append(typeValues[metric.MetricS.MetricName], metric.Value)
		default:
			return nil, nil, ErrUnexpectedMetricType
		}
	}

	err = rows.Err()
	if err != nil {
		return nil, nil, err
	}

	return typeValue, typeValues, nil
}
