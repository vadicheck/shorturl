package random

import (
	"crypto/rand"
)

const (
	letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
)

func GenerateRandomString(length int) (string, error) {
	result := make([]byte, length)
	maxInt := byte(len(letters))

	randomBytes := make([]byte, length)
	_, err := rand.Read(randomBytes)
	if err != nil {
		return "", err
	}

	for i := range result {
		result[i] = letters[randomBytes[i]%maxInt]
	}

	return string(result), nil
}
