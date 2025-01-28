package cryptoutils

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"os"
)

var PublicKey *rsa.PublicKey

func LoadPublicKey(path string) (*rsa.PublicKey, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	block, _ := pem.Decode(data)
	if block == nil {
		return nil, ErrParsePEM
	}
	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	rsaPub, ok := pub.(*rsa.PublicKey)
	if !ok {
		return nil, ErrNotPublicKey
	}
	return rsaPub, nil
}

func EncryptRSA(jsonData []byte, pub *rsa.PublicKey) (string, error) {
	encryptedBytes, err := rsa.EncryptOAEP(
		sha256.New(),
		rand.Reader,
		pub,
		jsonData,
		nil,
	)
	if err != nil {
		return "", err
	}
	// Обычно передаём двоичные данные как base64-строку
	return base64.StdEncoding.EncodeToString(encryptedBytes), nil
}
