package id

import (
	"crypto/rand"
	"math/big"
)

// charset for ID generation (alphanumeric, lowercase)
const charset = "abcdefghijklmnopqrstuvwxyz0123456789"

// Generate creates a new random 3-character ID
func Generate() (string, error) {
	return generateWithLength(3)
}

// GenerateNoteID creates a note ID in the format "taskID-xxx"
func GenerateNoteID(taskID string) (string, error) {
	suffix, err := generateWithLength(3)
	if err != nil {
		return "", err
	}
	return taskID + "-" + suffix, nil
}

// GenerateUnique creates a unique 3-character ID that doesn't exist in the given set
func GenerateUnique(existing map[string]bool) (string, error) {
	const maxAttempts = 100
	for i := 0; i < maxAttempts; i++ {
		id, err := Generate()
		if err != nil {
			return "", err
		}
		if !existing[id] {
			return id, nil
		}
	}
	// If we can't find a unique ID after maxAttempts, fall back to 4 characters
	return generateWithLength(4)
}

// generateWithLength creates a random ID of the specified length
func generateWithLength(length int) (string, error) {
	result := make([]byte, length)
	charsetLen := big.NewInt(int64(len(charset)))

	for i := 0; i < length; i++ {
		n, err := rand.Int(rand.Reader, charsetLen)
		if err != nil {
			return "", err
		}
		result[i] = charset[n.Int64()]
	}

	return string(result), nil
}
