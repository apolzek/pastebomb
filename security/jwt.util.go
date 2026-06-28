package security

import (
	"fmt"
	"os"
	"sync"

	"github.com/golang-jwt/jwt/v5"
)

const (
	envJWTSecret = "JWT_SECRET"
	// devFallbackSecret is only used when JWT_SECRET is unset (dev/test).
	// Production deployments MUST set JWT_SECRET.
	devFallbackSecret = "dev-only-do-not-use-in-prod"
)

var (
	secretOnce sync.Once
	secretKey  []byte
)

func loadSecret() []byte {
	secretOnce.Do(func() {
		if v := os.Getenv(envJWTSecret); v != "" {
			secretKey = []byte(v)
			return
		}
		// Allow tests / local dev to work without the env var, but make it
		// loud at startup so misconfiguration is visible.
		fmt.Fprintln(os.Stderr, "WARNING: JWT_SECRET not set, using insecure dev fallback")
		secretKey = []byte(devFallbackSecret)
	})
	return secretKey
}

func GenerateToken(claims *jwt.MapClaims) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	webtoken, err := token.SignedString(loadSecret())
	if err != nil {
		return "", err
	}
	return webtoken, nil
}

func VerifyToken(tokenString string) (*jwt.Token, error) {
	tokenJwt, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
		if _, isValid := t.Method.(*jwt.SigningMethodHMAC); !isValid {
			return nil, fmt.Errorf("invalid signing method: %v", t.Header["alg"])
		}
		return loadSecret(), nil
	})
	if err != nil {
		return nil, err
	}
	return tokenJwt, nil
}

func DecodeToken(tokenString string) (jwt.MapClaims, error) {
	token, err := VerifyToken(tokenString)
	if err != nil {
		return nil, err
	}

	claims, isOK := token.Claims.(jwt.MapClaims)
	if isOK && token.Valid {
		return claims, nil
	}
	return nil, fmt.Errorf("invalid token")
}
