-- +goose Up
-- +goose StatementBegin
ALTER TABLE users
    ADD COLUMN IF NOT EXISTS email_verified BOOLEAN NOT NULL DEFAULT FALSE;

ALTER TABLE users
    ADD COLUMN IF NOT EXISTS verify_token TEXT NOT NULL DEFAULT '';

CREATE INDEX IF NOT EXISTS idx_users_verify_token ON users(verify_token);

ALTER TABLE events
    ADD COLUMN IF NOT EXISTS winner_outcome TEXT;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_users_verify_token;

ALTER TABLE users
    DROP COLUMN IF EXISTS verify_token;

ALTER TABLE users
    DROP COLUMN IF EXISTS email_verified;

ALTER TABLE events
    DROP COLUMN IF EXISTS winner_outcome;
-- +goose StatementEnd
