package bets

import (
	"database/sql"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"bet/backend/internal/events"
	"bet/backend/internal/wallet"
)

var (
	ErrInvalidBetInput       = errors.New("invalid bet input")
	ErrMissingIdempotencyKey = errors.New("missing idempotency key")
	ErrEventUnavailable      = errors.New("event unavailable for betting")
	ErrInvalidSettlement     = errors.New("invalid settlement")
)

type Bet struct {
	ID              string     `json:"id"`
	UserID          string     `json:"user_id"`
	EventID         string     `json:"event_id"`
	OutcomeCode     string     `json:"outcome_code"`
	Stake           float64    `json:"stake"`
	OddsAtBet       float64    `json:"odds_at_bet"`
	PotentialPayout float64    `json:"potential_payout"`
	Status          string     `json:"status"`
	IdempotencyKey  string     `json:"idempotency_key"`
	PlacedAt        time.Time  `json:"placed_at"`
	SettledAt       *time.Time `json:"settled_at,omitempty"`
}

type Service struct {
	db            *sql.DB
	mu            sync.RWMutex
	eventsService *events.Service
	walletService *wallet.Service
	betsByID      map[string]*Bet
	betsByUser    map[string][]*Bet
	betsByUserKey map[string]map[string]*Bet
	betSeq        int64
}

func NewService(eventsService *events.Service, walletService *wallet.Service) *Service {
	return &Service{
		eventsService: eventsService,
		walletService: walletService,
		betsByID:      make(map[string]*Bet),
		betsByUser:    make(map[string][]*Bet),
		betsByUserKey: make(map[string]map[string]*Bet),
	}
}

func NewServiceWithDB(db *sql.DB, eventsService *events.Service, walletService *wallet.Service) *Service {
	if db == nil {
		return NewService(eventsService, walletService)
	}

	return &Service{
		db:            db,
		eventsService: eventsService,
		walletService: walletService,
	}
}

