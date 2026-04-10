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
	if !ok || (e.Status != "approved" && e.Status != "settlement_requested" && e.Status != "settled") {
		c.JSON(http.StatusNotFound, gin.H{"error": "event not found"})
		return
	}

	c.JSON(http.StatusOK, e)
}

func (h *EventsHandler) GetEventOdds(c *gin.Context) {
	odds, err := h.betsService.GetEventOdds(c.Param("id"))
	if err != nil {
		if errors.Is(err, bets.ErrInvalidBetInput) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid event id"})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch event odds"})
		return
	}

	c.JSON(http.StatusOK, odds)
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

type requestSettlementEvidenceFile struct {
	FileName string `json:"file_name"`
	FileData string `json:"file_data"`
}

type requestSettlementRequest struct {
	EvidenceURL  string                        `json:"evidence_url"`
	EvidenceFile requestSettlementEvidenceFile `json:"evidence_file"`
}

func (h *EventsHandler) RequestSettlement(c *gin.Context) {
	claims, ok := middleware.ClaimsFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req requestSettlementRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	e, err := h.service.RequestSettlement(
		c.Param("id"),
		claims.Subject,
		req.EvidenceURL,
		req.EvidenceFile.FileName,
		req.EvidenceFile.FileData,
	)
	if err != nil {
		switch {
		case errors.Is(err, events.ErrEventNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "event not found"})
		case errors.Is(err, events.ErrForbiddenSettlementRequest):
			c.JSON(http.StatusForbidden, gin.H{"error": "only event creator can request settlement"})
		case errors.Is(err, events.ErrInvalidSettlementEvidence):
			c.JSON(http.StatusBadRequest, gin.H{"error": "evidence url or evidence file (name+data) is required"})
		case errors.Is(err, events.ErrEventNotSettlable):
			c.JSON(http.StatusConflict, gin.H{"error": "event is not ready for settlement request"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to request settlement"})
		}
		return
	}

	c.JSON(http.StatusOK, e)
}

func (h *EventsHandler) ListSettlementRequests(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"items": h.service.ListSettlementRequests()})
}

func (h *EventsHandler) SettleEvent(c *gin.Context) {
	var req settleEventRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	e, settledBets, err := h.betsService.SettleEventAndBets(c.Param("id"), req.WinnerOutcome)
	if err != nil {
		switch {
		case errors.Is(err, events.ErrEventNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "event not found"})
		case errors.Is(err, events.ErrInvalidSettlementInput):
			c.JSON(http.StatusBadRequest, gin.H{"error": "winner_outcome must be yes or no"})
		case errors.Is(err, bets.ErrInvalidSettlement):
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid settlement"})
		case errors.Is(err, events.ErrEventNotSettlable):
			c.JSON(http.StatusConflict, gin.H{"error": "event is not settlable"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to settle event and bets"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"event":        e,
		"settled_bets": settledBets,
	})
}
