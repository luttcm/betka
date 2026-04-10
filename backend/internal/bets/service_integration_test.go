package bets

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"bet/backend/internal/auth"
	"bet/backend/internal/events"
	"bet/backend/internal/wallet"

	_ "github.com/lib/pq"
)

func TestPostgresPlaceBetIdempotencyAndWalletHold(t *testing.T) {
	db := openIntegrationDBOrSkip(t)
	prepareIntegrationSchema(t, db)

	authSvc := auth.NewServiceWithDB(db)
	eventsSvc := events.NewServiceWithDB(db)
	walletSvc := wallet.NewServiceWithDB(db, 1000)
	betsSvc := NewServiceWithDB(db, eventsSvc, walletSvc)

	creator := mustCreateVerifiedUser(t, authSvc, "creator.idem@example.com")
	bettor := mustCreateVerifiedUser(t, authSvc, "bettor.idem@example.com")

	event := mustCreateApprovedEvent(t, eventsSvc, creator.ID)

	first, created, err := betsSvc.PlaceBet(bettor.ID, event.ID, "yes", "idem-int-1", 100)
	if err != nil {
		t.Fatalf("unexpected error on first place bet: %v", err)
	}
	if !created {
		t.Fatal("expected created=true for first place bet")
	}

	second, created, err := betsSvc.PlaceBet(bettor.ID, event.ID, "yes", "idem-int-1", 100)
	if err != nil {
		t.Fatalf("unexpected error on idempotent retry: %v", err)
	}
	if created {
		t.Fatal("expected created=false for idempotent retry")
	}
	if second.ID != first.ID {
		t.Fatalf("expected same bet id on idempotent retry, first=%s second=%s", first.ID, second.ID)
	}

	walletState, err := walletSvc.GetWallet(bettor.ID)
	if err != nil {
		t.Fatalf("unexpected wallet get error: %v", err)
	}
	if walletState.BalanceTokens != 900 {
		t.Fatalf("expected balance=900 after one hold, got=%v", walletState.BalanceTokens)
	}

	uid, _ := strconv.ParseInt(bettor.ID, 10, 64)
	var holdCount int
	if err := db.QueryRow(
		`SELECT COUNT(1)
		 FROM wallet_transactions wt
		 JOIN wallets w ON w.id = wt.wallet_id
		 WHERE w.user_id = $1 AND wt.type = 'hold'`,
		uid,
	).Scan(&holdCount); err != nil {
		t.Fatalf("failed to count hold transactions: %v", err)
	}
	if holdCount != 1 {
		t.Fatalf("expected exactly one hold transaction, got=%d", holdCount)
	}
}

func TestPostgresSettleEventAndBetsIsAtomicAndNoDoublePayout(t *testing.T) {
	db := openIntegrationDBOrSkip(t)
	prepareIntegrationSchema(t, db)

	authSvc := auth.NewServiceWithDB(db)
	eventsSvc := events.NewServiceWithDB(db)
	walletSvc := wallet.NewServiceWithDB(db, 1000)
	betsSvc := NewServiceWithDB(db, eventsSvc, walletSvc)

	creator := mustCreateVerifiedUser(t, authSvc, "creator.settle@example.com")
	yesBettor := mustCreateVerifiedUser(t, authSvc, "yes.settle@example.com")
	noBettor := mustCreateVerifiedUser(t, authSvc, "no.settle@example.com")

	event := mustCreateApprovedEvent(t, eventsSvc, creator.ID)

	if _, _, err := betsSvc.PlaceBet(yesBettor.ID, event.ID, "yes", "idem-settle-yes", 100); err != nil {
		t.Fatalf("failed to place yes bet: %v", err)
	}
	if _, _, err := betsSvc.PlaceBet(noBettor.ID, event.ID, "no", "idem-settle-no", 100); err != nil {
		t.Fatalf("failed to place no bet: %v", err)
	}

	if _, err := eventsSvc.RequestSettlement(event.ID, creator.ID, "https://example.com/evidence", "", ""); err != nil {
		t.Fatalf("failed to request settlement: %v", err)
	}

	settledEvent, settledBets, err := betsSvc.SettleEventAndBets(event.ID, "yes")
	if err != nil {
		t.Fatalf("failed to settle event and bets: %v", err)
	}
	if settledEvent.Status != "settled" {
		t.Fatalf("expected settled event status, got=%s", settledEvent.Status)
	}
	if len(settledBets) != 2 {
		t.Fatalf("expected 2 settled bets, got=%d", len(settledBets))
	}

	yesWallet, err := walletSvc.GetWallet(yesBettor.ID)
	if err != nil {
		t.Fatalf("failed to get yes wallet: %v", err)
	}
	if yesWallet.BalanceTokens != 1090 {
		t.Fatalf("expected winner balance=1090, got=%v", yesWallet.BalanceTokens)
	}

	noWallet, err := walletSvc.GetWallet(noBettor.ID)
	if err != nil {
		t.Fatalf("failed to get no wallet: %v", err)
	}
	if noWallet.BalanceTokens != 900 {
		t.Fatalf("expected loser balance=900, got=%v", noWallet.BalanceTokens)
	}

	_, _, err = betsSvc.SettleEventAndBets(event.ID, "yes")
	if !errors.Is(err, events.ErrEventNotSettlable) {
		t.Fatalf("expected ErrEventNotSettlable on repeat settlement, got=%v", err)
	}

	winnerUID, _ := strconv.ParseInt(yesBettor.ID, 10, 64)
	var settleCount int
	if err := db.QueryRow(
		`SELECT COUNT(1)
		 FROM wallet_transactions wt
		 JOIN wallets w ON w.id = wt.wallet_id
		 WHERE w.user_id = $1 AND wt.type = 'settle'`,
		winnerUID,
	).Scan(&settleCount); err != nil {
		t.Fatalf("failed to count settle transactions: %v", err)
	}
	if settleCount != 1 {
		t.Fatalf("expected exactly one settle transaction for winner, got=%d", settleCount)
	}
}

