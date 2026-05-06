-- ============================================================
-- FILE: migrations/001_create_users.sql
-- WHAT IT IS:     Creates the users table for admin accounts
-- WHY IT EXISTS:  Reference SQL migration (GORM AutoMigrate handles
--                 actual table creation, but this documents the schema)
-- LAST UPDATED:   2026-05-07 — initial creation
-- ============================================================

CREATE TABLE IF NOT EXISTS users (
    id          SERIAL PRIMARY KEY,
    name        VARCHAR(100) NOT NULL,
    email       VARCHAR(255) UNIQUE NOT NULL,
    password    VARCHAR(255) NOT NULL,
    role        VARCHAR(50) DEFAULT 'admin',
    avatar      VARCHAR(500),
    created_at  TIMESTAMPTZ DEFAULT NOW(),
    updated_at  TIMESTAMPTZ DEFAULT NOW(),
    deleted_at  TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_users_deleted_at ON users(deleted_at);
