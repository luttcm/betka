-- +goose Up
-- +goose StatementBegin
ALTER TABLE events
    ADD COLUMN IF NOT EXISTS settlement_requested_by BIGINT REFERENCES users(id),
    ADD COLUMN IF NOT EXISTS settlement_requested_at TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS settlement_evidence_url TEXT,
    ADD COLUMN IF NOT EXISTS settlement_evidence_file_name TEXT,
    ADD COLUMN IF NOT EXISTS settlement_evidence_file_data TEXT;

ALTER TABLE events
    DROP CONSTRAINT IF EXISTS events_status_check;

ALTER TABLE events
    ADD CONSTRAINT events_status_check
    CHECK (status IN ('draft', 'pending', 'approved', 'settlement_requested', 'rejected', 'settled', 'canceled'));

CREATE INDEX IF NOT EXISTS idx_events_status_settlement_requested_at
    ON events(status, settlement_requested_at DESC);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_events_status_settlement_requested_at;

ALTER TABLE events
    DROP CONSTRAINT IF EXISTS events_status_check;

ALTER TABLE events
    ADD CONSTRAINT events_status_check
    CHECK (status IN ('draft', 'pending', 'approved', 'rejected', 'settled', 'canceled'));

ALTER TABLE events
    DROP COLUMN IF EXISTS settlement_evidence_file_data,
    DROP COLUMN IF EXISTS settlement_evidence_file_name,
    DROP COLUMN IF EXISTS settlement_evidence_url,
    DROP COLUMN IF EXISTS settlement_requested_at,
    DROP COLUMN IF EXISTS settlement_requested_by;
-- +goose StatementEnd
