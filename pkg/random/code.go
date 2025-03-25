// Package random is designed for generating random strings.
package random

import (
	"crypto/rand"
)

const (
	// letters contains the characters used for generating random strings.
	letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
)

// GenerateRandomString generates a random string of the specified length.
//
// The function uses a cryptographically secure random number generator (`crypto/rand`)
// and selects characters from the `letters` set.
//
// Parameters:
//   - length: the length of the generated string.
//
// Returns:
//   - a string of randomly generated characters;
//   - an error if the random data generation fails.
//
// Example usage:
//
//	str, err := GenerateRandomString(10)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Println("Random string:", str)
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