func (s *Service) PlaceBet(userID, eventID, outcomeCode, idempotencyKey string, stake float64) (Bet, bool, error) {
	userID = strings.TrimSpace(userID)
	eventID = strings.TrimSpace(eventID)
	outcomeCode = strings.ToLower(strings.TrimSpace(outcomeCode))
	idempotencyKey = strings.TrimSpace(idempotencyKey)

	if userID == "" || eventID == "" || stake <= 0 {
		return Bet{}, false, ErrInvalidBetInput
	}

	if idempotencyKey == "" {
		return Bet{}, false, ErrMissingIdempotencyKey
	}

	if outcomeCode != "yes" && outcomeCode != "no" {
		return Bet{}, false, ErrInvalidBetInput
	}

	if s.db != nil {
		return s.placeBetDB(userID, eventID, outcomeCode, idempotencyKey, stake)
	}

	e, ok := s.eventsService.GetEventByID(eventID)
	if !ok || e.Status != "approved" || !e.ResolveAt.After(time.Now().UTC()) {
		return Bet{}, false, ErrEventUnavailable
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if byKey, ok := s.betsByUserKey[userID]; ok {
		if existing, ok := byKey[idempotencyKey]; ok {
			return *existing, false, nil
		}
	}

	s.betSeq++
	betID := fmt.Sprintf("bet_%d", s.betSeq)

	if _, err := s.walletService.Hold(userID, stake, "bet", betID); err != nil {
		return Bet{}, false, err
	}

	odds := 2.0
	b := &Bet{
		ID:              betID,
		UserID:          userID,
		EventID:         eventID,
		OutcomeCode:     outcomeCode,
		Stake:           stake,
		OddsAtBet:       odds,
		PotentialPayout: stake * odds,
		Status:          "open",
		IdempotencyKey:  idempotencyKey,
		PlacedAt:        time.Now().UTC(),
	}

	s.betsByID[betID] = b
	s.betsByUser[userID] = append(s.betsByUser[userID], b)
	if _, ok := s.betsByUserKey[userID]; !ok {
		s.betsByUserKey[userID] = make(map[string]*Bet)
	}
	s.betsByUserKey[userID][idempotencyKey] = b

	return *b, true, nil
}

func (s *Service) ListMyBets(userID string) ([]Bet, error) {
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return nil, ErrInvalidBetInput
	}

	if s.db != nil {
		uid, err := strconv.ParseInt(userID, 10, 64)
		if err != nil {
			return nil, ErrInvalidBetInput
		}

		rows, err := s.db.Query(
			`SELECT b.id, b.user_id, b.event_id, eo.code, b.stake, b.odds_at_bet, b.potential_payout, b.status, b.idempotency_key, b.placed_at, b.settled_at
			 FROM bets b
			 JOIN event_outcomes eo ON eo.id = b.outcome_id
			 WHERE b.user_id = $1
			 ORDER BY b.placed_at DESC`,
			uid,
		)
		if err != nil {
			return nil, err
		}
		defer rows.Close()

		items := make([]Bet, 0)
		for rows.Next() {
			var (
				bid, uidOut, eid int64
				settledAt        sql.NullTime
				b                Bet
			)

			if err := rows.Scan(
				&bid,
				&uidOut,
				&eid,
				&b.OutcomeCode,
				&b.Stake,
				&b.OddsAtBet,
				&b.PotentialPayout,
				&b.Status,
				&b.IdempotencyKey,
				&b.PlacedAt,
				&settledAt,
			); err != nil {
				continue
			}

			b.ID = strconv.FormatInt(bid, 10)
			b.UserID = strconv.FormatInt(uidOut, 10)
			b.EventID = strconv.FormatInt(eid, 10)
			if settledAt.Valid {
				t := settledAt.Time
				b.SettledAt = &t
			}
			items = append(items, b)
		}

		return items, nil
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	bets := s.betsByUser[userID]
	items := make([]Bet, 0, len(bets))
	for _, b := range bets {
		items = append(items, *b)
	}

	sort.Slice(items, func(i, j int) bool {
		return items[i].PlacedAt.After(items[j].PlacedAt)
	})

	return items, nil
}

func (s *Service) SettleEventBets(eventID, winnerOutcome string) ([]Bet, error) {
	eventID = strings.TrimSpace(eventID)
	winnerOutcome = strings.ToLower(strings.TrimSpace(winnerOutcome))

	if eventID == "" || (winnerOutcome != "yes" && winnerOutcome != "no") {
		return nil, ErrInvalidSettlement
	}

	if s.db != nil {
		return s.settleEventBetsDB(eventID, winnerOutcome)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	updated := make([]Bet, 0)
	for _, b := range s.betsByID {
		if b.EventID != eventID || b.Status != "open" {
			continue
		}

		if b.OutcomeCode == winnerOutcome {
			b.Status = "won"
			if _, err := s.walletService.SettlePayout(b.UserID, b.PotentialPayout, "bet_settlement", b.ID); err != nil {
				return nil, err
			}
		} else {
			b.Status = "lost"
		}

		now := time.Now().UTC()
		b.SettledAt = &now
		updated = append(updated, *b)
	}

	sort.Slice(updated, func(i, j int) bool {
		return updated[i].PlacedAt.Before(updated[j].PlacedAt)
	})

	return updated, nil
}

func (s *Service) placeBetDB(userID, eventID, outcomeCode, idempotencyKey string, stake float64) (Bet, bool, error) {
	e, ok := s.eventsService.GetEventByID(eventID)
	if !ok || e.Status != "approved" || !e.ResolveAt.After(time.Now().UTC()) {
		return Bet{}, false, ErrEventUnavailable
	}

	uid, err := strconv.ParseInt(userID, 10, 64)
	if err != nil {
		return Bet{}, false, ErrInvalidBetInput
	}
	eid, err := strconv.ParseInt(eventID, 10, 64)
	if err != nil {
		return Bet{}, false, ErrInvalidBetInput
	}

	var existingCount int
	if err := s.db.QueryRow(`SELECT COUNT(1) FROM bets WHERE user_id = $1 AND idempotency_key = $2`, uid, idempotencyKey).Scan(&existingCount); err != nil {
		return Bet{}, false, err
	}
	if existingCount > 0 {
		items, err := s.ListMyBets(userID)
		if err != nil {
			return Bet{}, false, err
		}
		for _, it := range items {
			if it.IdempotencyKey == idempotencyKey {
				return it, false, nil
			}
		}
	}

	if _, err := s.walletService.Hold(userID, stake, "bet", ""); err != nil {
		return Bet{}, false, err
	}

	odds := 2.0
	var (
		outcomeID int64
		betID     int64
		placedAt  time.Time
	)
	if err := s.db.QueryRow(`SELECT id FROM event_outcomes WHERE event_id = $1 AND code = $2`, eid, outcomeCode).Scan(&outcomeID); err != nil {
		return Bet{}, false, ErrEventUnavailable
	}

	err = s.db.QueryRow(
		`INSERT INTO bets (user_id, event_id, outcome_id, stake, odds_at_bet, potential_payout, status, idempotency_key, placed_at)
		 VALUES ($1, $2, $3, $4, $5, $6, 'open', $7, NOW())
		 RETURNING id, placed_at`,
		uid,
		eid,
		outcomeID,
		stake,
		odds,
		stake*odds,
		idempotencyKey,
	).Scan(&betID, &placedAt)
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "unique") {
			items, listErr := s.ListMyBets(userID)
			if listErr != nil {
				return Bet{}, false, err
			}
			for _, it := range items {
				if it.IdempotencyKey == idempotencyKey {
					return it, false, nil
				}
			}
		}
		return Bet{}, false, err
	}

	return Bet{
		ID:              strconv.FormatInt(betID, 10),
		UserID:          strconv.FormatInt(uid, 10),
		EventID:         strconv.FormatInt(eid, 10),
		OutcomeCode:     outcomeCode,
		Stake:           stake,
		OddsAtBet:       odds,
		PotentialPayout: stake * odds,
		Status:          "open",
		IdempotencyKey:  idempotencyKey,
		PlacedAt:        placedAt,
	}, true, nil
}

