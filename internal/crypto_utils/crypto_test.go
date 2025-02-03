package cryptoutils

import (
	"bytes"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

// TestLoadPublicKey_Success проверяет успешную загрузку корректного RSA паблик-ключа.
func TestLoadPublicKey_Success(t *testing.T) {
	// 1. Генерируем временную пару ключей RSA
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	pubBytes, err := x509.MarshalPKIXPublicKey(&priv.PublicKey)
	require.NoError(t, err)

	// 2. Кодируем в PEM
	pubPem := pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: pubBytes,
	})

	// 3. Пишем во временный файл
	tmpDir := t.TempDir()
	pubPath := filepath.Join(tmpDir, "public.pem")
	err = os.WriteFile(pubPath, pubPem, 0600)
	require.NoError(t, err)

	// 4. Вызываем LoadPublicKey
	pubKey, err := LoadPublicKey(pubPath)
	require.NoError(t, err)
	require.NotNil(t, pubKey, "должны получить rsa.PublicKey")

	// 5. Доп.проверка - сверим, что открытый ключ совпадает
	require.Equal(t, priv.PublicKey.N, pubKey.N, "N public keys must match")
	require.Equal(t, priv.PublicKey.E, pubKey.E, "E must match")
}

// TestLoadPublicKey_FileNotFound проверяет, что получим ошибку, если файл не существует.
func TestLoadPublicKey_FileNotFound(t *testing.T) {
	pubKey, err := LoadPublicKey("non_existent_file.pem")
	require.Error(t, err)
	require.Nil(t, pubKey)
}

// TestLoadPublicKey_ParsePEMError проверяет, что будет ошибка, если в файле не PEM.
func TestLoadPublicKey_ParsePEMError(t *testing.T) {
	// Создаём "левый" файл
	tmpDir := t.TempDir()
	pubPath := filepath.Join(tmpDir, "public.pem")
	err := os.WriteFile(pubPath, []byte("not a valid pem data"), 0600)
	require.NoError(t, err)

	pubKey, err := LoadPublicKey(pubPath)
	require.ErrorIs(t, err, ErrParsePEMpubl)
	require.Nil(t, pubKey)
}

// TestEncryptRSA_Success проверяет, что шифрование проходит без ошибок
// и результат можно расшифровать нашим приватным ключом.
func TestEncryptRSA_Success(t *testing.T) {
	// Генерим пару ключей
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	pub := &priv.PublicKey

	originalData := []byte(`{"field":"value"}`)

	// Зашифровываем
	encryptedBase64, err := EncryptRSA(originalData, pub)
	require.NoError(t, err)
	require.NotEmpty(t, encryptedBase64)

	// Расшифруем через наш priv (проверка корректности)
	decrypted, err := rsaDecryptBase64(encryptedBase64, priv)
	require.NoError(t, err)
	require.Equal(t, originalData, decrypted)
}

//
// Тесты на LoadPrivateKey
//

func TestLoadPrivateKey_Success(t *testing.T) {
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	// Пишем приватный ключ в PEM
	privBytes := x509.MarshalPKCS1PrivateKey(priv)
	privPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: privBytes,
	})

	tmpDir := t.TempDir()
	privPath := filepath.Join(tmpDir, "private.pem")
	err = os.WriteFile(privPath, privPEM, 0600)
	require.NoError(t, err)

	loadedPriv, err := LoadPrivateKey(privPath)
	require.NoError(t, err)
	require.NotNil(t, loadedPriv)

	// Сверим поле N,E,D
	require.Equal(t, priv.N, loadedPriv.N)
	require.Equal(t, priv.E, loadedPriv.E)
	require.Equal(t, priv.D, loadedPriv.D)
}

func TestLoadPrivateKey_FileNotFound(t *testing.T) {
	privKey, err := LoadPrivateKey("no_such_file.pem")
	require.Error(t, err)
	require.Nil(t, privKey)
}

func TestLoadPrivateKey_ParsePEMError(t *testing.T) {
	tmpDir := t.TempDir()
	privPath := filepath.Join(tmpDir, "private.pem")
	// Кладём некорректный PEM
	err := os.WriteFile(privPath, []byte("not a valid key"), 0600)
	require.NoError(t, err)

	privKey, err := LoadPrivateKey(privPath)
	require.Error(t, err)
	require.Nil(t, privKey)
}

//
// Тесты на Middleware DecryptBody
//
func TestDecryptBodyMiddleware_Success(t *testing.T) {
	// 1. Генерим пару ключей
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	pub := &priv.PublicKey

	// 2. Устанавливаем глобальную переменную PrivateKey
	PrivateKey = priv
	defer func() { PrivateKey = nil }()

	// 3. Зашифруем тестовые данные
	originalJSON := []byte(`{"hello":"world"}`)
	encryptedBase64, err := EncryptRSA(originalJSON, pub)
	require.NoError(t, err)

	// 4. Подготовим JSON-оболочку вида {"data": "<base64>"}
	payload, err := json.Marshal(map[string]string{
		"data": encryptedBase64,
	})
	require.NoError(t, err)

	// 5. Поднимаем тестовый gin.Engine с нашим middleware
	r := gin.Default()
	r.Use(DecryptBody())
	r.POST("/test", func(c *gin.Context) {
		// Здесь тело уже должно быть расшифровано
		body, err := io.ReadAll(c.Request.Body)
		require.NoError(t, err)

		// Проверяем, что расшифровалось ровно в originalJSON
		require.Equal(t, originalJSON, body)

		// Например, просто вернём "ok"
		c.String(http.StatusOK, "ok")
	})

	// 6. Делаем тестовый запрос
	req := httptest.NewRequest(http.MethodPost, "/test", bytes.NewBuffer(payload))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	require.Equal(t, "ok", w.Body.String())
}

func TestDecryptBodyMiddleware_NoPrivateKey(t *testing.T) {
	// Убедимся, что PrivateKey = nil
	PrivateKey = nil
	defer func() { PrivateKey = nil }()

	// Формируем произвольный JSON. Мидлварь должна просто пропустить запрос,
	// т.к. внутри DecryptBody() написано: "if PrivateKey == nil { c.Next(); return }"
	payload := []byte(`{"some":"data"}`)

	// Поднимаем тестовый роутер
	r := gin.Default()
	r.Use(DecryptBody())
	r.POST("/test", func(c *gin.Context) {
		// Проверим, что тело НЕ расшифровалось (оно и не зашифровывалось).
		body, err := io.ReadAll(c.Request.Body)
		require.NoError(t, err)
		require.Equal(t, payload, body)

		c.String(http.StatusOK, "ok")
	})

	req := httptest.NewRequest(http.MethodPost, "/test", bytes.NewBuffer(payload))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	require.Equal(t, "ok", w.Body.String())
}

func TestDecryptBodyMiddleware_InvalidBase64(t *testing.T) {
	// Ставим приватный ключ
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	PrivateKey = priv
	defer func() { PrivateKey = nil }()

	// Формируем JSON вида {"data":"NOT_BASE64"}
	payload, _ := json.Marshal(map[string]string{
		"data": "NOT_BASE64",
	})

	r := gin.Default()
	r.Use(DecryptBody())
	r.POST("/test", func(c *gin.Context) {
		// Если дошли сюда — значит, не отработала проверка
		t.Error("Expected to fail earlier on invalid base64")
		c.String(http.StatusOK, "ok")
	})

	req := httptest.NewRequest(http.MethodPost, "/test", bytes.NewBuffer(payload))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	// Ожидаем 400 Bad Request
	require.Equal(t, http.StatusBadRequest, w.Code)
	require.Contains(t, w.Body.String(), "invalid base64 data")
}

func TestDecryptBodyMiddleware_InvalidJSON(t *testing.T) {
	// Ставим приватный ключ
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	PrivateKey = priv
	defer func() { PrivateKey = nil }()

	// Формируем некорректный JSON
	payload := []byte(`NOT VALID JSON`)

	r := gin.Default()
	r.Use(DecryptBody())
	r.POST("/test", func(c *gin.Context) {
		t.Error("Expected to fail earlier on invalid JSON")
		c.String(http.StatusOK, "ok")
	})

	req := httptest.NewRequest(http.MethodPost, "/test", bytes.NewBuffer(payload))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusBadRequest, w.Code)
	require.Contains(t, w.Body.String(), "invalid encrypted JSON")
}

//
// ------------------
// Вспомогательные функции
// ------------------
//

// rsaDecryptBase64 — утилита для расшифровки base64-закодированных данных приватным ключом.
// Используем в тестах, чтобы проверить корректность EncryptRSA.
func rsaDecryptBase64(b64 string, priv *rsa.PrivateKey) ([]byte, error) {
	encrypted, err := base64.StdEncoding.DecodeString(b64)
	if err != nil {
		return nil, err
	}
	decryptedBytes, err := rsa.DecryptOAEP(
		hashFunc().New(),
		rand.Reader,
		priv,
		encrypted,
		nil,
	)
	if err != nil {
		return nil, err
	}
	return decryptedBytes, nil
}

func hashFunc() crypto.Hash {
	return crypto.SHA256
}
