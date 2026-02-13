package auth

import (
	"encoding/base64"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// TokenProvider handles JWT creation and validation.
// Compatible with the existing Spring Boot TokenProvider (HS512, Base64-decoded key).
type TokenProvider struct {
	signingKey []byte
}

// NewTokenProvider creates a TokenProvider from a Base64-encoded secret key.
func NewTokenProvider(base64Secret string) (*TokenProvider, error) {
	keyBytes, err := base64.StdEncoding.DecodeString(base64Secret)
	if err != nil {
		// Try URL-safe base64
		keyBytes, err = base64.URLEncoding.DecodeString(base64Secret)
		if err != nil {
			// Try raw (no padding)
			keyBytes, err = base64.RawStdEncoding.DecodeString(base64Secret)
			if err != nil {
				return nil, fmt.Errorf("decoding JWT secret: %w", err)
			}
		}
	}
	return &TokenProvider{signingKey: keyBytes}, nil
}

// CreateToken creates a JWT token with the given username as subject.
// Matches Java: signWith(HS512), setSubject(username), setIssuer("LostarkTodo"), setIssuedAt(now)
func (tp *TokenProvider) CreateToken(username string) (string, error) {
	claims := jwt.MapClaims{
		"sub": username,
		"iss": "LostarkTodo",
		"iat": time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)
	return token.SignedString(tp.signingKey)
}

// ValidateToken validates a JWT token and returns the subject (username).
func (tp *TokenProvider) ValidateToken(tokenString string) (string, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return tp.signingKey, nil
	})
	if err != nil {
		return "", fmt.Errorf("parsing token: %w", err)
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return "", fmt.Errorf("invalid token claims")
	}

	subject, ok := claims["sub"].(string)
	if !ok {
		return "", fmt.Errorf("missing subject in token")
	}

	return subject, nil
}
