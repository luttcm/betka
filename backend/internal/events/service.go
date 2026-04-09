package events

import (
	"errors"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"
)

var (
	ErrInvalidEventInput        = errors.New("invalid event input")
	ErrEventNotFound            = errors.New("event not found")
	ErrModerationAlreadyHandled = errors.New("moderation task already handled")
	ErrInvalidModerationReason  = errors.New("invalid moderation reason")
)

type Event struct {
	ID            string    `json:"id"`
	CreatorUserID string    `json:"creator_user_id"`
	Title         string    `json:"title"`
	Description   string    `json:"description"`
	Category      string    `json:"category"`
	ResolveAt     time.Time `json:"resolve_at"`
	Status        string    `json:"status"`
	CreatedAt     time.Time `json:"created_at"`
}

type ModerationTask struct {
	ID          string     `json:"id"`
	EventID     string     `json:"event_id"`
	Status      string     `json:"status"`
	ModeratorID string     `json:"moderator_id,omitempty"`
	Reason      string     `json:"reason,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	ReviewedAt  *time.Time `json:"reviewed_at,omitempty"`
}

type Service struct {
	mu              sync.RWMutex
	eventsByID      map[string]*Event
	moderationByEvt map[string]*ModerationTask
	eventSeq        int64
	taskSeq         int64
}

func NewService() *Service {
	return &Service{
		eventsByID:      make(map[string]*Event),
		moderationByEvt: make(map[string]*ModerationTask),
	}
}

func (s *Service) CreateEvent(creatorUserID, title, description, category string, resolveAt time.Time) (Event, error) {
	title = strings.TrimSpace(title)
	description = strings.TrimSpace(description)
	creatorUserID = strings.TrimSpace(creatorUserID)

	if creatorUserID == "" || title == "" || description == "" || !resolveAt.After(time.Now().UTC()) {
		return Event{}, ErrInvalidEventInput
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.eventSeq++
	eventID := fmt.Sprintf("evt_%d", s.eventSeq)

	e := &Event{
		ID:            eventID,
		CreatorUserID: creatorUserID,
		Title:         title,
		Description:   description,
		Category:      strings.TrimSpace(category),
		ResolveAt:     resolveAt.UTC(),
		Status:        "pending",
		CreatedAt:     time.Now().UTC(),
	}

	s.taskSeq++
	taskID := fmt.Sprintf("mod_%d", s.taskSeq)
	t := &ModerationTask{
		ID:        taskID,
		EventID:   eventID,
		Status:    "pending",
		CreatedAt: time.Now().UTC(),
	}

	s.eventsByID[eventID] = e
	s.moderationByEvt[eventID] = t

	return *e, nil
}

func (s *Service) ListApprovedEvents() []Event {
	s.mu.RLock()
	defer s.mu.RUnlock()

	items := make([]Event, 0)
	for _, e := range s.eventsByID {
		if e.Status == "approved" {
			items = append(items, *e)
		}
	}

	sort.Slice(items, func(i, j int) bool {
		return items[i].CreatedAt.After(items[j].CreatedAt)
	})

	return items
}

func (s *Service) GetEventByID(id string) (Event, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	e, ok := s.eventsByID[strings.TrimSpace(id)]
	if !ok {
		return Event{}, false
	}

	return *e, true
}

type ModerationQueueItem struct {
	Task  ModerationTask `json:"task"`
	Event Event          `json:"event"`
}

func (s *Service) ListPendingModeration() []ModerationQueueItem {
	s.mu.RLock()
	defer s.mu.RUnlock()

	items := make([]ModerationQueueItem, 0)
	for eventID, task := range s.moderationByEvt {
		if task.Status != "pending" {
			continue
		}

		event, ok := s.eventsByID[eventID]
		if !ok {
			continue
		}

		items = append(items, ModerationQueueItem{Task: *task, Event: *event})
	}

	sort.Slice(items, func(i, j int) bool {
		return items[i].Task.CreatedAt.Before(items[j].Task.CreatedAt)
	})

	return items
}

func (s *Service) ApproveEvent(eventID, moderatorID string) (Event, error) {
	return s.review(eventID, moderatorID, "", "approved")
}

func (s *Service) RejectEvent(eventID, moderatorID, reason string) (Event, error) {
	if strings.TrimSpace(reason) == "" {
		return Event{}, ErrInvalidModerationReason
	}

	return s.review(eventID, moderatorID, reason, "rejected")
}

func (s *Service) review(eventID, moderatorID, reason, targetStatus string) (Event, error) {
	eventID = strings.TrimSpace(eventID)
	moderatorID = strings.TrimSpace(moderatorID)

	s.mu.Lock()
	defer s.mu.Unlock()

	e, ok := s.eventsByID[eventID]
	if !ok {
		return Event{}, ErrEventNotFound
	}

	t, ok := s.moderationByEvt[eventID]
	if !ok {
		return Event{}, ErrEventNotFound
	}

	if t.Status != "pending" || e.Status != "pending" {
		return Event{}, ErrModerationAlreadyHandled
	}

	now := time.Now().UTC()
	t.Status = targetStatus
	t.ModeratorID = moderatorID
	t.Reason = strings.TrimSpace(reason)
	t.ReviewedAt = &now
	e.Status = targetStatus

	return *e, nil
}
