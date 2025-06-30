// Package domain contains the core business models and logic.
package domain

// AssetType is an enumeration for the different kinds of assets the system can handle.
type AssetType string

const (
	// AssetTypeHost represents a physical or virtual server, identified by IP or hostname.
	// This is for traditional infrastructure vulnerability management.
	AssetTypeHost AssetType = "HOST"

	// AssetTypeGitRepository represents a source code repository.
	// This is for SAST, secret scanning, and dependency scanning (SCA).
	AssetTypeGitRepository AssetType = "GIT_REPOSITORY"

	// AssetTypeContainerImage represents a Docker/OCI container image.
	// This is for container vulnerability scanning.
	AssetTypeContainerImage AssetType = "CONTAINER_IMAGE"

	// AssetTypeWebApplication represents a web application or API endpoint.
	// This is for DAST scanning.
	AssetTypeWebApplication AssetType = "WEB_APPLICATION"
)
