-- +goose Up
-- +goose StatementBegin
ALTER TABLE users
add column verified_expiry TIMESTAMPTZ;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd
