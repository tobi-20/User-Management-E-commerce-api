-- +goose Up
-- +goose StatementBegin
ALTER TABLE verification_tokens
RENAME COLUMN verifier TO verifier_hash;

ALTER TABLE verification_tokens
RENAME COLUMN token TO selector;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd
