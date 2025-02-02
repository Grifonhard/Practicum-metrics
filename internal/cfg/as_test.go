package cfg

import (
	"encoding/json"
	"flag"
	"log"
	"os"
	"path/filepath"
	"testing"

	"github.com/Grifonhard/Practicum-metrics/internal/logger"
)

func TestLoadConfigFromFileAgent(t *testing.T) {
	err := logger.Init(os.Stdout, 4)
	if err != nil {
		log.Fatal(err)
	}

	// Создаём временный файл с тестовой конфигурацией
	tmpDir := t.TempDir()
	cfgFilePath := filepath.Join(tmpDir, "agent_config.json")

	type interm struct {
		Address        *string `json:"address"`
		ReportInterval *string `json:"report_interval"`
		PollInterval   *string `json:"poll_interval"`
		CryptoKey      *string `json:"crypto_key"`
		Key            *string `json:"key"`
		RateLimit      *int    `json:"rate_limit"`
	}

	testConfig := interm{
		Address:        strPtr("127.0.0.1:9999"),
		ReportInterval: strPtr("60s"),
		PollInterval:   strPtr("10s"),
		CryptoKey:      strPtr("/tmp/crypto_key.pub"),
		Key:            strPtr("my_secret_key"),
		RateLimit:      intPtr(10),
	}

	data, err := json.Marshal(testConfig)
	if err != nil {
		t.Fatalf("failed to marshal test config: %v", err)
	}

	if err := os.WriteFile(cfgFilePath, data, 0600); err != nil {
		t.Fatalf("failed to write test config file: %v", err)
	}

	// Ставим переменные окружения
	os.Setenv("CONFIG", cfgFilePath)
	defer func() {
		// Очищаем, чтобы не влиять на другие тесты
		os.Unsetenv("CONFIG")
	}()

	var agent Agent

	// Перед запуском Load нужно сбросить состояние flag,
	// иначе если в других тестах уже парсили флаги — они сохранятся.
	flag.CommandLine = flag.NewFlagSet("test", flag.ContinueOnError)

	// Загружаем конфигурацию
	if err := agent.Load(); err != nil {
		t.Fatalf("Agent.Load() returned an error: %v", err)
	}

	// Проверяем, что данные загружены корректно
	if agent.Addr == nil || *agent.Addr != *testConfig.Address {
		t.Errorf("expected address %q, got %v", *testConfig.Address, *agent.Addr)
	}
	if agent.ReportInterval == nil || *agent.ReportInterval != 60 {
		t.Errorf("expected report interval %s, got %v", *testConfig.ReportInterval, *agent.ReportInterval)
	}
	if agent.PollInterval == nil || *agent.PollInterval != 10 {
		t.Errorf("expected poll interval %s, got %v", *testConfig.PollInterval, *agent.PollInterval)
	}
	if agent.Key == nil || *agent.Key != *testConfig.Key {
		t.Errorf("expected key %q, got %v", *testConfig.Key, *agent.Key)
	}
	if agent.RateLimit == nil || *agent.RateLimit != *testConfig.RateLimit {
		t.Errorf("expected rate limit %d, got %v", *testConfig.RateLimit, *agent.RateLimit)
	}
	if agent.CryptoKey == nil || *agent.CryptoKey != *testConfig.CryptoKey {
		t.Errorf("expected crypto key %q, got %v", *testConfig.CryptoKey, *agent.CryptoKey)
	}
}

