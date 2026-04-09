package handlers

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"bet/backend/internal/bets"
	"bet/backend/internal/events"
	"bet/backend/internal/http/middleware"
)

type EventsHandler struct {
	service     *events.Service
	betsService *bets.Service
}

func NewEventsHandler(service *events.Service, betsService *bets.Service) *EventsHandler {
	return &EventsHandler{service: service, betsService: betsService}
}

type createEventRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Category    string `json:"category"`
	ResolveAt   string `json:"resolve_at"`
}

func (h *EventsHandler) ListEvents(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"items": h.service.ListApprovedEvents()})
}

func (h *EventsHandler) GetEvent(c *gin.Context) {
	e, ok := h.service.GetEventByID(c.Param("id"))
	if !ok || e.Status != "approved" {
		c.JSON(http.StatusNotFound, gin.H{"error": "event not found"})
		return
	}

	c.JSON(http.StatusOK, e)
}

func (h *EventsHandler) CreateEvent(c *gin.Context) {
	claims, ok := middleware.ClaimsFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req createEventRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	resolveAt, err := time.Parse(time.RFC3339, strings.TrimSpace(req.ResolveAt))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "resolve_at must be RFC3339"})
		return
	}

	e, err := h.service.CreateEvent(claims.Subject, req.Title, req.Description, req.Category, resolveAt)
	if err != nil {
		if errors.Is(err, events.ErrInvalidEventInput) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid event input"})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create event"})
		return
	}

	c.JSON(http.StatusCreated, e)
}

func (h *EventsHandler) ListModerationEvents(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"items": h.service.ListPendingModeration()})
}

func (h *EventsHandler) ApproveEvent(c *gin.Context) {
	claims, ok := middleware.ClaimsFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	e, err := h.service.ApproveEvent(c.Param("id"), claims.Subject)
	if err != nil {
		switch {
		case errors.Is(err, events.ErrEventNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "event not found"})
		case errors.Is(err, events.ErrModerationAlreadyHandled):
			c.JSON(http.StatusConflict, gin.H{"error": "moderation already handled"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to approve event"})
		}
		return
	}

	c.JSON(http.StatusOK, e)
}

type rejectEventRequest struct {
	Reason string `json:"reason"`
}

func (h *EventsHandler) RejectEvent(c *gin.Context) {
	claims, ok := middleware.ClaimsFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req rejectEventRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	e, err := h.service.RejectEvent(c.Param("id"), claims.Subject, req.Reason)
	if err != nil {
		switch {
		case errors.Is(err, events.ErrEventNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "event not found"})
		case errors.Is(err, events.ErrInvalidModerationReason):
			c.JSON(http.StatusBadRequest, gin.H{"error": "reason is required"})
		case errors.Is(err, events.ErrModerationAlreadyHandled):
			c.JSON(http.StatusConflict, gin.H{"error": "moderation already handled"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to reject event"})
		}
		return
	}

	c.JSON(http.StatusOK, e)
}

type settleEventRequest struct {
	WinnerOutcome string `json:"winner_outcome"`
}

func (h *EventsHandler) SettleEvent(c *gin.Context) {
	var req settleEventRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	e, err := h.service.SettleEvent(c.Param("id"), req.WinnerOutcome)
	if err != nil {
		switch {
		case errors.Is(err, events.ErrEventNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "event not found"})
		case errors.Is(err, events.ErrInvalidSettlementInput):
			c.JSON(http.StatusBadRequest, gin.H{"error": "winner_outcome must be yes or no"})
		case errors.Is(err, events.ErrEventNotSettlable):
			c.JSON(http.StatusConflict, gin.H{"error": "event is not settlable"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to settle event"})
		}
		return
	}

	settledBets, err := h.betsService.SettleEventBets(e.ID, e.WinnerOutcome)
	if err != nil {
		if errors.Is(err, bets.ErrInvalidSettlement) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid settlement"})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to settle bets"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"event":        e,
		"settled_bets": settledBets,
	})
}
