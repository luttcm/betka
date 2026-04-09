-- +goose Up
-- +goose StatementBegin
CREATE TABLE users (
    id BIGSERIAL PRIMARY KEY,
    email TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    role TEXT NOT NULL CHECK (role IN ('user', 'moderator', 'admin')),
    status TEXT NOT NULL CHECK (status IN ('active', 'blocked')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE wallets (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL UNIQUE REFERENCES users(id) ON DELETE CASCADE,
    balance_tokens NUMERIC(20, 8) NOT NULL DEFAULT 0 CHECK (balance_tokens >= 0),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE events (
    id BIGSERIAL PRIMARY KEY,
    creator_user_id BIGINT NOT NULL REFERENCES users(id),
    title TEXT NOT NULL,
    description TEXT NOT NULL,
    category TEXT,
    resolve_at TIMESTAMPTZ NOT NULL,
    status TEXT NOT NULL CHECK (status IN ('draft', 'pending', 'approved', 'rejected', 'settled', 'canceled')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE event_outcomes (
    id BIGSERIAL PRIMARY KEY,
    event_id BIGINT NOT NULL REFERENCES events(id) ON DELETE CASCADE,
    code TEXT NOT NULL,
    is_winner BOOLEAN,
    UNIQUE (event_id, code)
);

CREATE TABLE odds_snapshots (
    id BIGSERIAL PRIMARY KEY,
    event_id BIGINT NOT NULL REFERENCES events(id) ON DELETE CASCADE,
    outcome_id BIGINT NOT NULL REFERENCES event_outcomes(id) ON DELETE CASCADE,
    odds_decimal NUMERIC(10, 4) NOT NULL CHECK (odds_decimal > 1),
    margin_bps INTEGER NOT NULL CHECK (margin_bps >= 0),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE bets (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id),
    event_id BIGINT NOT NULL REFERENCES events(id),
    outcome_id BIGINT NOT NULL REFERENCES event_outcomes(id),
    stake NUMERIC(20, 8) NOT NULL CHECK (stake > 0),
    odds_at_bet NUMERIC(10, 4) NOT NULL CHECK (odds_at_bet > 1),
    potential_payout NUMERIC(20, 8) NOT NULL CHECK (potential_payout >= 0),
    status TEXT NOT NULL CHECK (status IN ('open', 'won', 'lost', 'refunded')),
    idempotency_key TEXT NOT NULL,
    placed_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    settled_at TIMESTAMPTZ,
    UNIQUE (user_id, idempotency_key)
);

CREATE TABLE moderation_tasks (
    id BIGSERIAL PRIMARY KEY,
    event_id BIGINT NOT NULL UNIQUE REFERENCES events(id) ON DELETE CASCADE,
    status TEXT NOT NULL CHECK (status IN ('pending', 'approved', 'rejected')),
    moderator_id BIGINT REFERENCES users(id),
    reason TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    reviewed_at TIMESTAMPTZ
);

CREATE TABLE wallet_transactions (
    id BIGSERIAL PRIMARY KEY,
    wallet_id BIGINT NOT NULL REFERENCES wallets(id) ON DELETE CASCADE,
    type TEXT NOT NULL CHECK (type IN ('subscribe', 'withdraw', 'hold', 'release', 'settle')),
    amount_tokens NUMERIC(20, 8) NOT NULL CHECK (amount_tokens > 0),
    ref_type TEXT,
    ref_id BIGINT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE audit_logs (
    id BIGSERIAL PRIMARY KEY,
    actor_user_id BIGINT REFERENCES users(id),
    action TEXT NOT NULL,
    entity_type TEXT NOT NULL,
    entity_id BIGINT,
    payload_json JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_events_status_resolve_at ON events(status, resolve_at);
CREATE INDEX idx_bets_user_placed_at ON bets(user_id, placed_at DESC);
CREATE INDEX idx_odds_snapshots_event_created_at ON odds_snapshots(event_id, created_at DESC);
CREATE INDEX idx_moderation_tasks_status_created_at ON moderation_tasks(status, created_at);
CREATE INDEX idx_wallet_transactions_wallet_created_at ON wallet_transactions(wallet_id, created_at DESC);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS audit_logs;
DROP TABLE IF EXISTS wallet_transactions;
DROP TABLE IF EXISTS moderation_tasks;
DROP TABLE IF EXISTS bets;
DROP TABLE IF EXISTS odds_snapshots;
DROP TABLE IF EXISTS event_outcomes;
DROP TABLE IF EXISTS events;
DROP TABLE IF EXISTS wallets;
DROP TABLE IF EXISTS users;
-- +goose StatementEnd
