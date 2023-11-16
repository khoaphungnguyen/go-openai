package utils

import (
	"crypto/rand"
	"encoding/base64"
	"errors"

	"golang.org/x/crypto/argon2"
)

func HashPassword(password string) (string, string, error) {
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return "", "", err
	}
	hash := argon2.IDKey([]byte(password), salt, 1, 64*1024, 4, 32)
	return base64.StdEncoding.EncodeToString(hash), base64.StdEncoding.EncodeToString(salt), nil
}

func CheckPassword(storedPassword, storedSalt, providedPassword string) error {

	salt, err := base64.StdEncoding.DecodeString(storedSalt)
	if err != nil {
		return err
	}

	hash := argon2.IDKey([]byte(providedPassword), salt, 1, 64*1024, 4, 32)
	if storedPassword != base64.StdEncoding.EncodeToString(hash) {
		return errors.New("invalid password")
	}
	return nil
}
