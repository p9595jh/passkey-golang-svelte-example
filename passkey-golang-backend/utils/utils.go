package utils

import "crypto/rand"

// Generates random bytes as the given size.
func RandID(size int) ([]byte, error) {
	id := make([]byte, size)
	_, err := rand.Read(id)
	return id, err
}
