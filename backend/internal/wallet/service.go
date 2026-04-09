package wallet

import (
	"errors"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"
)

var (
	ErrInvalidWalletInput = errors.New("invalid wallet input")
	ErrInsufficientFunds  = errors.New("insufficient funds")
)

type Wallet struct {
	UserID        string    `json:"user_id"`
	BalanceTokens float64   `json:"balance_tokens"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type Transaction struct {
	ID           string    `json:"id"`
	UserID       string    `json:"user_id"`
	Type         string    `json:"type"`
	AmountTokens float64   `json:"amount_tokens"`
	RefType      string    `json:"ref_type,omitempty"`
	RefID        string    `json:"ref_id,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
}

type Service struct {
	mu             sync.RWMutex
	walletsByUser  map[string]*Wallet
	txByUser       map[string][]Transaction
	transactionSeq int64
	initialBalance float64
}

func NewService(initialBalance float64) *Service {
	if initialBalance < 0 {
		initialBalance = 0
	}

	return &Service{
		walletsByUser:  make(map[string]*Wallet),
		txByUser:       make(map[string][]Transaction),
		initialBalance: initialBalance,
	}
}

func (s *Service) GetWallet(userID string) (Wallet, error) {
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return Wallet{}, ErrInvalidWalletInput
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	w := s.ensureWalletLocked(userID)
	return *w, nil
}

func (s *Service) ListTransactions(userID string) ([]Transaction, error) {
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return nil, ErrInvalidWalletInput
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.ensureWalletLocked(userID)
	items := append([]Transaction(nil), s.txByUser[userID]...)
	sort.Slice(items, func(i, j int) bool {
		return items[i].CreatedAt.After(items[j].CreatedAt)
	})

	return items, nil
}

func (s *Service) Hold(userID string, amount float64, refType, refID string) (Transaction, error) {
	userID = strings.TrimSpace(userID)
	refType = strings.TrimSpace(refType)
	refID = strings.TrimSpace(refID)

	if userID == "" || amount <= 0 {
		return Transaction{}, ErrInvalidWalletInput
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	w := s.ensureWalletLocked(userID)
	if w.BalanceTokens < amount {
		return Transaction{}, ErrInsufficientFunds
	}

	w.BalanceTokens -= amount
	w.UpdatedAt = time.Now().UTC()

	s.transactionSeq++
	tx := Transaction{
		ID:           fmt.Sprintf("wtx_%d", s.transactionSeq),
		UserID:       userID,
		Type:         "hold",
		AmountTokens: amount,
		RefType:      refType,
		RefID:        refID,
		CreatedAt:    time.Now().UTC(),
	}

	s.txByUser[userID] = append(s.txByUser[userID], tx)

	return tx, nil
}

func (s *Service) ensureWalletLocked(userID string) *Wallet {
	if w, ok := s.walletsByUser[userID]; ok {
		return w
	}

	w := &Wallet{
		UserID:        userID,
		BalanceTokens: s.initialBalance,
		UpdatedAt:     time.Now().UTC(),
	}
	s.walletsByUser[userID] = w
	return w
}
