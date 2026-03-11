-- Migration 002: Click tracking table
-- Each row records a single click event on a shortened URL.

CREATE TABLE IF NOT EXISTS clicks (
    id          BIGSERIAL    PRIMARY KEY,
    short_code  TEXT         NOT NULL REFERENCES urls(short_code) ON DELETE CASCADE,
    clicked_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

-- Index for fast per-code click lookups (used by GetByCode and GetCountByCode).
CREATE INDEX IF NOT EXISTS idx_clicks_short_code
    ON clicks (short_code);
