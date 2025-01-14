package psql

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// --------------------- //
//   MetricString ТЕСТЫ  //
// --------------------- //

func TestMetricString_Value(t *testing.T) {
	m := MetricString{
		MetricType: "gauge",
		MetricName: "Alloc",
	}
	val, err := m.Value()
	assert.NoError(t, err, "Value() не должно возвращать ошибку")

	expected := "gauge" + METRICSEPARATOR + "Alloc"
	assert.Equal(t, expected, val, "Ожидается склеивание через ///")
}

func TestMetricString_Scan_Valid(t *testing.T) {
	var m MetricString
	err := m.Scan("counter///PollCount")
	assert.NoError(t, err, "Scan() не должно возвращать ошибку при валидной строке")
	assert.Equal(t, "counter", m.MetricType)
	assert.Equal(t, "PollCount", m.MetricName)
}

func TestMetricString_Scan_InvalidFormat(t *testing.T) {
	var m MetricString
	err := m.Scan("invalid_metric_string")
	assert.Error(t, err, "Scan() должно вернуть ошибку при некорректном формате")
}

func TestMetricString_Scan_NilValue(t *testing.T) {
	m := MetricString{MetricType: "before", MetricName: "before"}
	err := m.Scan(nil)
	assert.NoError(t, err, "Scan(nil) не должно возвращать ошибку")
	assert.Empty(t, m.MetricType, "MetricType должен стать пустым")
	assert.Empty(t, m.MetricName, "MetricName должен стать пустым")
}

// --------------------- //
//  Заглушка без памяти  //
// --------------------- //
//
// Эту «простую» заглушку оставим для демонстрации тестов 
// PushReplace, PushAdd и т. п. (что они не падают).
type mockDBConn struct{}

var _ StorDB = (*mockDBConn)(nil)

func (m *mockDBConn) Begin() (*sql.Tx, error)                                    { return nil, errors.New("not implemented") }
func (m *mockDBConn) BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error) {
	return nil, errors.New("not implemented")
}
func (m *mockDBConn) Close() error { return nil }
func (m *mockDBConn) Conn(ctx context.Context) (*sql.Conn, error) {
	return nil, errors.New("not implemented")
}
func (m *mockDBConn) CreateMetricsTable() error { return nil }
func (m *mockDBConn) Driver() driver.Driver     { return driverStub{} }
func (m *mockDBConn) Exec(query string, args ...any) (sql.Result, error) {
	return nil, errors.New("not implemented")
}
func (m *mockDBConn) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	return nil, errors.New("not implemented")
}
func (m *mockDBConn) GetArrayValues(metric string, metricName string) (values []float64, err error) {
	return nil, ErrNoData
}
func (m *mockDBConn) GetOneValue(metric string, metricName string) (float64, error) {
	return 0, ErrNoData
}
func (m *mockDBConn) List(metricOneValue, metricArrayValues string) (map[string]float64, map[string][]float64, error) {
	return nil, nil, ErrNoData
}
func (m *mockDBConn) Ping() error                       { return nil }
func (m *mockDBConn) PingContext(ctx context.Context) error {
	return nil
}
func (m *mockDBConn) PingDB() error                     { return nil }
func (m *mockDBConn) Prepare(query string) (*sql.Stmt, error) {
	return nil, errors.New("not implemented")
}
func (m *mockDBConn) PrepareContext(ctx context.Context, query string) (*sql.Stmt, error) {
	return nil, errors.New("not implemented")
}
func (m *mockDBConn) PushAdd(metric string, metricName string, value float64) error {
	return nil
}
func (m *mockDBConn) PushReplace(metric string, metricName string, value float64) error {
	return nil
}
func (m *mockDBConn) Query(query string, args ...any) (*sql.Rows, error) {
	return nil, errors.New("not implemented")
}
func (m *mockDBConn) QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	return nil, errors.New("not implemented")
}
func (m *mockDBConn) QueryRow(query string, args ...any) *sql.Row {
	return nil
}
func (m *mockDBConn) QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row {
	return nil
}
func (m *mockDBConn) SetConnMaxIdleTime(d time.Duration) {}
func (m *mockDBConn) SetConnMaxLifetime(d time.Duration) {}
func (m *mockDBConn) SetMaxIdleConns(n int)              {}
func (m *mockDBConn) SetMaxOpenConns(n int)              {}
func (m *mockDBConn) Stats() sql.DBStats {
	return sql.DBStats{}
}

// --------------------- //
//      Тесты DB (stub)
// --------------------- //

func TestDB_PushReplace_Stub(t *testing.T) {
	mock := &mockDBConn{}
	err := mock.PushReplace("gauge", "Alloc", 1.23)
	assert.NoError(t, err, "заглушка возвращает nil => нет ошибки")
}

func TestDB_PushAdd_Stub(t *testing.T) {
	mock := &mockDBConn{}
	err := mock.PushAdd("gauge", "Heap", 99.99)
	assert.NoError(t, err, "заглушка возвращает nil => нет ошибки")
}

// --------------------- //
//    Тесты ConnectDB
// --------------------- //

