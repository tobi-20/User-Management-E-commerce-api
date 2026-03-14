-- +goose Up
-- +goose StatementBegin
ALTER TABLE verification_tokens
ADD COLUMN is_verified BOOLEAN NOT NULL DEFAULT FALSE;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd
