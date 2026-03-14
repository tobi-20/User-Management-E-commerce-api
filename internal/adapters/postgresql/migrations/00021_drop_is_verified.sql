-- +goose Up
-- +goose StatementBegin
ALTER TABLE verification_tokens
DROP COLUMN is_verified;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd
