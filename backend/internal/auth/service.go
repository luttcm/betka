package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
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
	db           *sql.DB
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

func NewServiceWithDB(db *sql.DB) *Service {
	if db == nil {
		return NewService()
	}

	return &Service{db: db}
}

func (s *Service) Register(email, password string) (User, error) {
	normalizedEmail := strings.ToLower(strings.TrimSpace(email))
	if normalizedEmail == "" || strings.TrimSpace(password) == "" {
		return User{}, ErrInvalidCredentials
	}

	if s.db != nil {
		u := User{
			Email:         normalizedEmail,
			PasswordHash:  hashPassword(password),
			Role:          "user",
			EmailVerified: false,
			VerifyToken:   randomToken(24),
		}

		var id int64
		err := s.db.QueryRow(
			`INSERT INTO users (email, password_hash, role, status, email_verified, verify_token, created_at)
			 VALUES ($1, $2, $3, 'active', $4, $5, NOW())
			 RETURNING id`,
			u.Email,
			u.PasswordHash,
			u.Role,
			u.EmailVerified,
			u.VerifyToken,
		).Scan(&id)
		if err != nil {
			if isUniqueViolation(err) {
				return User{}, ErrEmailExists
			}

			return User{}, err
		}

		u.ID = formatIntID(id)
		return u, nil
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
	if s.db != nil {
		var (
			id            int64
			dbEmail       string
			passwordHash  string
			role          string
			emailVerified bool
			verifyToken   sql.NullString
		)

		err := s.db.QueryRow(
			`SELECT id, email, password_hash, role, email_verified, verify_token
			 FROM users
			 WHERE email = $1 AND status = 'active'`,
			normalizedEmail,
		).Scan(&id, &dbEmail, &passwordHash, &role, &emailVerified, &verifyToken)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return User{}, ErrInvalidCredentials
			}

			return User{}, err
		}

		if passwordHash != hashPassword(password) {
			return User{}, ErrInvalidCredentials
		}

		if !emailVerified {
			return User{}, ErrEmailNotVerified
		}

		return User{
			ID:            formatIntID(id),
			Email:         dbEmail,
			PasswordHash:  passwordHash,
			Role:          role,
			EmailVerified: emailVerified,
			VerifyToken:   verifyToken.String,
		}, nil
	}

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

	if s.db != nil {
		var (
			id    int64
			email string
			role  string
		)
		err := s.db.QueryRow(
			`UPDATE users
			 SET email_verified = TRUE, verify_token = ''
			 WHERE verify_token = $1 AND verify_token <> ''
			 RETURNING id, email, role`,
			normalizedToken,
		).Scan(&id, &email, &role)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return User{}, ErrInvalidVerifyToken
			}

			return User{}, err
		}

		return User{
			ID:            formatIntID(id),
			Email:         email,
			Role:          role,
			EmailVerified: true,
			VerifyToken:   "",
		}, nil
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

	if s.db != nil {
		var existingID int64
		err := s.db.QueryRow(`SELECT id FROM users WHERE email = $1`, normalizedEmail).Scan(&existingID)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return User{}, err
		}

		passwordHash := hashPassword(password)
		if errors.Is(err, sql.ErrNoRows) {
			var createdID int64
			if err := s.db.QueryRow(
				`INSERT INTO users (email, password_hash, role, status, email_verified, verify_token, created_at)
				 VALUES ($1, $2, 'admin', 'active', TRUE, '', NOW())
				 RETURNING id`,
				normalizedEmail,
				passwordHash,
			).Scan(&createdID); err != nil {
				return User{}, err
			}
			existingID = createdID
		} else {
			if _, err := s.db.Exec(
				`UPDATE users
				 SET password_hash = $1, role = 'admin', status = 'active', email_verified = TRUE, verify_token = ''
				 WHERE id = $2`,
				passwordHash,
				existingID,
			); err != nil {
				return User{}, err
			}
		}

		return User{
			ID:            formatIntID(existingID),
			Email:         normalizedEmail,
			PasswordHash:  passwordHash,
			Role:          "admin",
			EmailVerified: true,
			VerifyToken:   "",
		}, nil
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

func formatIntID(id int64) string {
	return fmt.Sprintf("%d", id)
}

func isUniqueViolation(err error) bool {
	return strings.Contains(strings.ToLower(err.Error()), "unique")
}
