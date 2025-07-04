package feed

import "time"

// Vulnerability represents a security vulnerability from any feed source
type Vulnerability struct {
	ID               string            `json:"id"`
	Title            string            `json:"title"`
	Description      string            `json:"description"`
	Source           string            `json:"source"`
	Severity         string            `json:"severity"`
	ProductName      string            `json:"product_name"`
	ProductVersion   string            `json:"product_version"`
	PublishedDate    time.Time         `json:"published_date"`
	LastModifiedDate time.Time         `json:"last_modified_date"`
	References       []string          `json:"references"`
	CVEIDs           []string          `json:"cve_ids"`
	ExploitAvailable bool              `json:"exploit_available"`
	Tags             []string          `json:"tags"`
	CVSS             *CVSSInfo         `json:"cvss,omitempty"`
	Metadata         map[string]string `json:"metadata,omitempty"`
}

// CVSSInfo represents CVSS scoring information
type CVSSInfo struct {
	V2Score  float64 `json:"v2_score,omitempty"`
	V3Score  float64 `json:"v3_score,omitempty"`
	V2Vector string  `json:"v2_vector,omitempty"`
	V3Vector string  `json:"v3_vector,omitempty"`
}

// FeedConfig represents configuration for a feed source
type FeedConfig struct {
	Name    string            `yaml:"name"`
	URL     string            `yaml:"url"`
	Type    string            `yaml:"type"`   // rss, atom, json, api
	Source  string            `yaml:"source"` // source identifier
	Enabled bool              `yaml:"enabled"`
	Timeout string            `yaml:"timeout,omitempty"`
	Headers map[string]string `yaml:"headers,omitempty"`
}
