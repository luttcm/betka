package events

import (
	"database/sql"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

var (
	ErrInvalidEventInput        = errors.New("invalid event input")
	ErrEventNotFound            = errors.New("event not found")
	ErrModerationAlreadyHandled = errors.New("moderation task already handled")
	ErrInvalidModerationReason  = errors.New("invalid moderation reason")
	ErrInvalidSettlementInput   = errors.New("invalid settlement input")
	ErrEventNotSettlable        = errors.New("event is not settlable")
)

type Event struct {
	ID            string    `json:"id"`
	CreatorUserID string    `json:"creator_user_id"`
	Title         string    `json:"title"`
	Description   string    `json:"description"`
	Category      string    `json:"category"`
	ResolveAt     time.Time `json:"resolve_at"`
	Status        string    `json:"status"`
	WinnerOutcome string    `json:"winner_outcome,omitempty"`
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
	db              *sql.DB
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

func NewServiceWithDB(db *sql.DB) *Service {
	if db == nil {
		return NewService()
	}

	return &Service{db: db}
}

func (s *Service) CreateEvent(creatorUserID, title, description, category string, resolveAt time.Time) (Event, error) {
	title = strings.TrimSpace(title)
	description = strings.TrimSpace(description)
	creatorUserID = strings.TrimSpace(creatorUserID)

	if creatorUserID == "" || title == "" || description == "" || !resolveAt.After(time.Now().UTC()) {
		return Event{}, ErrInvalidEventInput
	}

	if s.db != nil {
		creatorID, err := parseIntID(creatorUserID)
		if err != nil {
			return Event{}, ErrInvalidEventInput
		}

		tx, err := s.db.Begin()
		if err != nil {
			return Event{}, err
		}
		defer tx.Rollback()

		var (
			eventID   int64
			createdAt time.Time
		)
		err = tx.QueryRow(
			`INSERT INTO events (creator_user_id, title, description, category, resolve_at, status, created_at)
			 VALUES ($1, $2, $3, $4, $5, 'pending', NOW())
			 RETURNING id, created_at`,
			creatorID,
			title,
			description,
			strings.TrimSpace(category),
			resolveAt.UTC(),
		).Scan(&eventID, &createdAt)
		if err != nil {
			return Event{}, err
		}

		if _, err := tx.Exec(`INSERT INTO event_outcomes (event_id, code) VALUES ($1, 'yes'), ($1, 'no')`, eventID); err != nil {
			return Event{}, err
		}

		if _, err := tx.Exec(
			`INSERT INTO moderation_tasks (event_id, status, created_at) VALUES ($1, 'pending', NOW())`,
			eventID,
		); err != nil {
			return Event{}, err
		}

		if err := tx.Commit(); err != nil {
			return Event{}, err
		}

		return Event{
			ID:            formatIntID(eventID),
			CreatorUserID: formatIntID(creatorID),
			Title:         title,
			Description:   description,
			Category:      strings.TrimSpace(category),
			ResolveAt:     resolveAt.UTC(),
			Status:        "pending",
			CreatedAt:     createdAt,
		}, nil
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
	if s.db != nil {
		rows, err := s.db.Query(
			`SELECT id, creator_user_id, title, description, category, resolve_at, status, COALESCE(winner_outcome, ''), created_at
			 FROM events
			 WHERE status = 'approved'
			 ORDER BY created_at DESC`,
		)
		if err != nil {
			return []Event{}
		}
		defer rows.Close()

		items := make([]Event, 0)
		for rows.Next() {
			var (
				eid, creatorID int64
				e              Event
			)
			if err := rows.Scan(&eid, &creatorID, &e.Title, &e.Description, &e.Category, &e.ResolveAt, &e.Status, &e.WinnerOutcome, &e.CreatedAt); err != nil {
				continue
			}

			e.ID = formatIntID(eid)
			e.CreatorUserID = formatIntID(creatorID)
			items = append(items, e)
		}

		return items
	}

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
	if s.db != nil {
		eventID, err := parseIntID(id)
		if err != nil {
			return Event{}, false
		}

		var (
			eid, creatorID int64
			e              Event
		)
		err = s.db.QueryRow(
			`SELECT id, creator_user_id, title, description, category, resolve_at, status, COALESCE(winner_outcome, ''), created_at
			 FROM events
			 WHERE id = $1`,
			eventID,
		).Scan(&eid, &creatorID, &e.Title, &e.Description, &e.Category, &e.ResolveAt, &e.Status, &e.WinnerOutcome, &e.CreatedAt)
		if err != nil {
			return Event{}, false
		}

		e.ID = formatIntID(eid)
		e.CreatorUserID = formatIntID(creatorID)
		return e, true
	}

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
	if s.db != nil {
		rows, err := s.db.Query(
			`SELECT
				m.id,
				m.event_id,
				m.status,
				COALESCE(m.moderator_id, 0),
				COALESCE(m.reason, ''),
				m.created_at,
				m.reviewed_at,
				e.creator_user_id,
				e.title,
				e.description,
				e.category,
				e.resolve_at,
				e.status,
				COALESCE(e.winner_outcome, ''),
				e.created_at
			 FROM moderation_tasks m
			 JOIN events e ON e.id = m.event_id
			 WHERE m.status = 'pending'
			 ORDER BY m.created_at ASC`,
		)
		if err != nil {
			return []ModerationQueueItem{}
		}
		defer rows.Close()

		items := make([]ModerationQueueItem, 0)
		for rows.Next() {
			var (
				taskID, eventID, moderatorID, creatorID int64
				reviewedAt                              sql.NullTime
				task                                    ModerationTask
				evt                                     Event
			)

			if err := rows.Scan(
				&taskID,
				&eventID,
				&task.Status,
				&moderatorID,
				&task.Reason,
				&task.CreatedAt,
				&reviewedAt,
				&creatorID,
				&evt.Title,
				&evt.Description,
				&evt.Category,
				&evt.ResolveAt,
				&evt.Status,
				&evt.WinnerOutcome,
				&evt.CreatedAt,
			); err != nil {
				continue
			}

			task.ID = formatIntID(taskID)
			task.EventID = formatIntID(eventID)
			if moderatorID > 0 {
				task.ModeratorID = formatIntID(moderatorID)
			}
			if reviewedAt.Valid {
				t := reviewedAt.Time
				task.ReviewedAt = &t
			}

			evt.ID = formatIntID(eventID)
			evt.CreatorUserID = formatIntID(creatorID)

			items = append(items, ModerationQueueItem{Task: task, Event: evt})
		}

		return items
	}

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

	if s.db != nil {
		eid, err := parseIntID(eventID)
		if err != nil {
			return Event{}, ErrEventNotFound
		}

		tx, err := s.db.Begin()
		if err != nil {
			return Event{}, err
		}
		defer tx.Rollback()

		var eventStatus string
		if err := tx.QueryRow(`SELECT status FROM events WHERE id = $1`, eid).Scan(&eventStatus); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return Event{}, ErrEventNotFound
			}
			return Event{}, err
		}

		var taskStatus string
		if err := tx.QueryRow(`SELECT status FROM moderation_tasks WHERE event_id = $1`, eid).Scan(&taskStatus); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return Event{}, ErrEventNotFound
			}
			return Event{}, err
		}

		if taskStatus != "pending" || eventStatus != "pending" {
			return Event{}, ErrModerationAlreadyHandled
		}

		modID, modErr := parseIntID(moderatorID)
		mod := sql.NullInt64{}
		if modErr == nil {
			mod = sql.NullInt64{Int64: modID, Valid: true}
		}

		if _, err := tx.Exec(
			`UPDATE moderation_tasks
			 SET status = $1, moderator_id = $2, reason = $3, reviewed_at = NOW()
			 WHERE event_id = $4`,
			targetStatus,
			mod,
			strings.TrimSpace(reason),
			eid,
		); err != nil {
			return Event{}, err
		}

		if _, err := tx.Exec(`UPDATE events SET status = $1 WHERE id = $2`, targetStatus, eid); err != nil {
			return Event{}, err
		}

		if err := tx.Commit(); err != nil {
			return Event{}, err
		}

		updated, ok := s.GetEventByID(eventID)
		if !ok {
			return Event{}, ErrEventNotFound
		}

		return updated, nil
	}

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

