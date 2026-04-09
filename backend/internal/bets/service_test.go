package bets

import (
	"testing"
	"time"

	"bet/backend/internal/events"
	"bet/backend/internal/wallet"
)

func TestPlaceBetIdempotencyAndHold(t *testing.T) {
	eventsSvc := events.NewService()
	walletSvc := wallet.NewService(1000)
	betsSvc := NewService(eventsSvc, walletSvc)

	e := createApprovedEventForBetsTest(t, eventsSvc)

	b, created, err := betsSvc.PlaceBet("usr_1", e.ID, "yes", "idem-1", 100)
	if err != nil {
		t.Fatalf("unexpected error on first place bet: %v", err)
	}
	if !created {
		t.Fatal("expected created=true on first place bet")
	}
	if b.ID == "" {
		t.Fatal("expected non-empty bet id")
	}

	w, err := walletSvc.GetWallet("usr_1")
	if err != nil {
		t.Fatalf("unexpected wallet get error: %v", err)
	}
	if w.BalanceTokens != 900 {
		t.Fatalf("expected wallet balance 900 after first hold, got %v", w.BalanceTokens)
	}

	b2, created2, err := betsSvc.PlaceBet("usr_1", e.ID, "yes", "idem-1", 100)
	if err != nil {
		t.Fatalf("unexpected error on idempotent retry: %v", err)
	}
	if created2 {
		t.Fatal("expected created=false on idempotent retry")
	}
	if b2.ID != b.ID {
		t.Fatalf("expected same bet id %q on idempotent retry, got %q", b.ID, b2.ID)
	}

	w, err = walletSvc.GetWallet("usr_1")
	if err != nil {
		t.Fatalf("unexpected wallet get error after retry: %v", err)
	}
	if w.BalanceTokens != 900 {
		t.Fatalf("expected wallet balance unchanged 900 after idempotent retry, got %v", w.BalanceTokens)
	}
}

func TestPlaceBetInsufficientFunds(t *testing.T) {
	eventsSvc := events.NewService()
	walletSvc := wallet.NewService(50)
	betsSvc := NewService(eventsSvc, walletSvc)

	e := createApprovedEventForBetsTest(t, eventsSvc)

	_, _, err := betsSvc.PlaceBet("usr_2", e.ID, "no", "idem-2", 100)
	if err == nil {
		t.Fatal("expected insufficient funds error, got nil")
	}
	if err != wallet.ErrInsufficientFunds {
		t.Fatalf("expected wallet.ErrInsufficientFunds, got %v", err)
	}
}

func TestSettleEventBets(t *testing.T) {
	eventsSvc := events.NewService()
	walletSvc := wallet.NewService(1000)
	betsSvc := NewService(eventsSvc, walletSvc)

	e := createApprovedEventForBetsTest(t, eventsSvc)

	_, _, err := betsSvc.PlaceBet("usr_yes", e.ID, "yes", "idem-yes", 100)
	if err != nil {
		t.Fatalf("unexpected error placing yes bet: %v", err)
	}

	_, _, err = betsSvc.PlaceBet("usr_no", e.ID, "no", "idem-no", 100)
	if err != nil {
		t.Fatalf("unexpected error placing no bet: %v", err)
	}

	updated, err := betsSvc.SettleEventBets(e.ID, "yes")
	if err != nil {
		t.Fatalf("unexpected settle error: %v", err)
	}
	if len(updated) != 2 {
		t.Fatalf("expected 2 settled bets, got %d", len(updated))
	}

	yesWallet, err := walletSvc.GetWallet("usr_yes")
	if err != nil {
		t.Fatalf("unexpected yes wallet get error: %v", err)
	}
	if yesWallet.BalanceTokens != 1100 {
		t.Fatalf("expected winner balance 1100, got %v", yesWallet.BalanceTokens)
	}

	noWallet, err := walletSvc.GetWallet("usr_no")
	if err != nil {
		t.Fatalf("unexpected no wallet get error: %v", err)
	}
	if noWallet.BalanceTokens != 900 {
		t.Fatalf("expected loser balance 900, got %v", noWallet.BalanceTokens)
	}
}

func createApprovedEventForBetsTest(t *testing.T, svc *events.Service) events.Event {
	t.Helper()

	e, err := svc.CreateEvent(
		"creator",
		"Will this be approved?",
		"Event for bets service tests",
		"tests",
		time.Now().UTC().Add(24*time.Hour),
	)
	if err != nil {
		t.Fatalf("failed to create event: %v", err)
	}

	e, err = svc.ApproveEvent(e.ID, "moderator")
	if err != nil {
		t.Fatalf("failed to approve event: %v", err)
	}

	return e
}
