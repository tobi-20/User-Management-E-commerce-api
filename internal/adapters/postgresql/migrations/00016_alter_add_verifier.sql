-- +goose Up
-- +goose StatementBegin
ALTER TABLE verification_tokens
ADD COLUMN verifier TEXT NOT NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd
