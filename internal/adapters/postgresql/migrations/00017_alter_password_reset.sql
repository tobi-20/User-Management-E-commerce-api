-- +goose Up
-- +goose StatementBegin
ALTER TABLE password_reset
ADD COLUMN selector TEXT;

ALTER TABLE password_reset
RENAME COLUMN hashed_reset_token TO verifier_hash;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd
