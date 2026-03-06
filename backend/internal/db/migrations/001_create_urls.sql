-- Migration 001: Core URL shortening table
-- Phase 1 — the only table needed to shorten and redirect URLs.

CREATE TABLE IF NOT EXISTS urls (
    id         BIGSERIAL    PRIMARY KEY,
    short_code TEXT         NOT NULL,
    long_url   TEXT         NOT NULL,
    created_at TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMPTZ,              -- NULL means this link never expires
    is_active  BOOLEAN      NOT NULL DEFAULT TRUE,  -- FALSE = soft-deleted

    CONSTRAINT uq_urls_short_code UNIQUE (short_code)
);

-- The UNIQUE constraint above already creates a full unique index on short_code.
-- This partial index is the one the redirect hot-path query actually uses:
--   WHERE short_code = $1 AND is_active = TRUE AND (expires_at IS NULL OR expires_at > NOW())
-- By indexing only active rows, the index stays small and fast as deleted links accumulate.
CREATE INDEX IF NOT EXISTS idx_urls_active_lookup
    ON urls (short_code)
    WHERE is_active = TRUE;
