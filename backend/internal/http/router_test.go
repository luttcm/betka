package http

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"bet/backend/internal/auth"
	"bet/backend/internal/config"
)

type healthResponse struct {
	Status    string `json:"status"`
	Timestamp string `json:"timestamp"`
}

type registerResponse struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	Role          string `json:"role"`
	EmailVerified bool   `json:"email_verified"`
}

type loginResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
}

type verifyEmailResponse struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	EmailVerified bool   `json:"email_verified"`
}

func TestHealthEndpoint(t *testing.T) {
	router := NewRouter(config.Config{AuthJWTSecret: "test-secret"})

	req := httptest.NewRequest(http.MethodGet, "/v1/health", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	var res healthResponse
	if err := json.Unmarshal(w.Body.Bytes(), &res); err != nil {
		t.Fatalf("failed to parse JSON response: %v", err)
	}

	if res.Status != "ok" {
		t.Fatalf("expected status field to be 'ok', got %q", res.Status)
	}

	if res.Timestamp == "" {
		t.Fatal("expected timestamp to be present")
	}
}

func TestAuthRegisterLoginAndMe(t *testing.T) {
	logBuf := captureLogs(t)

	router := NewRouter(config.Config{
		AuthJWTSecret:      "test-secret",
		EmailVerifyBaseURL: "http://localhost:3000/v1/auth/verify-email",
	})

	registerBody := map[string]string{
		"email":    "user@example.com",
		"password": "strong-password",
	}
	registerRaw, _ := json.Marshal(registerBody)

	registerReq := httptest.NewRequest(http.MethodPost, "/v1/auth/register", bytes.NewBuffer(registerRaw))
	registerReq.Header.Set("Content-Type", "application/json")
	registerW := httptest.NewRecorder()
	router.ServeHTTP(registerW, registerReq)

	if registerW.Code != http.StatusCreated {
		t.Fatalf("expected status %d on register, got %d", http.StatusCreated, registerW.Code)
	}

	var regRes registerResponse
	if err := json.Unmarshal(registerW.Body.Bytes(), &regRes); err != nil {
		t.Fatalf("failed to parse register response: %v", err)
	}

	if regRes.ID == "" || regRes.Email != "user@example.com" {
		t.Fatalf("unexpected register response: %+v", regRes)
	}

	loginBody := map[string]string{
		"email":    "user@example.com",
		"password": "strong-password",
	}
	loginRaw, _ := json.Marshal(loginBody)

	loginReq := httptest.NewRequest(http.MethodPost, "/v1/auth/login", bytes.NewBuffer(loginRaw))
	loginReq.Header.Set("Content-Type", "application/json")
	loginW := httptest.NewRecorder()
	router.ServeHTTP(loginW, loginReq)

	if loginW.Code != http.StatusForbidden {
		t.Fatalf("expected status %d on login before verify, got %d", http.StatusForbidden, loginW.Code)
	}

	verifyReq := httptest.NewRequest(http.MethodGet, "/v1/auth/verify-email?token=", nil)
	verifyW := httptest.NewRecorder()
	router.ServeHTTP(verifyW, verifyReq)

	if verifyW.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d on verify with empty token, got %d", http.StatusBadRequest, verifyW.Code)
	}

	verifyToken := extractVerifyTokenFromRegisterEmailLog(logBuf.String())
	if verifyToken == "" {
		t.Fatal("expected verify token to be present in register email log")
	}

	verifyReq = httptest.NewRequest(http.MethodGet, "/v1/auth/verify-email?token="+verifyToken, nil)
	verifyW = httptest.NewRecorder()
	router.ServeHTTP(verifyW, verifyReq)

	if verifyW.Code != http.StatusOK {
		t.Fatalf("expected status %d on verify, got %d", http.StatusOK, verifyW.Code)
	}

	var verifyRes verifyEmailResponse
	if err := json.Unmarshal(verifyW.Body.Bytes(), &verifyRes); err != nil {
		t.Fatalf("failed to parse verify response: %v", err)
	}

	if !verifyRes.EmailVerified {
		t.Fatalf("expected email to be verified, got %+v", verifyRes)
	}

	loginReq = httptest.NewRequest(http.MethodPost, "/v1/auth/login", bytes.NewBuffer(loginRaw))
	loginReq.Header.Set("Content-Type", "application/json")
	loginW = httptest.NewRecorder()
	router.ServeHTTP(loginW, loginReq)

	if loginW.Code != http.StatusOK {
		t.Fatalf("expected status %d on login after verify, got %d", http.StatusOK, loginW.Code)
	}

	var loginRes loginResponse
	if err := json.Unmarshal(loginW.Body.Bytes(), &loginRes); err != nil {
		t.Fatalf("failed to parse login response: %v", err)
	}

	if loginRes.AccessToken == "" || loginRes.TokenType != "Bearer" {
		t.Fatalf("unexpected login response: %+v", loginRes)
	}

	meReq := httptest.NewRequest(http.MethodGet, "/v1/me", nil)
	meReq.Header.Set("Authorization", "Bearer "+loginRes.AccessToken)
	meW := httptest.NewRecorder()
	router.ServeHTTP(meW, meReq)

	if meW.Code != http.StatusOK {
		t.Fatalf("expected status %d on me endpoint, got %d", http.StatusOK, meW.Code)
	}
}