func openIntegrationDBOrSkip(t *testing.T) *sql.DB {
	t.Helper()

	dsn := strings.TrimSpace(os.Getenv("INTEGRATION_DATABASE_URL"))
	if dsn == "" {
		t.Skip("INTEGRATION_DATABASE_URL is empty; skipping postgres integration tests")
	}

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatalf("failed to open postgres integration db: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	if err := db.Ping(); err != nil {
		t.Fatalf("failed to ping postgres integration db: %v", err)
	}

	return db
}

func prepareIntegrationSchema(t *testing.T, db *sql.DB) {
	t.Helper()

	if _, err := db.Exec(`
		DROP TABLE IF EXISTS audit_logs;
		DROP TABLE IF EXISTS wallet_transactions;
		DROP TABLE IF EXISTS moderation_tasks;
		DROP TABLE IF EXISTS bets;
		DROP TABLE IF EXISTS odds_snapshots;
		DROP TABLE IF EXISTS event_outcomes;
		DROP TABLE IF EXISTS events;
		DROP TABLE IF EXISTS wallets;
		DROP TABLE IF EXISTS users;
	`); err != nil {
		t.Fatalf("failed to drop existing schema: %v", err)
	}

	for _, fileName := range []string{
		"00001_init_schema.sql",
		"00002_add_auth_and_events_columns.sql",
		"00003_add_settlement_request_and_dynamic_odds.sql",
	} {
		path := filepath.Join("..", "..", "migrations", fileName)
		raw, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("failed to read migration file %s: %v", fileName, err)
		}

		upPart := string(raw)
		if idx := strings.Index(upPart, "-- +goose Down"); idx >= 0 {
			upPart = upPart[:idx]
		}

		if _, err := db.Exec(upPart); err != nil {
			t.Fatalf("failed to apply migration %s: %v", fileName, err)
		}
	}
}

func mustCreateVerifiedUser(t *testing.T, authSvc *auth.Service, email string) auth.User {
	t.Helper()

	password := "strong-password"
	u, err := authSvc.Register(email, password)
	if err != nil {
		t.Fatalf("failed to register user %s: %v", email, err)
	}

	verified, err := authSvc.VerifyEmail(u.VerifyToken)
	if err != nil {
		t.Fatalf("failed to verify user %s: %v", email, err)
	}

	loggedIn, err := authSvc.Login(email, password)
	if err != nil {
		t.Fatalf("failed to login user %s: %v", email, err)
	}

	if loggedIn.ID != verified.ID {
		t.Fatalf("unexpected login user id mismatch for %s: verified=%s login=%s", email, verified.ID, loggedIn.ID)
	}

	return loggedIn
}

func mustCreateApprovedEvent(t *testing.T, eventsSvc *events.Service, creatorID string) events.Event {
	t.Helper()

	e, err := eventsSvc.CreateEvent(
		creatorID,
		fmt.Sprintf("Integration event %d", time.Now().UnixNano()),
		"integration test event",
		"tests",
		time.Now().UTC().Add(24*time.Hour),
	)
	if err != nil {
		t.Fatalf("failed to create event: %v", err)
	}

	e, err = eventsSvc.ApproveEvent(e.ID, creatorID)
	if err != nil {
		t.Fatalf("failed to approve event: %v", err)
	}

	return e
}
