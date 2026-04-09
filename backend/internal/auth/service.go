package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"sync"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrEmailExists        = errors.New("email already exists")
	ErrEmailNotVerified   = errors.New("email is not verified")
	ErrInvalidVerifyToken = errors.New("invalid verify token")
)

type User struct {
	ID            string
	Email         string
	PasswordHash  string
	Role          string
	EmailVerified bool
	VerifyToken   string
}

type Service struct {
	mu           sync.RWMutex
	usersByEmail map[string]*User
	usersByID    map[string]*User
}

func NewService() *Service {
	return &Service{
		usersByEmail: make(map[string]*User),
		usersByID:    make(map[string]*User),
	}
}

func (s *Service) Register(email, password string) (User, error) {
	normalizedEmail := strings.ToLower(strings.TrimSpace(email))
	if normalizedEmail == "" || strings.TrimSpace(password) == "" {
		return User{}, ErrInvalidCredentials
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.usersByEmail[normalizedEmail]; exists {
		return User{}, ErrEmailExists
	}

	u := &User{
		ID:            "usr_" + randomToken(12),
		Email:         normalizedEmail,
		PasswordHash:  hashPassword(password),
		Role:          "user",
		EmailVerified: false,
		VerifyToken:   randomToken(24),
	}

	s.usersByEmail[normalizedEmail] = u
	s.usersByID[u.ID] = u

	return *u, nil
}

func (s *Service) Login(email, password string) (User, error) {
	normalizedEmail := strings.ToLower(strings.TrimSpace(email))

	s.mu.RLock()
	u, exists := s.usersByEmail[normalizedEmail]
	s.mu.RUnlock()
	if !exists {
		return User{}, ErrInvalidCredentials
	}

	if u.PasswordHash != hashPassword(password) {
		return User{}, ErrInvalidCredentials
	}

	if !u.EmailVerified {
		return User{}, ErrEmailNotVerified
	}

	return *u, nil
}

func (s *Service) VerifyEmail(token string) (User, error) {
	normalizedToken := strings.TrimSpace(token)
	if normalizedToken == "" {
		return User{}, ErrInvalidVerifyToken
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	for _, u := range s.usersByID {
		if u.VerifyToken == normalizedToken {
			u.EmailVerified = true
			u.VerifyToken = ""
			return *u, nil
		}
	}

	return User{}, ErrInvalidVerifyToken
}

func (s *Service) BootstrapAdmin(email, password string) (User, error) {
	normalizedEmail := strings.ToLower(strings.TrimSpace(email))
	if normalizedEmail == "" || strings.TrimSpace(password) == "" {
		return User{}, ErrInvalidCredentials
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if existing, exists := s.usersByEmail[normalizedEmail]; exists {
		existing.PasswordHash = hashPassword(password)
		existing.Role = "admin"
		existing.EmailVerified = true
		existing.VerifyToken = ""
		return *existing, nil
	}

	u := &User{
		ID:            "usr_" + randomToken(12),
		Email:         normalizedEmail,
		PasswordHash:  hashPassword(password),
		Role:          "admin",
		EmailVerified: true,
		VerifyToken:   "",
	}

	s.usersByEmail[normalizedEmail] = u
	s.usersByID[u.ID] = u

	return *u, nil
}

func hashPassword(password string) string {
	digest := sha256.Sum256([]byte(password))
	return base64.RawURLEncoding.EncodeToString(digest[:])
}

func BuildVerifyLink(baseURL, token string) string {
	if strings.Contains(baseURL, "?") {
		return fmt.Sprintf("%s&token=%s", baseURL, token)
	}

	return fmt.Sprintf("%s?token=%s", baseURL, token)
}

func randomToken(size int) string {
	buf := make([]byte, size)
	if _, err := rand.Read(buf); err != nil {
		fallback := sha256.Sum256([]byte(fmt.Sprintf("fallback-%d", size)))
		return hex.EncodeToString(fallback[:])
	}

	return hex.EncodeToString(buf)
}
