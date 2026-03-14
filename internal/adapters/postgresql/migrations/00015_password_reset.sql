-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS password_reset (
  hashed_reset_token TEXT NOT NULL,
  user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  is_used BOOLEAN NOT NULL DEFAULT FALSE,
  expiry TIMESTAMPTZ NOT NULL

);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd
