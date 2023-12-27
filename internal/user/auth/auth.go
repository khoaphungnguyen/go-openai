package userauth

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt"
)

const (
	// Define the default expiration times for tokens.
	DefaultAccessTokenDuration  = 2 * time.Hour
	DefaultRefreshTokenDuration = 7 * 24 * time.Hour
)

type JwtWrapper struct {
	SecretKey              string        // Key used for signing the JWT token
	Issuer                 string        // Issuer of the JWT token
	AccessTokenExpiration  time.Duration // Expiration time of the JWT token
	RefreshTokenExpiration time.Duration // Expiration time of the Refresh token
}

type CustomClaims struct {
	UserID   string `json:"userId"`
	FullName string `json:"fullName"`
	jwt.StandardClaims
}

// GenerateToken generates a JWT token with custom claims.
func (j *JwtWrapper) GenerateToken(userID, fullName string) (string, error) {
	claims := &CustomClaims{
		UserID:   userID,
		FullName: fullName,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().UTC().Add(j.AccessTokenExpiration).Unix(),
			Issuer:    j.Issuer,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(j.SecretKey))
}

// RefreshToken generates a refresh JWT token with a longer lifespan.
func (j *JwtWrapper) RefreshToken(userID, fullName string) (string, error) {
	claims := &CustomClaims{
		UserID:   userID,
		FullName: fullName,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().UTC().Add(j.RefreshTokenExpiration).Unix(),
			Issuer:    j.Issuer,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(j.SecretKey))
}

// ValidateToken validates the JWT token.
func (j *JwtWrapper) ValidateToken(signedToken string) (*CustomClaims, error) {
	token, err := jwt.ParseWithClaims(signedToken, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(j.SecretKey), nil
	})

	if err != nil || !token.Valid {
		return nil, fmt.Errorf("invalid or expired token")
	}

	claims, ok := token.Claims.(*CustomClaims)
	if !ok {
		return nil, fmt.Errorf("invalid token claims")
	}

	return claims, nil
}