func TestLoadConfigFromFileServer(t *testing.T) {
	err := logger.Init(os.Stdout, 4)
	if err != nil {
		log.Fatal(err)
	}

	// Создаём временный файл с тестовой конфигурацией
	tmpDir := t.TempDir()
	cfgFilePath := filepath.Join(tmpDir, "server_config.json")

	type interm struct {
		Address       *string `json:"address"`
		StoreInterval *string `json:"store_interval"`
		Restore       *bool   `json:"restore"`
		StoreFile     *string `json:"store_file"`
		DatabaseDSN   *string `json:"database_dsn"`
		Key           *string `json:"key"`
		CryptoKey     *string `json:"crypto_key"`
	}

	testConfig := interm{
		Address:       strPtr("127.0.0.1:9999"),
		StoreInterval: strPtr("60s"),
		Restore:       boolPtr(true),
		StoreFile:     strPtr("/some/backup"),
		DatabaseDSN:   strPtr("/path/to/db"),
		Key:           strPtr("my_secret_key"),
		CryptoKey:     strPtr("/tmp/crypto_key.pub"),
	}

	data, err := json.Marshal(testConfig)
	if err != nil {
		t.Fatalf("failed to marshal test config: %v", err)
	}

	if err := os.WriteFile(cfgFilePath, data, 0600); err != nil {
		t.Fatalf("failed to write test config file: %v", err)
	}

	// Ставим переменные окружения
	os.Setenv("CONFIG", cfgFilePath)
	defer func() {
		// Очищаем, чтобы не влиять на другие тесты
		os.Unsetenv("CONFIG")
	}()

	var server Server

	// Перед запуском Load нужно сбросить состояние flag,
	// иначе если в других тестах уже парсили флаги — они сохранятся.
	flag.CommandLine = flag.NewFlagSet("test", flag.ContinueOnError)

	// Загружаем конфигурацию
	if err := server.Load(); err != nil {
		t.Fatalf("Agent.Load() returned an error: %v", err)
	}

	// Проверяем, что данные загружены корректно
	if server.Addr == nil || *server.Addr != *testConfig.Address {
		t.Errorf("expected address %q, got %v", *testConfig.Address, *server.Addr)
	}
	if server.StoreInterval == nil || *server.StoreInterval != 60 {
		t.Errorf("expected store interval %s, got %v", *testConfig.StoreInterval, *server.StoreInterval)
	}
	if server.Restore == nil || *server.Restore != *testConfig.Restore {
		t.Errorf("expected restore %t, got %v", *testConfig.Restore, *server.Restore)
	}
	if server.FileStoragePath == nil || *server.FileStoragePath != *testConfig.StoreFile {
		t.Errorf("expected file store path %s, got %v", *testConfig.StoreFile, *server.FileStoragePath)
	}
	if server.DatabaseDsn == nil || *server.DatabaseDsn != *testConfig.DatabaseDSN {
		t.Errorf("expected db path %s, got %v", *testConfig.DatabaseDSN, *server.DatabaseDsn)
	}
	if server.Key == nil || *server.Key != *testConfig.Key {
		t.Errorf("expected key %s, got %v", *testConfig.Key, *server.Key)
	}
	if server.CryptoKey == nil || *server.CryptoKey != *testConfig.CryptoKey {
		t.Errorf("expected crypto key %s, got %v", *testConfig.CryptoKey, *server.CryptoKey)
	}
}

func TestLoadConfigFromFlags(t *testing.T) {
	err := logger.Init(os.Stdout, 4)
	if err != nil {
		log.Fatal(err)
	}

	// Сбросим состояние флагов
	flag.CommandLine = flag.NewFlagSet("test", flag.ContinueOnError)

	// Подменить os.Args — как будто программа запущена с этими аргументами
	os.Args = []string{
		"testbinary", // первый аргумент – это имя бинаря
		"-a", "localhost:9000",
		"-r", "15",
		"-p", "5",
		"-k", "flag_secret_key",
		"-l", "11",
		"-crypto-key", "/tmp/crypto_key.pub",
	}

	agent := Agent{}

	// Вызваем Load(), которое внутри само объявит и спарсит флаги
	if err := agent.Load(); err != nil {
		t.Fatalf("Agent.Load() returned an error: %v", err)
	}

	// Проверка
	if agent.Addr == nil || *agent.Addr != "localhost:9000" {
		t.Errorf("expected address %q, got %v", "localhost:9000", agent.Addr)
	}
	if agent.ReportInterval == nil || *agent.ReportInterval != 15 {
		t.Errorf("expected report interval %d, got %v", 15, agent.ReportInterval)
	}
	if agent.PollInterval == nil || *agent.PollInterval != 5 {
		t.Errorf("expected poll interval %d, got %v", 5, agent.PollInterval)
	}
	if agent.Key == nil || *agent.Key != "flag_secret_key" {
		t.Errorf("expected key %q, got %v", "flag_secret_key", agent.Key)
	}
	if agent.RateLimit == nil || *agent.RateLimit != 11 {
		t.Errorf("expected rate limit %d, got %v", 11, agent.RateLimit)
	}
	if agent.CryptoKey == nil || *agent.CryptoKey != "/tmp/crypto_key.pub" {
		t.Errorf("expected cryptoKey %q, got %v", "/tmp/crypto_key.pub", agent.CryptoKey)
	}
}

func TestServerLoadConfigFromFlags(t *testing.T) {
	err := logger.Init(os.Stdout, 4)
	if err != nil {
		log.Fatal(err)
	}

	flag.CommandLine = flag.NewFlagSet("test", flag.ContinueOnError)

	os.Args = []string{
		"testbinary",
		"-a", "localhost:9001",
		"-i", "300",
		"-f", "/tmp/server_storage.json",
		"-r",
		"-d", "postgres://usr:pwd@localhost/db",
		"-k", "server_secret_key",
		"-crypto-key", "/tmp/server_crypto_key",
	}

	srv := Server{}
	if err := srv.Load(); err != nil {
		t.Fatalf("Server.Load() returned an error: %v", err)
	}

	if srv.Addr == nil || *srv.Addr != "localhost:9001" {
		t.Errorf("expected address = localhost:9001, got %v", srv.Addr)
	}
	if srv.StoreInterval == nil || *srv.StoreInterval != 300 {
		t.Errorf("expected store interval = 300, got %v", srv.StoreInterval)
	}
	if srv.FileStoragePath == nil || *srv.FileStoragePath != "/tmp/server_storage.json" {
		t.Errorf("expected file storage path = /tmp/server_storage.json, got %v", srv.FileStoragePath)
	}
	if srv.Restore == nil || *srv.Restore != true {
		t.Errorf("expected restore = true, got %v", srv.Restore)
	}
	if srv.DatabaseDsn == nil || *srv.DatabaseDsn != "postgres://usr:pwd@localhost/db" {
		t.Errorf("expected database dsn = postgres://usr:pwd@localhost/db, got %v", srv.DatabaseDsn)
	}
	if srv.Key == nil || *srv.Key != "server_secret_key" {
		t.Errorf("expected key = server_secret_key, got %v", srv.Key)
	}
	if srv.CryptoKey == nil || *srv.CryptoKey != "/tmp/server_crypto_key" {
		t.Errorf("expected crypto key = /tmp/server_crypto_key, got %v", srv.CryptoKey)
	}
}

func TestLoadConfigFromEnv(t *testing.T) {
	err := logger.Init(os.Stdout, 4)
	if err != nil {
		log.Fatal(err)
	}

	// Сброс флагов (на всякий случай, чтобы флаги не влияли)
	flag.CommandLine = flag.NewFlagSet("test", flag.ContinueOnError)

	// Ставим переменные окружения
	os.Setenv("ADDRESS", "env:8080")
	os.Setenv("REPORT_INTERVAL", "25")
	os.Setenv("POLL_INTERVAL", "7")
	os.Setenv("KEY", "env_secret_key")
	os.Setenv("RATE_LIMIT", "13")
	os.Setenv("CRYPTO_KEY", "/env/crypto/key")
	defer func() {
		// Очищаем, чтобы не влиять на другие тесты
		os.Unsetenv("ADDRESS")
		os.Unsetenv("REPORT_INTERVAL")
		os.Unsetenv("POLL_INTERVAL")
		os.Unsetenv("KEY")
		os.Unsetenv("RATE_LIMIT")
		os.Unsetenv("CRYPTO_KEY")
	}()

	agent := Agent{}
	if err := agent.Load(); err != nil {
		t.Fatalf("Agent.Load() returned an error: %v", err)
	}

	// Проверяем, что данные загружены из ENV
	if agent.Addr == nil || *agent.Addr != "env:8080" {
		t.Errorf("expected address %q, got %v", "env:8080", agent.Addr)
	}
	if agent.ReportInterval == nil || *agent.ReportInterval != 25 {
		t.Errorf("expected report interval %d, got %v", 25, agent.ReportInterval)
	}
	if agent.PollInterval == nil || *agent.PollInterval != 7 {
		t.Errorf("expected poll interval %d, got %v", 7, agent.PollInterval)
	}
	if agent.Key == nil || *agent.Key != "env_secret_key" {
		t.Errorf("expected key %q, got %v", "env_secret_key", agent.Key)
	}
	if agent.RateLimit == nil || *agent.RateLimit != 13 {
		t.Errorf("expected rate limit %d, got %v", 13, agent.RateLimit)
	}
	if agent.CryptoKey == nil || *agent.CryptoKey != "/env/crypto/key" {
		t.Errorf("expected cryptoKey %q, got %v", "/env/crypto/key", agent.CryptoKey)
	}
}