func (s *Service) settleEventBetsDB(eventID, winnerOutcome string) ([]Bet, error) {
	eid, err := strconv.ParseInt(strings.TrimSpace(eventID), 10, 64)
	if err != nil {
		return nil, ErrInvalidSettlement
	}

	rows, err := s.db.Query(
		`SELECT b.id, b.user_id, b.event_id, eo.code, b.stake, b.odds_at_bet, b.potential_payout, b.status, b.idempotency_key, b.placed_at, b.settled_at
		 FROM bets b
		 JOIN event_outcomes eo ON eo.id = b.outcome_id
		 WHERE b.event_id = $1 AND b.status = 'open'
		 ORDER BY b.placed_at ASC`,
		eid,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	updated := make([]Bet, 0)
	for rows.Next() {
		var (
			bid, uid, eventIDValue int64
			settledAt              sql.NullTime
			b                      Bet
		)

		if err := rows.Scan(
			&bid,
			&uid,
			&eventIDValue,
			&b.OutcomeCode,
			&b.Stake,
			&b.OddsAtBet,
			&b.PotentialPayout,
			&b.Status,
			&b.IdempotencyKey,
			&b.PlacedAt,
			&settledAt,
		); err != nil {
			continue
		}

		status := "lost"
		if b.OutcomeCode == winnerOutcome {
			status = "won"
			if _, err := s.walletService.SettlePayout(strconv.FormatInt(uid, 10), b.PotentialPayout, "bet_settlement", strconv.FormatInt(bid, 10)); err != nil {
				return nil, err
			}
		}

		if _, err := s.db.Exec(`UPDATE bets SET status = $1, settled_at = NOW() WHERE id = $2`, status, bid); err != nil {
			return nil, err
		}

		now := time.Now().UTC()
		b.ID = strconv.FormatInt(bid, 10)
		b.UserID = strconv.FormatInt(uid, 10)
		b.EventID = strconv.FormatInt(eventIDValue, 10)
		b.Status = status
		b.SettledAt = &now
		updated = append(updated, b)
	}

	return updated, nil
}
