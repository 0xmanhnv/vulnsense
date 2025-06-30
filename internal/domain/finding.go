package domain

import "time"

// Finding represents a specific vulnerability detected on a specific asset.
// It is the central aggregate root for tracking the lifecycle of a vulnerability detection.
type Finding struct {
	ID              string // A unique identifier for this specific finding, e.g., a UUID.
	AssetID         string
	VulnerabilityID string
	DetectedAt      time.Time
	Status          FindingStatus // e.g., "OPEN", "PATCHED", "RISK_ACCEPTED"
	PatchedAt       *time.Time    // The time the finding was confirmed as patched. nil if not patched.
	Scanner         string        // The tool that detected the finding, e.g., "Qualys", "Snyk", "Trivy"
	DetectionMethod string        // How it was detected, e.g., "agent", "container-scan", "sast"
	PackagePath     string        // Optional: The path to the affected file or package, e.g., "/usr/lib/openssl"
	Evidence        string        // Optional: A snippet or log entry proving the finding
}

// FindingStatus represents the state of a finding in its lifecycle.
type FindingStatus string

const (
	FindingStatusOpen          FindingStatus = "OPEN"
	FindingStatusPatched       FindingStatus = "PATCHED"
	FindingStatusRiskAccepted  FindingStatus = "RISK_ACCEPTED"
	FindingStatusFalsePositive FindingStatus = "FALSE_POSITIVE"
)
