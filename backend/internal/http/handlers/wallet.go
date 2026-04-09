package handlers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"bet/backend/internal/http/middleware"
	"bet/backend/internal/wallet"
)

type WalletHandler struct {
	service *wallet.Service
}

func NewWalletHandler(service *wallet.Service) *WalletHandler {
	return &WalletHandler{service: service}
}

func (h *WalletHandler) GetWallet(c *gin.Context) {
	claims, ok := middleware.ClaimsFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	w, err := h.service.GetWallet(claims.Subject)
	if err != nil {
		if errors.Is(err, wallet.ErrInvalidWalletInput) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid wallet request"})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get wallet"})
		return
	}

	c.JSON(http.StatusOK, w)
}

func (h *WalletHandler) ListTransactions(c *gin.Context) {
	claims, ok := middleware.ClaimsFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	items, err := h.service.ListTransactions(claims.Subject)
	if err != nil {
		if errors.Is(err, wallet.ErrInvalidWalletInput) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid wallet request"})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list wallet transactions"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"items": items})
}
