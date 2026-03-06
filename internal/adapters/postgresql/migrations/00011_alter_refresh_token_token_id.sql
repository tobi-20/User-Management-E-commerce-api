-- +goose Up
-- +goose StatementBegin
ALTER TABLE refresh_tokens
ADD COLUMN token_id TEXT UNIQUE NOT NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd
