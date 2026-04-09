package handlers

import (
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"bet/backend/internal/auth"
	"bet/backend/internal/notifications"
)

type AuthHandler struct {
	service       *auth.Service
	emailSender   notifications.Sender
	authJWTSecret string
	authTokenTTL  time.Duration
	verifyBaseURL string
}

func NewAuthHandler(service *auth.Service, sender notifications.Sender, jwtSecret string, tokenTTL time.Duration, verifyBaseURL string) *AuthHandler {
	return &AuthHandler{
		service:       service,
		emailSender:   sender,
		authJWTSecret: jwtSecret,
		authTokenTTL:  tokenTTL,
		verifyBaseURL: verifyBaseURL,
	}
}

type registerRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req registerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	user, err := h.service.Register(req.Email, req.Password)
	if err != nil {
		switch {
		case errors.Is(err, auth.ErrEmailExists):
			c.JSON(http.StatusConflict, gin.H{"error": "email already exists"})
		case errors.Is(err, auth.ErrInvalidCredentials):
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid credentials"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to register"})
		}
		return
	}

	verifyLink := auth.BuildVerifyLink(h.verifyBaseURL, user.VerifyToken)
	if err := h.emailSender.Send(notifications.Message{
		To:      user.Email,
		Subject: "Bet MVP: подтвердите email",
		Body:    "Спасибо за регистрацию! Подтвердите email: " + verifyLink,
	}); err != nil {
		log.Printf("failed to send verify email to %s: %v", user.Email, err)
	}

	c.JSON(http.StatusCreated, gin.H{
		"id":             user.ID,
		"email":          user.Email,
		"role":           user.Role,
		"email_verified": user.EmailVerified,
	})
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	user, err := h.service.Login(req.Email, req.Password)
	if err != nil {
		if errors.Is(err, auth.ErrInvalidCredentials) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
			return
		}

		if errors.Is(err, auth.ErrEmailNotVerified) {
			c.JSON(http.StatusForbidden, gin.H{"error": "email is not verified"})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to login"})
		return
	}

	token, err := auth.IssueToken(h.authJWTSecret, h.tokenTTLDuration(), user.ID, user.Role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to issue token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"access_token": token,
		"token_type":   "Bearer",
	})
}

func (h *AuthHandler) VerifyEmail(c *gin.Context) {
	token := c.Query("token")
	if token == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "token is required"})
		return
	}

	user, err := h.service.VerifyEmail(token)
	if err != nil {
		if errors.Is(err, auth.ErrInvalidVerifyToken) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid token"})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to verify email"})
		return
	}

	if err := h.emailSender.Send(notifications.Message{
		To:      user.Email,
		Subject: "Bet MVP: email подтвержден",
		Body:    "Ваш email успешно подтвержден.",
	}); err != nil {
		log.Printf("failed to send email confirmation notice to %s: %v", user.Email, err)
	}

	c.JSON(http.StatusOK, gin.H{
		"id":             user.ID,
		"email":          user.Email,
		"email_verified": user.EmailVerified,
	})
}

func (h *AuthHandler) Me(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "authorized"})
}

func (h *AuthHandler) tokenTTLDuration() time.Duration {
	if h.authTokenTTL <= 0 {
		return time.Hour
	}

	return h.authTokenTTL
}