func TestServerLoadConfigFromEnv(t *testing.T) {
	err := logger.Init(os.Stdout, 4)
	if err != nil {
		log.Fatal(err)
	}

	// Сброс флагов (чтобы предыдущие тесты не влияли на текущий)
	flag.CommandLine = flag.NewFlagSet("test", flag.ContinueOnError)

	// Устанавливаем переменные окружения
	os.Setenv("ADDRESS", "env.server:8080")
	os.Setenv("STORE_INTERVAL", "123")
	os.Setenv("FILE_STORAGE_PATH", "/env/server_storage.json")
	os.Setenv("RESTORE", "true")
	os.Setenv("DATABASE_DSN", "env_postgres://usr:pwd@localhost/env_db")
	os.Setenv("KEY", "env_server_secret_key")
	os.Setenv("CRYPTO_KEY", "/env/server_crypto_key")

	// Очищаем переменные окружения в конце теста, чтобы не влиять на другие тесты
	defer func() {
		os.Unsetenv("ADDRESS")
		os.Unsetenv("STORE_INTERVAL")
		os.Unsetenv("FILE_STORAGE_PATH")
		os.Unsetenv("RESTORE")
		os.Unsetenv("DATABASE_DSN")
		os.Unsetenv("KEY")
		os.Unsetenv("CRYPTO_KEY")
	}()

	// Создаём структуру Server
	srv := Server{}

	// Вызываем метод Load
	if err := srv.Load(); err != nil {
		t.Fatalf("Server.Load() returned an error: %v", err)
	}

	// Проверяем, что все поля заполнены значениями из ENV
	if srv.Addr == nil || *srv.Addr != "env.server:8080" {
		t.Errorf("expected Address = %q, got %v", "env.server:8080", srv.Addr)
	}

	if srv.StoreInterval == nil || *srv.StoreInterval != 123 {
		t.Errorf("expected StoreInterval = %d, got %v", 123, srv.StoreInterval)
	}

	if srv.FileStoragePath == nil || *srv.FileStoragePath != "/env/server_storage.json" {
		t.Errorf("expected FileStoragePath = %q, got %v", "/env/server_storage.json", srv.FileStoragePath)
	}

	if srv.Restore == nil || *srv.Restore != true {
		t.Errorf("expected Restore = true, got %v", srv.Restore)
	}

	if srv.DatabaseDsn == nil || *srv.DatabaseDsn != "env_postgres://usr:pwd@localhost/env_db" {
		t.Errorf("expected DatabaseDsn = %q, got %v", "env_postgres://usr:pwd@localhost/env_db", srv.DatabaseDsn)
	}

	if srv.Key == nil || *srv.Key != "env_server_secret_key" {
		t.Errorf("expected Key = %q, got %v", "env_server_secret_key", srv.Key)
	}

	if srv.CryptoKey == nil || *srv.CryptoKey != "/env/server_crypto_key" {
		t.Errorf("expected CryptoKey = %q, got %v", "/env/server_crypto_key", srv.CryptoKey)
	}
}

