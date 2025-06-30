// Package domain contains the core business models and logic.
package domain

import "time"

// Asset is a polymorphic entity representing anything that can have vulnerabilities.
// It serves as a generic container, with specific data stored in the Details field.
type Asset struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Type      AssetType `json:"type"` // Crucial for determining the type of data in Details.
	OwnerTeam string    `json:"owner_team,omitempty"`
	Tags      []string  `json:"tags,omitempty"`
	FirstSeen time.Time `json:"first_seen,omitempty"`
	LastSeen  time.Time `json:"last_seen,omitempty"`

	// Details holds the specific information for the asset type (e.g., HostDetails, GitRepositoryDetails).
	// Using `any` makes this field polymorphic.
	Details any `json:"details"`

	// InstalledProducts can still be relevant for various asset types (e.g., libraries in a container or host).
	InstalledProducts []InstalledProduct `json:"installed_products,omitempty"`
	Metadata          map[string]any     `json:"metadata,omitempty"`
}

// OS represents the operating system information of an asset.
// This is a Value Object.
type OS struct {
	Name          string `json:"name,omitempty"`
	Version       string `json:"version,omitempty"`
	KernelVersion string `json:"kernel_version,omitempty"`
	Architecture  string `json:"architecture,omitempty"`
}

// CloudInfo represents cloud-specific details for an asset.
// This is a Value Object.
type CloudInfo struct {
	Provider   string `json:"provider,omitempty"`
	Region     string `json:"region,omitempty"`
	InstanceID string `json:"instance_id,omitempty"`
	AccountID  string `json:"account_id,omitempty"`
}

// InstalledProduct represents a specific piece of software installed on an asset.
// This is a Value Object within the Asset aggregate.
type InstalledProduct struct {
	Name          string   `json:"name,omitempty"`
	Version       string   `json:"version,omitempty"`
	CPE           string   `json:"cpe,omitempty"`
	Vendor        string   `json:"vendor,omitempty"`
	InstallPath   string   `json:"install_path,omitempty"`
	PackageSource string   `json:"package_source,omitempty"`
	Language      string   `json:"language,omitempty"`
	Signed        bool     `json:"signed,omitempty"`
	Tags          []string `json:"tags,omitempty"`
}