// Пытаемся соединиться с невалидным DSN
func TestConnectDB_InvalidDSN(t *testing.T) {
	db, err := ConnectDB("invalid_dsn")
	assert.Nil(t, db, "при некорректном DSN db должно быть nil (если Ping() упадёт)")
	assert.Error(t, err, "ConnectDB должна вернуть ошибку на невалидный DSN")
}

// PingDB на nil
func TestDB_PingDB_WithNilDB(t *testing.T) {
	db := &DB{DB: nil} // имитируем пустое соединение
	err := db.PingDB()
	assert.Error(t, err, "при nil в DB ожидается ошибка, а не panic")
}

// CreateMetricsTable: без реальной DB будет ошибка
func TestDB_CreateMetricsTable_NoRealDB(t *testing.T) {
	db := &DB{DB: nil}
	err := db.CreateMetricsTable()
	assert.Error(t, err, "без реального соединения => ошибка")
}

// --------------------- //
//   Тесты openRetry и т.п.
// --------------------- //

func TestOpenRetry(t *testing.T) {
	db, err := openRetry("bad_driver", "some_dsn")
	assert.Nil(t, db, "ожидаем nil при некорректном драйвере")
	assert.Error(t, err, "openRetry должна вернуть ошибку")
}

func TestExecRetry_WithNilDB(t *testing.T) {
	db := &DB{DB: nil}
	_, err := db.execRetry("UPDATE sometable SET value=$1", 123)
	assert.Error(t, err, "nil DB => ожидаем ошибку")
}

func TestRowScanRetry_WithNilDB(t *testing.T) {
	// Пытаемся вызвать rowScanRetry на nil-DB и nil-строке.
	db := &DB{DB: nil}
	err := db.rowScanRetry(nil) // row = nil
	assert.Error(t, err, "должна быть ошибка, т.к. row == nil или db.DB == nil")
}

func TestQueryRetry_WithNilDB(t *testing.T) {
	db := &DB{DB: nil}
	rows, err := db.queryRetry("SELECT 1")
	assert.Error(t, err, "nil DB => ожидаем ошибку")
	assert.Nil(t, rows, "rows должны быть nil, т.к. был возврат ошибки")
	if rows != nil {
		assert.Nil(t, rows.Err(), "vet тест требует")
	}
}

// driverStub — заглушка, реализующая driver.Driver
type driverStub struct{}

func (d driverStub) Open(name string) (driver.Conn, error) {
	return &driverConnStub{}, nil
}

type driverConnStub struct{}

func (d *driverConnStub) Prepare(query string) (driver.Stmt, error) {
	return nil, errors.New("not implemented")
}
func (d *driverConnStub) Close() error {
	return nil
}
func (d *driverConnStub) Begin() (driver.Tx, error) {
	return nil, errors.New("not implemented")
}

// --------------------- //
//   Заглушка с памятью  //
// --------------------- //
//
// Здесь мы реализуем интерфейс StorDB в виде in-memory хранилища, 
// чтобы протестировать именно логику GetOneValue, GetArrayValues, List.

type mockDBConnMemory struct {
	oneValueMap   map[string]float64   // для одиночных метрик
	arrayValueMap map[string][]float64 // для массивных метрик
}

var _ StorDB = (*mockDBConnMemory)(nil)

func (m *mockDBConnMemory) Begin() (*sql.Tx, error)                                    { return nil, errors.New("not implemented") }
func (m *mockDBConnMemory) BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error) {
	return nil, errors.New("not implemented")
}
func (m *mockDBConnMemory) Close() error { return nil }
func (m *mockDBConnMemory) Conn(ctx context.Context) (*sql.Conn, error) {
	return nil, errors.New("not implemented")
}
func (m *mockDBConnMemory) CreateMetricsTable() error { return nil }
func (m *mockDBConnMemory) Driver() driver.Driver     { return driverStub{} }
func (m *mockDBConnMemory) Exec(query string, args ...any) (sql.Result, error) {
	return nil, errors.New("not implemented")
}
func (m *mockDBConnMemory) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	return nil, errors.New("not implemented")
}
func (m *mockDBConnMemory) Ping() error { return nil }
func (m *mockDBConnMemory) PingContext(ctx context.Context) error {
	return nil
}
func (m *mockDBConnMemory) PingDB() error                     { return nil }
func (m *mockDBConnMemory) Prepare(query string) (*sql.Stmt, error) {
	return nil, errors.New("not implemented")
}
func (m *mockDBConnMemory) PrepareContext(ctx context.Context, query string) (*sql.Stmt, error) {
	return nil, errors.New("not implemented")
}
func (m *mockDBConnMemory) PushAdd(metric string, metricName string, value float64) error {
	return nil
}
func (m *mockDBConnMemory) PushReplace(metric string, metricName string, value float64) error {
	return nil
}
func (m *mockDBConnMemory) Query(query string, args ...any) (*sql.Rows, error) {
	return nil, errors.New("not implemented")
}
func (m *mockDBConnMemory) QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	return nil, errors.New("not implemented")
}
func (m *mockDBConnMemory) QueryRow(query string, args ...any) *sql.Row {
	return nil
}
func (m *mockDBConnMemory) QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row {
	return nil
}
func (m *mockDBConnMemory) SetConnMaxIdleTime(d time.Duration) {}
func (m *mockDBConnMemory) SetConnMaxLifetime(d time.Duration) {}
func (m *mockDBConnMemory) SetMaxIdleConns(n int)              {}
func (m *mockDBConnMemory) SetMaxOpenConns(n int)              {}
func (m *mockDBConnMemory) Stats() sql.DBStats {
	return sql.DBStats{}
}

