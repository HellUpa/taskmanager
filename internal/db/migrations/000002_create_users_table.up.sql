BEGIN;

CREATE TABLE users (
    id UUID PRIMARY KEY,
    kratos_id VARCHAR(255) UNIQUE NOT NULL
);

ALTER TABLE tasks ADD COLUMN user_id UUID REFERENCES users(id) ON DELETE CASCADE;

COMMIT;