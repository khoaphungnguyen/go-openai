package utils

import (
    "crypto/rand"
    "crypto/subtle"
    "encoding/base64"
    "errors"

    "golang.org/x/crypto/argon2"
)

const (
    saltLength = 32 // 32-bit salt length for added security
)

func HashPassword(password string) (string, string, error) {
    // Generate a salt with increased length
    salt := make([]byte, saltLength)
    if _, err := rand.Read(salt); err != nil {
        return "", "", err
    }

    // Argon2 hashing with recommended parameters
    hash := argon2.IDKey([]byte(password), salt, 1, 64*1024, 4, 32)
    return base64.StdEncoding.EncodeToString(hash), base64.StdEncoding.EncodeToString(salt), nil
}

func CheckPassword(storedPassword, storedSalt, providedPassword string) error {
    salt, err := base64.StdEncoding.DecodeString(storedSalt)
    if err != nil {
        return err
    }

    hash := argon2.IDKey([]byte(providedPassword), salt, 1, 64*1024, 4, 32)
    providedHash := base64.StdEncoding.EncodeToString(hash)

    // Constant-time comparison to prevent timing attacks
    if subtle.ConstantTimeCompare([]byte(storedPassword), []byte(providedHash)) != 1 {
        return errors.New("invalid credentials") // Generic error message
    }
    return nil
}
