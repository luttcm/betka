package wallet

import "testing"

func TestHoldAndSettlePayoutFlow(t *testing.T) {
	svc := NewService(1000)

	w, err := svc.GetWallet("usr_1")
	if err != nil {
		t.Fatalf("unexpected error on get wallet: %v", err)
	}
	if w.BalanceTokens != 1000 {
		t.Fatalf("expected initial balance 1000, got %v", w.BalanceTokens)
	}

	holdTx, err := svc.Hold("usr_1", 100, "bet", "bet_1")
	if err != nil {
		t.Fatalf("unexpected error on hold: %v", err)
	}
	if holdTx.Type != "hold" {
		t.Fatalf("expected hold transaction type, got %q", holdTx.Type)
	}

	w, err = svc.GetWallet("usr_1")
	if err != nil {
		t.Fatalf("unexpected error on get wallet after hold: %v", err)
	}
	if w.BalanceTokens != 900 {
		t.Fatalf("expected balance 900 after hold, got %v", w.BalanceTokens)
	}

	settleTx, err := svc.SettlePayout("usr_1", 200, "bet_settlement", "bet_1")
	if err != nil {
		t.Fatalf("unexpected error on settle payout: %v", err)
	}
	if settleTx.Type != "settle" {
		t.Fatalf("expected settle transaction type, got %q", settleTx.Type)
	}

	w, err = svc.GetWallet("usr_1")
	if err != nil {
		t.Fatalf("unexpected error on get wallet after settle: %v", err)
	}
	if w.BalanceTokens != 1100 {
		t.Fatalf("expected balance 1100 after settle payout, got %v", w.BalanceTokens)
	}
}

func TestHoldInsufficientFunds(t *testing.T) {
	svc := NewService(50)

	_, err := svc.Hold("usr_2", 100, "bet", "bet_2")
	if err == nil {
		t.Fatal("expected insufficient funds error, got nil")
	}
	if err != ErrInsufficientFunds {
		t.Fatalf("expected ErrInsufficientFunds, got %v", err)
	}
}