func (s *Service) SettleEvent(eventID, winnerOutcome string) (Event, error) {
	eventID = strings.TrimSpace(eventID)
	winnerOutcome = strings.ToLower(strings.TrimSpace(winnerOutcome))

	if winnerOutcome != "yes" && winnerOutcome != "no" {
		return Event{}, ErrInvalidSettlementInput
	}

	if s.db != nil {
		eid, err := parseIntID(eventID)
		if err != nil {
			return Event{}, ErrEventNotFound
		}

		var (
			creatorID int64
			event     Event
		)
		err = s.db.QueryRow(
			`UPDATE events
			 SET status = 'settled', winner_outcome = $2
			 WHERE id = $1 AND status = 'approved'
			 RETURNING id, creator_user_id, title, description, category, resolve_at, status, COALESCE(winner_outcome, ''), created_at`,
			eid,
			winnerOutcome,
		).Scan(&eid, &creatorID, &event.Title, &event.Description, &event.Category, &event.ResolveAt, &event.Status, &event.WinnerOutcome, &event.CreatedAt)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				if _, ok := s.GetEventByID(eventID); !ok {
					return Event{}, ErrEventNotFound
				}
				return Event{}, ErrEventNotSettlable
			}
			return Event{}, err
		}

		event.ID = formatIntID(eid)
		event.CreatorUserID = formatIntID(creatorID)
		return event, nil
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	e, ok := s.eventsByID[eventID]
	if !ok {
		return Event{}, ErrEventNotFound
	}

	if e.Status != "approved" {
		return Event{}, ErrEventNotSettlable
	}

	e.Status = "settled"
	e.WinnerOutcome = winnerOutcome

	return *e, nil
}

func parseIntID(raw string) (int64, error) {
	return strconv.ParseInt(strings.TrimSpace(raw), 10, 64)
}

func formatIntID(id int64) string {
	return strconv.FormatInt(id, 10)
}
