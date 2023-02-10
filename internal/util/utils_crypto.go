package util

import (
	"crypto/rand"
	"encoding/hex"
)

func RandomHex(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		logger.Errorf("cannot create %d-byte long random hex: %v\n", n, err)
		return "", ErrorInternal
	}

	return hex.EncodeToString(b), nil
}