func TestModerationEndpointRequiresRole(t *testing.T) {
	logBuf := captureLogs(t)

	router := NewRouter(config.Config{
		AuthJWTSecret:      "test-secret",
		EmailVerifyBaseURL: "http://localhost:3000/v1/auth/verify-email",
	})

	registerBody := map[string]string{
		"email":    "user2@example.com",
		"password": "strong-password",
	}
	registerRaw, _ := json.Marshal(registerBody)

	registerReq := httptest.NewRequest(http.MethodPost, "/v1/auth/register", bytes.NewBuffer(registerRaw))
	registerReq.Header.Set("Content-Type", "application/json")
	registerW := httptest.NewRecorder()
	router.ServeHTTP(registerW, registerReq)

	if registerW.Code != http.StatusCreated {
		t.Fatalf("expected status %d on register, got %d", http.StatusCreated, registerW.Code)
	}

	loginBody := map[string]string{
		"email":    "user2@example.com",
		"password": "strong-password",
	}
	loginRaw, _ := json.Marshal(loginBody)

	verifyToken := extractVerifyTokenFromRegisterEmailLog(logBuf.String())
	if verifyToken == "" {
		t.Fatal("expected verify token to be present in register email log")
	}

	verifyReq := httptest.NewRequest(http.MethodGet, "/v1/auth/verify-email?token="+verifyToken, nil)
	verifyW := httptest.NewRecorder()
	router.ServeHTTP(verifyW, verifyReq)

	if verifyW.Code != http.StatusOK {
		t.Fatalf("expected status %d on verify, got %d", http.StatusOK, verifyW.Code)
	}

	loginReq := httptest.NewRequest(http.MethodPost, "/v1/auth/login", bytes.NewBuffer(loginRaw))
	loginReq.Header.Set("Content-Type", "application/json")
	loginW := httptest.NewRecorder()
	router.ServeHTTP(loginW, loginReq)

	if loginW.Code != http.StatusOK {
		t.Fatalf("expected status %d on login, got %d", http.StatusOK, loginW.Code)
	}

	var loginRes loginResponse
	if err := json.Unmarshal(loginW.Body.Bytes(), &loginRes); err != nil {
		t.Fatalf("failed to parse login response: %v", err)
	}

	moderationReq := httptest.NewRequest(http.MethodGet, "/v1/moderation/health", nil)
	moderationReq.Header.Set("Authorization", "Bearer "+loginRes.AccessToken)
	moderationW := httptest.NewRecorder()
	router.ServeHTTP(moderationW, moderationReq)

	if moderationW.Code != http.StatusForbidden {
		t.Fatalf("expected status %d for user role on moderation endpoint, got %d", http.StatusForbidden, moderationW.Code)
	}
}

func TestEventModerationPublishFlow(t *testing.T) {
	router := NewRouter(config.Config{AuthJWTSecret: "test-secret"})

	creatorToken, err := auth.IssueToken("test-secret", time.Hour, "usr_creator", "user")
	if err != nil {
		t.Fatalf("failed to issue creator token: %v", err)
	}

	resolveAt := time.Now().UTC().Add(24 * time.Hour).Format(time.RFC3339)
	createBody := map[string]string{
		"title":       "Will company X close in 2026?",
		"description": "Community event for moderation flow test",
		"category":    "business",
		"resolve_at":  resolveAt,
	}
	createRaw, _ := json.Marshal(createBody)

	createReq := httptest.NewRequest(http.MethodPost, "/v1/events", bytes.NewBuffer(createRaw))
	createReq.Header.Set("Content-Type", "application/json")
	createReq.Header.Set("Authorization", "Bearer "+creatorToken)
	createW := httptest.NewRecorder()
	router.ServeHTTP(createW, createReq)

	if createW.Code != http.StatusCreated {
		t.Fatalf("expected status %d on event create, got %d", http.StatusCreated, createW.Code)
	}

	var created map[string]any
	if err := json.Unmarshal(createW.Body.Bytes(), &created); err != nil {
		t.Fatalf("failed to parse create event response: %v", err)
	}

	eventID, _ := created["id"].(string)
	if eventID == "" {
		t.Fatalf("expected event id in create response, got: %s", createW.Body.String())
	}

	listReq := httptest.NewRequest(http.MethodGet, "/v1/events", nil)
	listW := httptest.NewRecorder()
	router.ServeHTTP(listW, listReq)

	if listW.Code != http.StatusOK {
		t.Fatalf("expected status %d on events list before moderation, got %d", http.StatusOK, listW.Code)
	}

	if bytes.Contains(listW.Body.Bytes(), []byte(eventID)) {
		t.Fatalf("event %q must not appear in public list before approval", eventID)
	}

	moderatorToken, err := auth.IssueToken("test-secret", time.Hour, "usr_mod", "moderator")
	if err != nil {
		t.Fatalf("failed to issue moderator token: %v", err)
	}

	moderationListReq := httptest.NewRequest(http.MethodGet, "/v1/moderation/events", nil)
	moderationListReq.Header.Set("Authorization", "Bearer "+moderatorToken)
	moderationListW := httptest.NewRecorder()
	router.ServeHTTP(moderationListW, moderationListReq)

	if moderationListW.Code != http.StatusOK {
		t.Fatalf("expected status %d on moderation queue list, got %d", http.StatusOK, moderationListW.Code)
	}

	if !bytes.Contains(moderationListW.Body.Bytes(), []byte(eventID)) {
		t.Fatalf("expected event %q in moderation queue, body=%s", eventID, moderationListW.Body.String())
	}

	approveReq := httptest.NewRequest(http.MethodPost, "/v1/moderation/events/"+eventID+"/approve", nil)
	approveReq.Header.Set("Authorization", "Bearer "+moderatorToken)
	approveW := httptest.NewRecorder()
	router.ServeHTTP(approveW, approveReq)

	if approveW.Code != http.StatusOK {
		t.Fatalf("expected status %d on approve, got %d", http.StatusOK, approveW.Code)
	}

	getReq := httptest.NewRequest(http.MethodGet, "/v1/events/"+eventID, nil)
	getW := httptest.NewRecorder()
	router.ServeHTTP(getW, getReq)

	if getW.Code != http.StatusOK {
		t.Fatalf("expected status %d on get approved event, got %d", http.StatusOK, getW.Code)
	}

	listReq = httptest.NewRequest(http.MethodGet, "/v1/events", nil)
	listW = httptest.NewRecorder()
	router.ServeHTTP(listW, listReq)

	if listW.Code != http.StatusOK {
		t.Fatalf("expected status %d on events list after moderation, got %d", http.StatusOK, listW.Code)
	}

	if !bytes.Contains(listW.Body.Bytes(), []byte(eventID)) {
		t.Fatalf("expected approved event %q in public list, body=%s", eventID, listW.Body.String())
	}
}

func TestBetsFlowWithWalletHoldAndIdempotency(t *testing.T) {
	router := NewRouter(config.Config{AuthJWTSecret: "test-secret"})

	creatorToken, err := auth.IssueToken("test-secret", time.Hour, "usr_creator_bets", "user")
	if err != nil {
		t.Fatalf("failed to issue creator token: %v", err)
	}

	moderatorToken, err := auth.IssueToken("test-secret", time.Hour, "usr_mod_bets", "moderator")
	if err != nil {
		t.Fatalf("failed to issue moderator token: %v", err)
	}

	bettorToken, err := auth.IssueToken("test-secret", time.Hour, "usr_bettor", "user")
	if err != nil {
		t.Fatalf("failed to issue bettor token: %v", err)
	}

	eventID := createApprovedEvent(t, router, creatorToken, moderatorToken)

	walletReq := httptest.NewRequest(http.MethodGet, "/v1/wallet", nil)
	walletReq.Header.Set("Authorization", "Bearer "+bettorToken)
	walletW := httptest.NewRecorder()
	router.ServeHTTP(walletW, walletReq)

	if walletW.Code != http.StatusOK {
		t.Fatalf("expected status %d on wallet get, got %d", http.StatusOK, walletW.Code)
	}

	if !bytes.Contains(walletW.Body.Bytes(), []byte("1000")) {
		t.Fatalf("expected initial wallet balance in response, body=%s", walletW.Body.String())
	}

	betBody := map[string]any{
		"event_id":     eventID,
		"outcome_code": "yes",
		"stake":        100.0,
	}
	betRaw, _ := json.Marshal(betBody)

	placeReq := httptest.NewRequest(http.MethodPost, "/v1/bets", bytes.NewBuffer(betRaw))
	placeReq.Header.Set("Content-Type", "application/json")
	placeReq.Header.Set("Authorization", "Bearer "+bettorToken)
	placeReq.Header.Set("Idempotency-Key", "idem-bet-1")
	placeW := httptest.NewRecorder()
	router.ServeHTTP(placeW, placeReq)

	if placeW.Code != http.StatusCreated {
		t.Fatalf("expected status %d on first place bet, got %d, body=%s", http.StatusCreated, placeW.Code, placeW.Body.String())
	}

	var createdBet map[string]any
	if err := json.Unmarshal(placeW.Body.Bytes(), &createdBet); err != nil {
		t.Fatalf("failed to parse created bet response: %v", err)
	}

	betID, _ := createdBet["id"].(string)
	if betID == "" {
		t.Fatalf("expected created bet id, body=%s", placeW.Body.String())
	}

	repeatReq := httptest.NewRequest(http.MethodPost, "/v1/bets", bytes.NewBuffer(betRaw))
	repeatReq.Header.Set("Content-Type", "application/json")
	repeatReq.Header.Set("Authorization", "Bearer "+bettorToken)
	repeatReq.Header.Set("Idempotency-Key", "idem-bet-1")
	repeatW := httptest.NewRecorder()
	router.ServeHTTP(repeatW, repeatReq)

	if repeatW.Code != http.StatusOK {
		t.Fatalf("expected status %d on idempotent retry, got %d, body=%s", http.StatusOK, repeatW.Code, repeatW.Body.String())
	}

	if !bytes.Contains(repeatW.Body.Bytes(), []byte(betID)) {
		t.Fatalf("expected same bet id %q on idempotent retry, body=%s", betID, repeatW.Body.String())
	}

	walletReq = httptest.NewRequest(http.MethodGet, "/v1/wallet", nil)
	walletReq.Header.Set("Authorization", "Bearer "+bettorToken)
	walletW = httptest.NewRecorder()
	router.ServeHTTP(walletW, walletReq)

	if walletW.Code != http.StatusOK {
		t.Fatalf("expected status %d on wallet get after hold, got %d", http.StatusOK, walletW.Code)
	}

	if !bytes.Contains(walletW.Body.Bytes(), []byte("900")) {
		t.Fatalf("expected wallet balance reduced after hold, body=%s", walletW.Body.String())
	}

	myBetsReq := httptest.NewRequest(http.MethodGet, "/v1/bets/my", nil)
	myBetsReq.Header.Set("Authorization", "Bearer "+bettorToken)
	myBetsW := httptest.NewRecorder()
	router.ServeHTTP(myBetsW, myBetsReq)

	if myBetsW.Code != http.StatusOK {
		t.Fatalf("expected status %d on my bets list, got %d", http.StatusOK, myBetsW.Code)
	}

	if !bytes.Contains(myBetsW.Body.Bytes(), []byte(betID)) {
		t.Fatalf("expected bet id %q in my bets, body=%s", betID, myBetsW.Body.String())
	}

	txReq := httptest.NewRequest(http.MethodGet, "/v1/wallet/transactions", nil)
	txReq.Header.Set("Authorization", "Bearer "+bettorToken)
	txW := httptest.NewRecorder()
	router.ServeHTTP(txW, txReq)

	if txW.Code != http.StatusOK {
		t.Fatalf("expected status %d on transactions list, got %d", http.StatusOK, txW.Code)
	}

	if !bytes.Contains(txW.Body.Bytes(), []byte("\"type\":\"hold\"")) {
		t.Fatalf("expected hold transaction in wallet transactions, body=%s", txW.Body.String())
	}

	insufficientBody := map[string]any{
		"event_id":     eventID,
		"outcome_code": "no",
		"stake":        2000.0,
	}
	insufficientRaw, _ := json.Marshal(insufficientBody)

	insufficientReq := httptest.NewRequest(http.MethodPost, "/v1/bets", bytes.NewBuffer(insufficientRaw))
	insufficientReq.Header.Set("Content-Type", "application/json")
	insufficientReq.Header.Set("Authorization", "Bearer "+bettorToken)
	insufficientReq.Header.Set("Idempotency-Key", "idem-bet-2")
	insufficientW := httptest.NewRecorder()
	router.ServeHTTP(insufficientW, insufficientReq)

	if insufficientW.Code != http.StatusConflict {
		t.Fatalf("expected status %d on insufficient funds, got %d, body=%s", http.StatusConflict, insufficientW.Code, insufficientW.Body.String())
	}
}

