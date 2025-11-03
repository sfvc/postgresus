-- +goose Up
-- +goose StatementBegin

-- Add status column to users table
ALTER TABLE users 
ADD COLUMN status TEXT NOT NULL DEFAULT 'ACTIVE';

-- Create index for status column for better performance
CREATE INDEX idx_users_status ON users (status);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- Remove status column from users table
DROP INDEX IF EXISTS idx_users_status;
ALTER TABLE users DROP COLUMN status;

-- +goose StatementEnd