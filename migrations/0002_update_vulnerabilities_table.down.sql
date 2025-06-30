-- migrations/0002_update_vulnerabilities_table.down.sql

ALTER TABLE vulnerabilities
    DROP COLUMN IF EXISTS title,
    DROP COLUMN IF EXISTS severity,
    DROP COLUMN IF EXISTS affected_versions,
    DROP COLUMN IF EXISTS fixed_versions,
    DROP COLUMN IF EXISTS aliases,
    DROP COLUMN IF EXISTS tags,
    DROP COLUMN IF EXISTS cwe,
    DROP COLUMN IF EXISTS epss; 