-- migrations/0004_create_assets_table.up.sql

-- Drop the table if it exists to ensure a clean slate for the new structure.
DROP TABLE IF EXISTS assets;

CREATE TABLE IF NOT EXISTS assets (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    type TEXT NOT NULL, -- Will store 'HOST', 'GIT_REPOSITORY', etc.
    owner_team TEXT,
    first_seen TIMESTAMPTZ,
    last_seen TIMESTAMPTZ,

    -- This single JSONB column will store polymorphic data: HostDetails, GitRepositoryDetails, etc.
    details JSONB,

    -- These columns remain useful for generic asset information.
    installed_products JSONB,
    tags JSONB, -- Storing tags as a JSON array of strings is more flexible.
    metadata JSONB
);

-- Indexes for common query patterns.
CREATE INDEX IF NOT EXISTS idx_assets_type ON assets (type);
CREATE INDEX IF NOT EXISTS idx_assets_owner_team ON assets (owner_team);
-- A GIN index allows for efficient searching within the JSONB 'details' column.
CREATE INDEX IF NOT EXISTS idx_assets_details_gin ON assets USING GIN (details); 