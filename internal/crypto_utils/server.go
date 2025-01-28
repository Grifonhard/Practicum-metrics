package cryptoutils

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"io"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

var PrivateKey *rsa.PrivateKey

func LoadPrivateKey(path string) (*rsa.PrivateKey, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	block, _ := pem.Decode(data)
	if block == nil {
		return nil, ErrParsePEMprivate
	}
	priv, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	return priv, nil
}

// DecryptBody Middleware для расшифровки тела запроса
func DecryptBody() gin.HandlerFunc {
	return func(c *gin.Context) {
		if PrivateKey == nil {
			c.Next()
			return
		}

		body, err := io.ReadAll(c.Request.Body)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "cannot read body"})
			return
		}

		c.Request.Body.Close()

		var wrapper struct {
			Data string `json:"data"`
		}
		if err := json.Unmarshal(body, &wrapper); err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid encrypted JSON"})
			return
		}

		encrypted, err := base64.StdEncoding.DecodeString(wrapper.Data)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid base64 data"})
			return
		}

		decryptedBytes, err := rsa.DecryptOAEP(
			sha256.New(),
			rand.Reader,
			PrivateKey,
			encrypted,
			nil,
		)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "decryption error"})
			return
		}

		c.Request.Body = io.NopCloser(bytes.NewBuffer(decryptedBytes))

		c.Next()
	}
}