func TestLoadConfigPriority(t *testing.T) {
	err := logger.Init(os.Stdout, 4)
	if err != nil {
		log.Fatal(err)
	}

	// Приоритет: ENV > flags > file > default.
	// Зададим все три источника и проверим, что итоговые значения взяты из ENV.

	// 1. Создаём файл с конфигом (самый низкий приоритет после defaults)
	tmpDir := t.TempDir()
	cfgFilePath := filepath.Join(tmpDir, "agent_config.json")

	fileConfig := AgentFile{
		Address:        strPtr("file.address:1111"),
		ReportInterval: intPtr(2222),
		PollInterval:   intPtr(3333),
		CryptoKey:      strPtr("file_crypto"),
		Key:            strPtr("file_key"),
		RateLimit:      intPtr(44),
	}

	data, err := json.Marshal(fileConfig)
	if err != nil {
		t.Fatalf("failed to marshal fileConfig: %v", err)
	}
	if err := os.WriteFile(cfgFilePath, data, 0600); err != nil {
		t.Fatalf("failed to write file config: %v", err)
	}

	// 2. Подготавливаем «флаги» (средний приоритет)
	// Вместо ручного Parse() — подменим os.Args так,
	// будто программу запустили с этими флагами.
	flag.CommandLine = flag.NewFlagSet("test", flag.ContinueOnError)
	os.Args = []string{
		"testbinary",
		"-a", "flag.address:2222",
		"-r", "50",
		"-p", "60",
		"-k", "flag_key",
		"-l", "70",
		"-crypto-key", "flag_crypto",
		"-c", cfgFilePath, // указываем файл конфигурации
	}

	// 3. Устанавливаем ENV (самый высокий приоритет)
	os.Setenv("ADDRESS", "env.address:3333")
	os.Setenv("REPORT_INTERVAL", "99")
	os.Setenv("POLL_INTERVAL", "88")
	os.Setenv("KEY", "env_key")
	os.Setenv("RATE_LIMIT", "77")
	os.Setenv("CRYPTO_KEY", "env_crypto")
	defer func() {
		os.Unsetenv("ADDRESS")
		os.Unsetenv("REPORT_INTERVAL")
		os.Unsetenv("POLL_INTERVAL")
		os.Unsetenv("KEY")
		os.Unsetenv("RATE_LIMIT")
		os.Unsetenv("CRYPTO_KEY")
	}()

	// 4. Создаём Agent и вызываем Load()
	agent := Agent{}
	if err := agent.Load(); err != nil {
		t.Fatalf("Agent.Load() returned an error: %v", err)
	}

	// 5. Проверяем, что в итоге взяты именно ENV (а не флаги или файл)
	if agent.Addr == nil || *agent.Addr != "env.address:3333" {
		t.Errorf("expected address %q, got %v", "env.address:3333", agent.Addr)
	}
	if agent.ReportInterval == nil || *agent.ReportInterval != 99 {
		t.Errorf("expected report interval %d, got %v", 99, agent.ReportInterval)
	}
	if agent.PollInterval == nil || *agent.PollInterval != 88 {
		t.Errorf("expected poll interval %d, got %v", 88, agent.PollInterval)
	}
	if agent.Key == nil || *agent.Key != "env_key" {
		t.Errorf("expected key %q, got %v", "env_key", agent.Key)
	}
	if agent.RateLimit == nil || *agent.RateLimit != 77 {
		t.Errorf("expected rate limit %d, got %v", 77, agent.RateLimit)
	}
	if agent.CryptoKey == nil || *agent.CryptoKey != "env_crypto" {
		t.Errorf("expected cryptoKey %q, got %v", "env_crypto", agent.CryptoKey)
	}
}

// Хэлпер-функции для сокращения записи
func strPtr(s string) *string {
	return &s
}

func intPtr(i int) *int {
	return &i
}

func boolPtr(b bool) *bool {
	return &b
}