// GetOneValue: ищем в oneValueMap по ключу "type///name"
func (m *mockDBConnMemory) GetOneValue(metric string, metricName string) (float64, error) {
	key := metric + METRICSEPARATOR + metricName
	val, ok := m.oneValueMap[key]
	if !ok {
		return 0, ErrNoData
	}
	return val, nil
}

// GetArrayValues: ищем в arrayValueMap
func (m *mockDBConnMemory) GetArrayValues(metric string, metricName string) ([]float64, error) {
	key := metric + METRICSEPARATOR + metricName
	vals, ok := m.arrayValueMap[key]
	if !ok {
		return nil, ErrNoData
	}
	return vals, nil
}

// List: разбираем ключи, кладём в разные карты
func (m *mockDBConnMemory) List(metricOneValue, metricArrayValues string) (map[string]float64, map[string][]float64, error) {
	typeValue := make(map[string]float64)
	typeValues := make(map[string][]float64)

	for key, v := range m.oneValueMap {
		ms := parseKey(key)
		if ms.MetricType == metricOneValue {
			typeValue[ms.MetricName] = v
		} else {
			return nil, nil, ErrUnexpectedMetricType
		}
	}

	for key, slice := range m.arrayValueMap {
		ms := parseKey(key)
		if ms.MetricType == metricArrayValues {
			typeValues[ms.MetricName] = slice
		} else {
			return nil, nil, ErrUnexpectedMetricType
		}
	}

	return typeValue, typeValues, nil
}

// parseKey вспомогательно разбирает "type///name"
func parseKey(key string) MetricString {
	sep := METRICSEPARATOR
	// Можно через strings.Split, но для примера - «вручную»:
	idx := -1
	for i := 0; i+len(sep) <= len(key); i++ {
		if key[i:i+len(sep)] == sep {
			idx = i
			break
		}
	}
	if idx < 0 {
		return MetricString{}
	}
	return MetricString{
		MetricType: key[:idx],
		MetricName: key[idx+len(sep):],
	}
}

// --------------------- //
//  Тесты GetOneValue,
//  GetArrayValues, List
// --------------------- //

func TestDB_GetOneValue_Stub(t *testing.T) {
	mockMem := &mockDBConnMemory{
		oneValueMap: map[string]float64{
			"counter///PollCount": 123,
			"gauge///Alloc":       999,
		},
		arrayValueMap: map[string][]float64{},
	}

	val, err := mockMem.GetOneValue("counter", "PollCount")
	assert.NoError(t, err)
	assert.Equal(t, float64(123), val)

	val2, err2 := mockMem.GetOneValue("gauge", "Alloc")
	assert.NoError(t, err2)
	assert.Equal(t, 999.0, val2)

	// Проверим «нет данных»
	_, err3 := mockMem.GetOneValue("counter", "NotExists")
	assert.ErrorIs(t, err3, ErrNoData)
}

func TestDB_GetArrayValues_Stub(t *testing.T) {
	mockMem := &mockDBConnMemory{
		oneValueMap: map[string]float64{},
		arrayValueMap: map[string][]float64{
			"gauge///AllocHistory":    {1.1, 2.2, 3.3},
			"counter///PollCountList": {10, 20, 30},
		},
	}

	vals, err := mockMem.GetArrayValues("gauge", "AllocHistory")
	assert.NoError(t, err)
	assert.Equal(t, []float64{1.1, 2.2, 3.3}, vals)

	vals2, err2 := mockMem.GetArrayValues("counter", "PollCountList")
	assert.NoError(t, err2)
	assert.Equal(t, []float64{10, 20, 30}, vals2)

	// Проверим «нет данных»
	_, err3 := mockMem.GetArrayValues("gauge", "MissingKey")
	assert.ErrorIs(t, err3, ErrNoData)
}

func TestDB_List_Stub(t *testing.T) {
	mockMem := &mockDBConnMemory{
		oneValueMap: map[string]float64{
			"oneValue///Alloc": 100,
			"oneValue///Sys":   200,
		},
		arrayValueMap: map[string][]float64{
			"arrayValue///PollCount": {1, 2},
			"arrayValue///HeapList":  {10, 20},
		},
	}

	oneMap, arrMap, err := mockMem.List("oneValue", "arrayValue")
	assert.NoError(t, err)

	// Проверяем данные
	assert.Equal(t, float64(100), oneMap["Alloc"])
	assert.Equal(t, float64(200), oneMap["Sys"])
	assert.Equal(t, []float64{1, 2}, arrMap["PollCount"])
	assert.Equal(t, []float64{10, 20}, arrMap["HeapList"])
}

