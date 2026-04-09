package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
)

type Claims struct {
	Subject string `json:"sub"`
	Role    string `json:"role"`
	Exp     int64  `json:"exp"`
}

func IssueToken(secret string, ttl time.Duration, subject, role string) (string, error) {
	if secret == "" {
		return "", errors.New("token secret is empty")
	}

	claims := Claims{
		Subject: subject,
		Role:    role,
		Exp:     time.Now().Add(ttl).Unix(),
	}

	headerRaw, err := json.Marshal(map[string]string{
		"alg": "HS256",
		"typ": "JWT",
	})
	if err != nil {
		return "", fmt.Errorf("marshal header: %w", err)
	}

	payloadRaw, err := json.Marshal(claims)
	if err != nil {
		return "", fmt.Errorf("marshal payload: %w", err)
	}

	header := base64.RawURLEncoding.EncodeToString(headerRaw)
	payload := base64.RawURLEncoding.EncodeToString(payloadRaw)
	signingInput := header + "." + payload

	signature := signHS256(secret, signingInput)

	return signingInput + "." + signature, nil
}

func ParseToken(secret, token string) (Claims, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return Claims{}, errors.New("invalid token format")
	}

	signingInput := parts[0] + "." + parts[1]
	expectedSig := signHS256(secret, signingInput)
	if !hmac.Equal([]byte(parts[2]), []byte(expectedSig)) {
		return Claims{}, errors.New("invalid token signature")
	}

	payloadRaw, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return Claims{}, errors.New("invalid token payload")
	}

	var claims Claims
	if err := json.Unmarshal(payloadRaw, &claims); err != nil {
		return Claims{}, errors.New("invalid token claims")
	}

	if claims.Subject == "" || claims.Role == "" {
		return Claims{}, errors.New("missing token claims")
	}

	if time.Now().Unix() >= claims.Exp {
		return Claims{}, errors.New("token expired")
	}

	return claims, nil
}

func signHS256(secret, data string) string {
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(data))
	return base64.RawURLEncoding.EncodeToString(h.Sum(nil))
}
