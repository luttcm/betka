package bets

import (
	"errors"
	"fmt"
	"sort"
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
