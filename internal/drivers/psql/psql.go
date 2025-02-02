// Модуль отвечает за взаимодействие с базой данных postgres
package psql

import (
	"database/sql"
	"errors"
	"fmt"

	_ "github.com/jackc/pgx/v5/stdlib"
)

// Названия полей и типы полей в базе данных
const (
	TABLENAME             = "metrics"
	COLUMNMETRIC          = "metric"
	COLUMNMETRICTYPE      = "TEXT"
	COLUMNMETRICVALUE     = "value"
	COLUMNMETRICVALUETYPE = "DOUBLE PRECISION"
)

// DB хранит в себе подключение к базе данных
type DB struct {
	*sql.DB
}

// ConnectDB подключение к базе данных
func ConnectDB(dsn string) (*DB, error) {
	db, err := openRetry("pgx", dsn)
	if err != nil {
		return nil, err
	}
	err = db.Ping()
	if err != nil {
		return nil, err
	}
	return &DB{
		DB: db,
	}, nil
}

// PingDB пинг базы данных
func (db *DB) PingDB() error {
	return db.DB.Ping()
}

// Close закрытие соединения с базой данных
func (db *DB) Close() error {
	return db.DB.Close()
}

// CreateMetricsTable создание таблицы для хранения метрик
func (db *DB) CreateMetricsTable() error {
	if db == nil {
		return ErrNotInit
	}
	if db.DB == nil {
		return ErrNotInit
	}
	query := fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s (
							%s %s,
							%s %s
						);`,
		TABLENAME,
		COLUMNMETRIC, COLUMNMETRICTYPE,
		COLUMNMETRICVALUE, COLUMNMETRICVALUETYPE)

	_, err := db.execRetry(query)
	return err
}

// PushReplace апдейт данных по метрикам
func (db *DB) PushReplace(metric, metricName string, value float64) error {
	ms := MetricString{
		MetricType: metric,
		MetricName: metricName,
	}
	query := `UPDATE ` + TABLENAME + ` ` +
		`SET ` + COLUMNMETRICVALUE + ` = $1 ` +
		`WHERE ` + COLUMNMETRIC + ` = $2;`

	// pgx НЕ ПОДДЕРЖИВАЕТ Value()
	result, err := db.execRetry(query, value, ms.MetricType+METRICSEPARATOR+ms.MetricName)
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		insertQuery := `INSERT INTO ` + TABLENAME + ` (` + COLUMNMETRIC + `, ` + COLUMNMETRICVALUE + `) VALUES ($1, $2);`
		_, err = db.execRetry(insertQuery, ms.MetricType+METRICSEPARATOR+ms.MetricName, value)
	}
	return err
}

// PushAdd добавление данных о метриках
func (db *DB) PushAdd(metric, metricName string, value float64) error {
	ms := MetricString{
		MetricType: metric,
		MetricName: metricName,
	}
	query := `INSERT INTO ` + TABLENAME +
		`(` + COLUMNMETRIC + `, ` + COLUMNMETRICVALUE + `) ` +
		`VALUES ($1, $2);`

	// pgx НЕ ПОДДЕРЖИВАЕТ Value()
	_, err := db.execRetry(query, ms.MetricType+METRICSEPARATOR+ms.MetricName, value)
	return err
}

// GetOneValue получение значения одной метрики
func (db *DB) GetOneValue(metric, metricName string) (float64, error) {
	ms := MetricString{
		MetricType: metric,
		MetricName: metricName,
	}
	query := `SELECT ` + COLUMNMETRICVALUE + ` ` +
		`FROM ` + TABLENAME + ` ` +
		`WHERE ` + COLUMNMETRIC + `=$1;`

	// pgx НЕ ПОДДЕРЖИВАЕТ Value()
	row := db.QueryRow(query, ms.MetricType+METRICSEPARATOR+ms.MetricName)
	var value sql.NullFloat64

	err := db.rowScanRetry(row, &value)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, ErrNoData
	} else if err != nil {
		return 0, err
	}
	if !value.Valid {
		return 0, ErrNoData
	}

	return value.Float64, nil
}

// GetArrayValues получения множества значений одной метрики
func (db *DB) GetArrayValues(metric, metricName string) (values []float64, err error) {
	ms := MetricString{
		MetricType: metric,
		MetricName: metricName,
	}
	query := `SELECT ` + COLUMNMETRICVALUE + ` ` +
		`FROM ` + TABLENAME + ` ` +
		`WHERE ` + COLUMNMETRIC + `=$1;`

	// pgx НЕ ПОДДЕРЖИВАЕТ Value()
	rows, err := db.queryRetry(query, ms.MetricType+METRICSEPARATOR+ms.MetricName)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNoData
	} else if err != nil {
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
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNoData
	} else if err != nil {
		return nil, err
	}

	return values, nil
}

// List получение всех данных по метрикам
func (db *DB) List(metricOneValue, metricArrayValues string) (map[string]float64, map[string][]float64, error) {
	typeValue := make(map[string]float64)
	typeValues := make(map[string][]float64)

	query := `SELECT * FROM ` + TABLENAME + `;`

	rows, err := db.queryRetry(query)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil, ErrNoData
	} else if err != nil {
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
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil, ErrNoData
	} else if err != nil {
		return nil, nil, err
	}

	return typeValue, typeValues, nil
}
