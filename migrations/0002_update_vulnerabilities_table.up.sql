-- migrations/0002_update_vulnerabilities_table.up.sql

ALTER TABLE vulnerabilities
    ADD COLUMN title TEXT,
    ADD COLUMN severity TEXT,
    ADD COLUMN affected_versions TEXT[],
    ADD COLUMN fixed_versions TEXT[],
    ADD COLUMN aliases TEXT[],
    ADD COLUMN tags TEXT[],
    ADD COLUMN cwe TEXT[],
    ADD COLUMN epss JSONB;

-- Add an index on the new severity column for faster filtering.
CREATE INDEX IF NOT EXISTS idx_vuln_severity ON vulnerabilities (severity);

-- Add a GIN index to efficiently query array columns.
-- This allows checking for the existence of an element in aliases or tags.
CREATE INDEX IF NOT EXISTS idx_vuln_aliases_gin ON vulnerabilities USING GIN (aliases);
CREATE INDEX IF NOT EXISTS idx_vuln_tags_gin ON vulnerabilities USING GIN (tags); 