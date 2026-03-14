-- +goose Up
-- +goose StatementBegin
ALTER TABLE password_reset
ALTER COLUMN selector TYPE TEXT;

ALTER TABLE password_reset
ALTER COLUMN selector SET NOT NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd
