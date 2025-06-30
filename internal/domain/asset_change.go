package domain

import "time"

// AssetChange represents a single change to an asset's attribute over time.
// It serves as an audit log entry.
type AssetChange struct {
	AssetID   string    // The ID of the asset that was changed
	ChangedAt time.Time // Timestamp of when the change occurred
	Field     string    // The name of the field that was changed (e.g., "Status", "OwnerTeam")
	OldValue  string    // The previous value of the field
	NewValue  string    // The new value of the field
	ChangedBy string    // Who or what made the change (e.g., "system-scan", "user:john.doe")
}
