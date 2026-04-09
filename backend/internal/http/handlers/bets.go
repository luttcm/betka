package handlers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"bet/backend/internal/bets"
	"bet/backend/internal/http/middleware"
	"bet/backend/internal/wallet"
)

type BetsHandler struct {
	service *bets.Service
}

func NewBetsHandler(service *bets.Service) *BetsHandler {
	return &BetsHandler{service: service}
}

type placeBetRequest struct {
	EventID     string  `json:"event_id"`
	OutcomeCode string  `json:"outcome_code"`
	Stake       float64 `json:"stake"`
}

func (h *BetsHandler) PlaceBet(c *gin.Context) {
	claims, ok := middleware.ClaimsFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	idempotencyKey := c.GetHeader("Idempotency-Key")
	if idempotencyKey == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Idempotency-Key header is required"})
		return
	}

	var req placeBetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	bet, created, err := h.service.PlaceBet(claims.Subject, req.EventID, req.OutcomeCode, idempotencyKey, req.Stake)
	if err != nil {
		switch {
		case errors.Is(err, bets.ErrMissingIdempotencyKey):
			c.JSON(http.StatusBadRequest, gin.H{"error": "Idempotency-Key header is required"})
		case errors.Is(err, bets.ErrInvalidBetInput):
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid bet input"})
		case errors.Is(err, bets.ErrEventUnavailable):
			c.JSON(http.StatusConflict, gin.H{"error": "event is unavailable for betting"})
		case errors.Is(err, wallet.ErrInsufficientFunds):
			c.JSON(http.StatusConflict, gin.H{"error": "insufficient funds"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to place bet"})
		}
		return
	}

	if created {
		c.JSON(http.StatusCreated, bet)
		return
	}

	c.JSON(http.StatusOK, bet)
}

func (h *BetsHandler) ListMyBets(c *gin.Context) {
	claims, ok := middleware.ClaimsFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	items, err := h.service.ListMyBets(claims.Subject)
	if err != nil {
		if errors.Is(err, bets.ErrInvalidBetInput) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid bet request"})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list bets"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"items": items})
}
