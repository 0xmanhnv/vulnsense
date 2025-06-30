-- migrations/0003_create_findings_table.up.sql

CREATE TABLE IF NOT EXISTS findings (
    id TEXT PRIMARY KEY,
    asset_id TEXT NOT NULL,
    vulnerability_id TEXT NOT NULL,
    detected_at TIMESTAMPTZ NOT NULL,
    status TEXT NOT NULL,
    patched_at TIMESTAMPTZ,
    scanner TEXT,
    detection_method TEXT,
    package_path TEXT,
    evidence TEXT,

    -- A finding is unique for a specific vulnerability on a specific asset.
    -- This constraint prevents duplicate entries.
    CONSTRAINT uq_finding UNIQUE(asset_id, vulnerability_id)
);

-- Indexes for common query patterns.
CREATE INDEX IF NOT EXISTS idx_findings_asset_id ON findings (asset_id);
CREATE INDEX IF NOT EXISTS idx_findings_vulnerability_id ON findings (vulnerability_id);
CREATE INDEX IF NOT EXISTS idx_findings_status ON findings (status); 