package wallet

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
	db             *sql.DB
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

func NewServiceWithDB(db *sql.DB, initialBalance float64) *Service {
	if db == nil {
		return NewService(initialBalance)
	}
	if initialBalance < 0 {
		initialBalance = 0
	}

	return &Service{db: db, initialBalance: initialBalance}
}

func (s *Service) GetWallet(userID string) (Wallet, error) {
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return Wallet{}, ErrInvalidWalletInput
	}

	if s.db != nil {
		uid, err := strconv.ParseInt(userID, 10, 64)
		if err != nil {
			return Wallet{}, ErrInvalidWalletInput
		}

		if err := s.ensureWalletInDB(uid); err != nil {
			return Wallet{}, err
		}

		var w Wallet
		err = s.db.QueryRow(
			`SELECT user_id, balance_tokens, updated_at FROM wallets WHERE user_id = $1`,
			uid,
		).Scan(&uid, &w.BalanceTokens, &w.UpdatedAt)
		if err != nil {
			return Wallet{}, err
		}

		w.UserID = strconv.FormatInt(uid, 10)
		return w, nil
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

	if s.db != nil {
		uid, err := strconv.ParseInt(userID, 10, 64)
		if err != nil {
			return nil, ErrInvalidWalletInput
		}

		if err := s.ensureWalletInDB(uid); err != nil {
			return nil, err
		}

		rows, err := s.db.Query(
			`SELECT wt.id, w.user_id, wt.type, wt.amount_tokens, COALESCE(wt.ref_type, ''), COALESCE(wt.ref_id, 0), wt.created_at
			 FROM wallet_transactions wt
			 JOIN wallets w ON w.id = wt.wallet_id
			 WHERE w.user_id = $1
			 ORDER BY wt.created_at DESC`,
			uid,
		)
		if err != nil {
			return nil, err
		}
		defer rows.Close()

		items := make([]Transaction, 0)
		for rows.Next() {
			var (
				txID    int64
				uidOut  int64
				refID   int64
				created time.Time
				tx      Transaction
			)
			if err := rows.Scan(&txID, &uidOut, &tx.Type, &tx.AmountTokens, &tx.RefType, &refID, &created); err != nil {
				continue
			}

			tx.ID = strconv.FormatInt(txID, 10)
			tx.UserID = strconv.FormatInt(uidOut, 10)
			if refID > 0 {
				tx.RefID = strconv.FormatInt(refID, 10)
			}
			tx.CreatedAt = created
			items = append(items, tx)
		}

		return items, nil
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

	if s.db != nil {
		return s.applyLedgerTx(userID, "hold", amount, refType, refID, true)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	w := s.ensureWalletLocked(userID)
	if w.BalanceTokens < amount {
		return Transaction{}, ErrInsufficientFunds
	}

	w.BalanceTokens -= amount
	w.UpdatedAt = time.Now().UTC()

	tx := s.appendTransactionLocked(userID, "hold", amount, refType, refID)
	return tx, nil
}

func (s *Service) SettlePayout(userID string, amount float64, refType, refID string) (Transaction, error) {
	userID = strings.TrimSpace(userID)
	refType = strings.TrimSpace(refType)
	refID = strings.TrimSpace(refID)

	if userID == "" || amount <= 0 {
		return Transaction{}, ErrInvalidWalletInput
	}

	if s.db != nil {
		return s.applyLedgerTx(userID, "settle", amount, refType, refID, false)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	w := s.ensureWalletLocked(userID)
	w.BalanceTokens += amount
	w.UpdatedAt = time.Now().UTC()

	tx := s.appendTransactionLocked(userID, "settle", amount, refType, refID)
	return tx, nil
}

func (s *Service) appendTransactionLocked(userID, txType string, amount float64, refType, refID string) Transaction {
	s.transactionSeq++
	tx := Transaction{
		ID:           fmt.Sprintf("wtx_%d", s.transactionSeq),
		UserID:       userID,
		Type:         txType,
		AmountTokens: amount,
		RefType:      refType,
		RefID:        refID,
		CreatedAt:    time.Now().UTC(),
	}

	s.txByUser[userID] = append(s.txByUser[userID], tx)
	return tx
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

func (s *Service) ensureWalletInDB(userID int64) error {
	_, err := s.db.Exec(
		`INSERT INTO wallets (user_id, balance_tokens, updated_at)
		 SELECT $1, $2, NOW()
		 WHERE EXISTS (SELECT 1 FROM users WHERE id = $1)
		 ON CONFLICT (user_id) DO NOTHING`,
		userID,
		s.initialBalance,
	)
	return err
}

func (s *Service) applyLedgerTx(userID, txType string, amount float64, refType, refID string, deduct bool) (Transaction, error) {
	uid, err := strconv.ParseInt(strings.TrimSpace(userID), 10, 64)
	if err != nil {
		return Transaction{}, ErrInvalidWalletInput
	}

	var parsedRefID sql.NullInt64
	if strings.TrimSpace(refID) != "" {
		if rid, err := strconv.ParseInt(strings.TrimSpace(refID), 10, 64); err == nil {
			parsedRefID = sql.NullInt64{Int64: rid, Valid: true}
		}
	}

	tx, err := s.db.Begin()
	if err != nil {
		return Transaction{}, err
	}
	defer tx.Rollback()

	if _, err := tx.Exec(
		`INSERT INTO wallets (user_id, balance_tokens, updated_at)
		 SELECT $1, $2, NOW()
		 WHERE EXISTS (SELECT 1 FROM users WHERE id = $1)
		 ON CONFLICT (user_id) DO NOTHING`,
		uid,
		s.initialBalance,
	); err != nil {
		return Transaction{}, err
	}

	var (
		walletID int64
		balance  float64
	)
	if err := tx.QueryRow(`SELECT id, balance_tokens FROM wallets WHERE user_id = $1 FOR UPDATE`, uid).Scan(&walletID, &balance); err != nil {
		return Transaction{}, err
	}

	newBalance := balance
	if deduct {
		if newBalance < amount {
			return Transaction{}, ErrInsufficientFunds
		}
		newBalance -= amount
	} else {
		newBalance += amount
	}

	if _, err := tx.Exec(`UPDATE wallets SET balance_tokens = $1, updated_at = NOW() WHERE id = $2`, newBalance, walletID); err != nil {
		return Transaction{}, err
	}

	var (
		createdTxID int64
		createdAt   time.Time
	)
	err = tx.QueryRow(
		`INSERT INTO wallet_transactions (wallet_id, type, amount_tokens, ref_type, ref_id, created_at)
		 VALUES ($1, $2, $3, NULLIF($4, ''), $5, NOW())
		 RETURNING id, created_at`,
		walletID,
		txType,
		amount,
		strings.TrimSpace(refType),
		parsedRefID,
	).Scan(&createdTxID, &createdAt)
	if err != nil {
		return Transaction{}, err
	}

	if err := tx.Commit(); err != nil {
		return Transaction{}, err
	}

	out := Transaction{
		ID:           strconv.FormatInt(createdTxID, 10),
		UserID:       strconv.FormatInt(uid, 10),
		Type:         txType,
		AmountTokens: amount,
		RefType:      strings.TrimSpace(refType),
		CreatedAt:    createdAt,
	}
	if parsedRefID.Valid {
		out.RefID = strconv.FormatInt(parsedRefID.Int64, 10)
	}

	return out, nil
}
