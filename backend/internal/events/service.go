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
	ErrInvalidEventInput          = errors.New("invalid event input")
	ErrEventNotFound              = errors.New("event not found")
	ErrModerationAlreadyHandled   = errors.New("moderation task already handled")
	ErrInvalidModerationReason    = errors.New("invalid moderation reason")
	ErrInvalidSettlementInput     = errors.New("invalid settlement input")
	ErrEventNotSettlable          = errors.New("event is not settlable")
	ErrForbiddenSettlementRequest = errors.New("forbidden settlement request")
	ErrInvalidSettlementEvidence  = errors.New("invalid settlement evidence")
)

type Event struct {
	ID                         string     `json:"id"`
	CreatorUserID              string     `json:"creator_user_id"`
	Title                      string     `json:"title"`
	Description                string     `json:"description"`
	Category                   string     `json:"category"`
	ResolveAt                  time.Time  `json:"resolve_at"`
	Status                     string     `json:"status"`
	WinnerOutcome              string     `json:"winner_outcome,omitempty"`
	SettlementRequestedBy      string     `json:"settlement_requested_by,omitempty"`
	SettlementRequestedAt      *time.Time `json:"settlement_requested_at,omitempty"`
	SettlementEvidenceURL      string     `json:"settlement_evidence_url,omitempty"`
	SettlementEvidenceFileName string     `json:"settlement_evidence_file_name,omitempty"`
	SettlementEvidenceFileData string     `json:"settlement_evidence_file_data,omitempty"`
	CreatedAt                  time.Time  `json:"created_at"`
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
			`SELECT id, creator_user_id, title, description, category, resolve_at, status,
			        COALESCE(winner_outcome, ''),
			        COALESCE(settlement_requested_by, 0), settlement_requested_at,
			        COALESCE(settlement_evidence_url, ''), COALESCE(settlement_evidence_file_name, ''), COALESCE(settlement_evidence_file_data, ''),
			        created_at
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
				eid, creatorID, settlementRequestedBy int64
				settlementRequestedAt                 sql.NullTime
				e                                     Event
			)
			if err := rows.Scan(
				&eid,
				&creatorID,
				&e.Title,
				&e.Description,
				&e.Category,
				&e.ResolveAt,
				&e.Status,
				&e.WinnerOutcome,
				&settlementRequestedBy,
				&settlementRequestedAt,
				&e.SettlementEvidenceURL,
				&e.SettlementEvidenceFileName,
				&e.SettlementEvidenceFileData,
				&e.CreatedAt,
			); err != nil {
				continue
			}

			e.ID = formatIntID(eid)
			e.CreatorUserID = formatIntID(creatorID)
			if settlementRequestedBy > 0 {
				e.SettlementRequestedBy = formatIntID(settlementRequestedBy)
			}
			if settlementRequestedAt.Valid {
				t := settlementRequestedAt.Time
				e.SettlementRequestedAt = &t
			}
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
			eid, creatorID, settlementRequestedBy int64
			settlementRequestedAt                 sql.NullTime
			e                                     Event
		)
		err = s.db.QueryRow(
			`SELECT id, creator_user_id, title, description, category, resolve_at, status,
			        COALESCE(winner_outcome, ''),
			        COALESCE(settlement_requested_by, 0), settlement_requested_at,
			        COALESCE(settlement_evidence_url, ''), COALESCE(settlement_evidence_file_name, ''), COALESCE(settlement_evidence_file_data, ''),
			        created_at
			 FROM events
			 WHERE id = $1`,
			eventID,
		).Scan(
			&eid,
			&creatorID,
			&e.Title,
			&e.Description,
			&e.Category,
			&e.ResolveAt,
			&e.Status,
			&e.WinnerOutcome,
			&settlementRequestedBy,
			&settlementRequestedAt,
			&e.SettlementEvidenceURL,
			&e.SettlementEvidenceFileName,
			&e.SettlementEvidenceFileData,
			&e.CreatedAt,
		)
		if err != nil {
			return Event{}, false
		}

		e.ID = formatIntID(eid)
		e.CreatorUserID = formatIntID(creatorID)
		if settlementRequestedBy > 0 {
			e.SettlementRequestedBy = formatIntID(settlementRequestedBy)
		}
		if settlementRequestedAt.Valid {
			t := settlementRequestedAt.Time
			e.SettlementRequestedAt = &t
		}
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
				COALESCE(e.settlement_requested_by, 0),
				e.settlement_requested_at,
				COALESCE(e.settlement_evidence_url, ''),
				COALESCE(e.settlement_evidence_file_name, ''),
				COALESCE(e.settlement_evidence_file_data, ''),
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
				taskID, eventID, moderatorID, creatorID, settlementRequestedBy int64
				reviewedAt, settlementRequestedAt                              sql.NullTime
				task                                                           ModerationTask
				evt                                                            Event
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
				&settlementRequestedBy,
				&settlementRequestedAt,
				&evt.SettlementEvidenceURL,
				&evt.SettlementEvidenceFileName,
				&evt.SettlementEvidenceFileData,
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
			if settlementRequestedBy > 0 {
				evt.SettlementRequestedBy = formatIntID(settlementRequestedBy)
			}
			if settlementRequestedAt.Valid {
				t := settlementRequestedAt.Time
				evt.SettlementRequestedAt = &t
			}

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
			creatorID, settlementRequestedBy int64
			settlementRequestedAt            sql.NullTime
			event                            Event
		)
		err = s.db.QueryRow(
			`UPDATE events
			 SET status = 'settled', winner_outcome = $2
			 WHERE id = $1 AND status = 'settlement_requested'
			 RETURNING id, creator_user_id, title, description, category, resolve_at, status,
			        COALESCE(winner_outcome, ''),
			        COALESCE(settlement_requested_by, 0), settlement_requested_at,
			        COALESCE(settlement_evidence_url, ''), COALESCE(settlement_evidence_file_name, ''), COALESCE(settlement_evidence_file_data, ''),
			        created_at`,
			eid,
			winnerOutcome,
		).Scan(
			&eid,
			&creatorID,
			&event.Title,
			&event.Description,
			&event.Category,
			&event.ResolveAt,
			&event.Status,
			&event.WinnerOutcome,
			&settlementRequestedBy,
			&settlementRequestedAt,
			&event.SettlementEvidenceURL,
			&event.SettlementEvidenceFileName,
			&event.SettlementEvidenceFileData,
			&event.CreatedAt,
		)
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
		if settlementRequestedBy > 0 {
			event.SettlementRequestedBy = formatIntID(settlementRequestedBy)
		}
		if settlementRequestedAt.Valid {
			t := settlementRequestedAt.Time
			event.SettlementRequestedAt = &t
		}
		return event, nil
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	e, ok := s.eventsByID[eventID]
	if !ok {
		return Event{}, ErrEventNotFound
	}

	if e.Status != "settlement_requested" {
		return Event{}, ErrEventNotSettlable
	}

	e.Status = "settled"
	e.WinnerOutcome = winnerOutcome

	return *e, nil
}

func (s *Service) RequestSettlement(eventID, requesterUserID, evidenceURL, evidenceFileName, evidenceFileData string) (Event, error) {
	eventID = strings.TrimSpace(eventID)
	requesterUserID = strings.TrimSpace(requesterUserID)
	evidenceURL = strings.TrimSpace(evidenceURL)
	evidenceFileName = strings.TrimSpace(evidenceFileName)
	evidenceFileData = strings.TrimSpace(evidenceFileData)

	if eventID == "" || requesterUserID == "" {
		return Event{}, ErrInvalidSettlementInput
	}

	if evidenceURL == "" && (evidenceFileName == "" || evidenceFileData == "") {
		return Event{}, ErrInvalidSettlementEvidence
	}

	if s.db != nil {
		eid, err := parseIntID(eventID)
		if err != nil {
			return Event{}, ErrEventNotFound
		}

		rid, err := parseIntID(requesterUserID)
		if err != nil {
			return Event{}, ErrInvalidSettlementInput
		}

		var (
			creatorID int64
			status    string
		)
		if err := s.db.QueryRow(`SELECT creator_user_id, status FROM events WHERE id = $1`, eid).Scan(&creatorID, &status); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return Event{}, ErrEventNotFound
			}
			return Event{}, err
		}

		if creatorID != rid {
			return Event{}, ErrForbiddenSettlementRequest
		}

		if status != "approved" {
			return Event{}, ErrEventNotSettlable
		}

		if _, err := s.db.Exec(
			`UPDATE events
			 SET status = 'settlement_requested',
			     settlement_requested_by = $2,
			     settlement_requested_at = NOW(),
			     settlement_evidence_url = NULLIF($3, ''),
			     settlement_evidence_file_name = NULLIF($4, ''),
			     settlement_evidence_file_data = NULLIF($5, '')
			 WHERE id = $1`,
			eid,
			rid,
			evidenceURL,
			evidenceFileName,
			evidenceFileData,
		); err != nil {
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

	if e.CreatorUserID != requesterUserID {
		return Event{}, ErrForbiddenSettlementRequest
	}

	if e.Status != "approved" {
		return Event{}, ErrEventNotSettlable
	}

	now := time.Now().UTC()
	e.Status = "settlement_requested"
	e.SettlementRequestedBy = requesterUserID
	e.SettlementRequestedAt = &now
	e.SettlementEvidenceURL = evidenceURL
	e.SettlementEvidenceFileName = evidenceFileName
	e.SettlementEvidenceFileData = evidenceFileData

	return *e, nil
}

func (s *Service) ListSettlementRequests() []Event {
	if s.db != nil {
		rows, err := s.db.Query(
			`SELECT id, creator_user_id, title, description, category, resolve_at, status,
			        COALESCE(winner_outcome, ''),
			        COALESCE(settlement_requested_by, 0), settlement_requested_at,
			        COALESCE(settlement_evidence_url, ''), COALESCE(settlement_evidence_file_name, ''), COALESCE(settlement_evidence_file_data, ''),
			        created_at
			 FROM events
			 WHERE status = 'settlement_requested'
			 ORDER BY settlement_requested_at DESC NULLS LAST`,
		)
		if err != nil {
			return []Event{}
		}
		defer rows.Close()

		items := make([]Event, 0)
		for rows.Next() {
			var (
				eid, creatorID, settlementRequestedBy int64
				settlementRequestedAt                 sql.NullTime
				e                                     Event
			)
			if err := rows.Scan(
				&eid,
				&creatorID,
				&e.Title,
				&e.Description,
				&e.Category,
				&e.ResolveAt,
				&e.Status,
				&e.WinnerOutcome,
				&settlementRequestedBy,
				&settlementRequestedAt,
				&e.SettlementEvidenceURL,
				&e.SettlementEvidenceFileName,
				&e.SettlementEvidenceFileData,
				&e.CreatedAt,
			); err != nil {
				continue
			}

			e.ID = formatIntID(eid)
			e.CreatorUserID = formatIntID(creatorID)
			if settlementRequestedBy > 0 {
				e.SettlementRequestedBy = formatIntID(settlementRequestedBy)
			}
			if settlementRequestedAt.Valid {
				t := settlementRequestedAt.Time
				e.SettlementRequestedAt = &t
			}

			items = append(items, e)
		}

		return items
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	items := make([]Event, 0)
	for _, e := range s.eventsByID {
		if e.Status == "settlement_requested" {
			items = append(items, *e)
		}
	}

	sort.Slice(items, func(i, j int) bool {
		iAt := items[i].SettlementRequestedAt
		jAt := items[j].SettlementRequestedAt
		if iAt == nil {
			return false
		}
		if jAt == nil {
			return true
		}
		return iAt.After(*jAt)
	})

	return items
}

func parseIntID(raw string) (int64, error) {
	return strconv.ParseInt(strings.TrimSpace(raw), 10, 64)
}

func formatIntID(id int64) string {
	return strconv.FormatInt(id, 10)
}
