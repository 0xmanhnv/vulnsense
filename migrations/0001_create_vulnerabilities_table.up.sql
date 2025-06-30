-- migrations/0001_create_vulnerabilities_table.up.sql

-- This table stores vulnerability information based on the Unified Vulnerability Format (UVF).
CREATE TABLE IF NOT EXISTS vulnerabilities (
    id TEXT PRIMARY KEY,
    source TEXT NOT NULL,
    description TEXT,
    published_date TIMESTAMPTZ,
    last_modified_date TIMESTAMPTZ,
    exploit_available BOOLEAN DEFAULT FALSE,

    -- JSONB columns for nested objects, providing both structure and flexibility.
    product JSONB,
    cvss JSONB,
    "references" JSONB
);

-- Indexes for common query patterns.
CREATE INDEX IF NOT EXISTS idx_vuln_source ON vulnerabilities (source);
CREATE INDEX IF NOT EXISTS idx_vuln_published_date ON vulnerabilities (published_date);

-- It's often useful to query for vulnerabilities affecting a specific product name.
-- PostgreSQL allows indexing on expressions and paths within a JSONB document.
CREATE INDEX IF NOT EXISTS idx_vuln_product_name ON vulnerabilities ((product->>'Name')); 