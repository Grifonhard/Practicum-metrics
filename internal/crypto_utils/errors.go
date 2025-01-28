package cryptoutils

import "errors"

var (
	ErrParsePEM = errors.New("failed to parse PEM block from public key")
	ErrNotPublicKey = errors.New("not RSA public key")
)