func TestAdminSettlementFlow(t *testing.T) {
	router := NewRouter(config.Config{AuthJWTSecret: "test-secret"})

	creatorToken, err := auth.IssueToken("test-secret", time.Hour, "usr_creator_settle", "user")
	if err != nil {
		t.Fatalf("failed to issue creator token: %v", err)
	}

	moderatorToken, err := auth.IssueToken("test-secret", time.Hour, "usr_mod_settle", "moderator")
	if err != nil {
		t.Fatalf("failed to issue moderator token: %v", err)
	}

	adminToken, err := auth.IssueToken("test-secret", time.Hour, "usr_admin_settle", "admin")
	if err != nil {
		t.Fatalf("failed to issue admin token: %v", err)
	}

	bettorToken, err := auth.IssueToken("test-secret", time.Hour, "usr_bettor_settle", "user")
	if err != nil {
		t.Fatalf("failed to issue bettor token: %v", err)
	}

	eventID := createApprovedEvent(t, router, creatorToken, moderatorToken)

	betBody := map[string]any{
		"event_id":     eventID,
		"outcome_code": "yes",
		"stake":        100.0,
	}
	betRaw, _ := json.Marshal(betBody)

	placeReq := httptest.NewRequest(http.MethodPost, "/v1/bets", bytes.NewBuffer(betRaw))
	placeReq.Header.Set("Content-Type", "application/json")
	placeReq.Header.Set("Authorization", "Bearer "+bettorToken)
	placeReq.Header.Set("Idempotency-Key", "idem-settle-1")
	placeW := httptest.NewRecorder()
	router.ServeHTTP(placeW, placeReq)

	if placeW.Code != http.StatusCreated {
		t.Fatalf("expected status %d on place bet, got %d, body=%s", http.StatusCreated, placeW.Code, placeW.Body.String())
	}

	settleBody := map[string]string{"winner_outcome": "yes"}
	settleRaw, _ := json.Marshal(settleBody)

	settleReq := httptest.NewRequest(http.MethodPost, "/v1/admin/events/"+eventID+"/settle", bytes.NewBuffer(settleRaw))
	settleReq.Header.Set("Content-Type", "application/json")
	settleReq.Header.Set("Authorization", "Bearer "+adminToken)
	settleW := httptest.NewRecorder()
	router.ServeHTTP(settleW, settleReq)

	if settleW.Code != http.StatusOK {
		t.Fatalf("expected status %d on settlement, got %d, body=%s", http.StatusOK, settleW.Code, settleW.Body.String())
	}

	if !bytes.Contains(settleW.Body.Bytes(), []byte("\"status\":\"settled\"")) {
		t.Fatalf("expected settled event in response, body=%s", settleW.Body.String())
	}

	if !bytes.Contains(settleW.Body.Bytes(), []byte("\"status\":\"won\"")) {
		t.Fatalf("expected won bet status in settlement response, body=%s", settleW.Body.String())
	}

	walletReq := httptest.NewRequest(http.MethodGet, "/v1/wallet", nil)
	walletReq.Header.Set("Authorization", "Bearer "+bettorToken)
	walletW := httptest.NewRecorder()
	router.ServeHTTP(walletW, walletReq)

	if walletW.Code != http.StatusOK {
		t.Fatalf("expected status %d on wallet get after settlement, got %d", http.StatusOK, walletW.Code)
	}

	if !bytes.Contains(walletW.Body.Bytes(), []byte("1100")) {
		t.Fatalf("expected wallet payout after settlement, body=%s", walletW.Body.String())
	}

	txReq := httptest.NewRequest(http.MethodGet, "/v1/wallet/transactions", nil)
	txReq.Header.Set("Authorization", "Bearer "+bettorToken)
	txW := httptest.NewRecorder()
	router.ServeHTTP(txW, txReq)

	if txW.Code != http.StatusOK {
		t.Fatalf("expected status %d on transactions list after settlement, got %d", http.StatusOK, txW.Code)
	}

	if !bytes.Contains(txW.Body.Bytes(), []byte("\"type\":\"settle\"")) {
		t.Fatalf("expected settle transaction in wallet transactions, body=%s", txW.Body.String())
	}
}

