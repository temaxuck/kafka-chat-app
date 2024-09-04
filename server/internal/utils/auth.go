package utils

import (
	"crypto/rand"
	"encoding/base64"

	"github.com/google/uuid"
)

func GenerateRoomId(length int) (string, error) {
	bytes := make([]byte, length)
	_, err := rand.Read(bytes)

	if err != nil {
		return "", nil
	}

	return base64.RawStdEncoding.EncodeToString(bytes)[:length], nil
}

func GenerateUUID() string {
	return uuid.NewString()
}
