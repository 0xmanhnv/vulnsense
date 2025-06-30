// Package domain contains the core business models and logic.
package domain

// HostDetails contains information specific to a host asset (server, VM).
type HostDetails struct {
	IPAddress       string    `json:"ip_address,omitempty"`
	OperatingSystem OS        `json:"operating_system,omitempty"`
	CloudInfo       CloudInfo `json:"cloud_info,omitempty"`
	Location        string    `json:"location,omitempty"`
}

// GitRepositoryDetails contains information specific to a Git repository asset.
type GitRepositoryDetails struct {
	URL           string `json:"url"`
	DefaultBranch string `json:"default_branch,omitempty"`
}

// ContainerImageDetails contains information specific to a container image asset.
type ContainerImageDetails struct {
	Registry string `json:"registry"`
	Name     string `json:"name"`
	Tag      string `json:"tag"`
	Digest   string `json:"digest,omitempty"`
}

// WebApplicationDetails contains information specific to a web application or API endpoint.
type WebApplicationDetails struct {
	URL string `json:"url"`
}
