package utils

import (
	"crypto/rand"
	"encoding/base64"
)

// Generates random bytes as the given size.
func RandID(size int) ([]byte, error) {
	id := make([]byte, size)
	_, err := rand.Read(id)
	return id, err
}

func Base64(b []byte) string {
	return base64.URLEncoding.EncodeToString(b)
}
