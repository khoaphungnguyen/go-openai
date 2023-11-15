package userauth

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt"
)

type JwtWrapper struct {
	SecretKey         string // key used for signing the JWT token
	Issuer            string // Issuer of the JWT token
	ExpirationMinutes int64  // Number of minutes the JWT token will be valid for
	ExpirationHours   int64  // Expiration time of the JWT token in hours
}

// GenerateToken generates a jwt token with email as the subject
func (j *JwtWrapper) GenerateToken(email string) (signedToken string, err error) {
	claims := &jwt.StandardClaims{
		Subject:   email,
		ExpiresAt: time.Now().UTC().Add(time.Minute * time.Duration(j.ExpirationMinutes)).Unix(),
		Issuer:    j.Issuer,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err = token.SignedString([]byte(j.SecretKey))
	if err != nil {
		return "", fmt.Errorf("error signing token: %w", err)
	}
	return
}

// RefreshToken generates a refresh jwt token with a longer lifespan
func (j *JwtWrapper) RefreshToken(email string) (signedtoken string, err error) {
	claims := &jwt.StandardClaims{
		Subject:   email,
		ExpiresAt: time.Now().UTC().Add(time.Hour * time.Duration(j.ExpirationHours)).Unix(),
		Issuer:    j.Issuer,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedtoken, err = token.SignedString([]byte(j.SecretKey))
	if err != nil {
		return "", fmt.Errorf("error signing refresh token: %w", err)
	}
	return
}

// ValidateToken validates the jwt token
func (j *JwtWrapper) ValidateToken(signedToken string) (claims *jwt.StandardClaims, err error) {
	token, err := jwt.ParseWithClaims(
		signedToken,
		&jwt.StandardClaims{},
		func(token *jwt.Token) (interface{}, error) {
			// Validate the alg is what you expect:
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}

			return []byte(j.SecretKey), nil
		},
	)

	if err != nil {
		return nil, fmt.Errorf("error parsing token: %w", err)
	}

	claims, ok := token.Claims.(*jwt.StandardClaims)
	if !ok {
		return nil, errors.New("could not parse claims")
	}

	if claims.ExpiresAt < time.Now().UTC().Unix() {
		return nil, errors.New("JWT is expired")
	}
	return
}
