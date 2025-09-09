package auth

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"math/big"
)

func RandString(nByte int) (string, error) {
	b := make([]byte, nByte)
	if _, err := io.ReadFull(rand.Reader, b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%^&*()_+"
const length = 24

func GenerateSecurePassword() ([]byte, error) {
	password := make([]byte, length)
	charsetLength := big.NewInt(int64(len(charset)))
	for i := range password {
		index, err := rand.Int(rand.Reader, charsetLength)
		if err != nil {
			return nil, fmt.Errorf("error generating random index: %v", err)
		}
		password[i] = charset[index.Int64()]
	}

	return password, nil
}