func extractVerifyTokenFromRegisterEmailLog(body string) string {
	const marker = "token="
	idx := bytes.Index([]byte(body), []byte(marker))
	if idx == -1 {
		return ""
	}

	start := idx + len(marker)
	if start >= len(body) {
		return ""
	}

	end := start
	for end < len(body) {
		ch := body[end]
		if (ch < '0' || ch > '9') && (ch < 'a' || ch > 'f') {
			break
		}
		end++
	}

	if end == start {
		return ""
	}

	return body[start:end]
}

func captureLogs(t *testing.T) *bytes.Buffer {
	t.Helper()

	buf := &bytes.Buffer{}
	log.SetOutput(buf)
	t.Cleanup(func() {
		log.SetOutput(os.Stderr)
	})

	return buf
}

func createApprovedEvent(t *testing.T, router http.Handler, creatorToken, moderatorToken string) string {
	t.Helper()

	resolveAt := time.Now().UTC().Add(24 * time.Hour).Format(time.RFC3339)
	createBody := map[string]string{
		"title":       "Will product hit 10k users?",
		"description": "Event for bets flow test",
		"category":    "product",
		"resolve_at":  resolveAt,
	}
	createRaw, _ := json.Marshal(createBody)

	createReq := httptest.NewRequest(http.MethodPost, "/v1/events", bytes.NewBuffer(createRaw))
	createReq.Header.Set("Content-Type", "application/json")
	createReq.Header.Set("Authorization", "Bearer "+creatorToken)
	createW := httptest.NewRecorder()
	router.ServeHTTP(createW, createReq)

	if createW.Code != http.StatusCreated {
		t.Fatalf("expected status %d on event create, got %d, body=%s", http.StatusCreated, createW.Code, createW.Body.String())
	}

	var created map[string]any
	if err := json.Unmarshal(createW.Body.Bytes(), &created); err != nil {
		t.Fatalf("failed to parse create event response: %v", err)
	}

	eventID, _ := created["id"].(string)
	if eventID == "" {
		t.Fatalf("expected event id in create response, got: %s", createW.Body.String())
	}

	approveReq := httptest.NewRequest(http.MethodPost, "/v1/moderation/events/"+eventID+"/approve", nil)
	approveReq.Header.Set("Authorization", "Bearer "+moderatorToken)
	approveW := httptest.NewRecorder()
	router.ServeHTTP(approveW, approveReq)

	if approveW.Code != http.StatusOK {
		t.Fatalf("expected status %d on approve, got %d, body=%s", http.StatusOK, approveW.Code, approveW.Body.String())
	}

	return eventID
}
