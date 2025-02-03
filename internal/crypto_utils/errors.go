package cryptoutils

import "errors"

var (
	ErrParsePEMpubl    = errors.New("failed to parse PEM block from public key")
	ErrParsePEMprivate = errors.New("failed to parse PEM block from private key")
	ErrNotPublicKey    = errors.New("not RSA public key")
)
