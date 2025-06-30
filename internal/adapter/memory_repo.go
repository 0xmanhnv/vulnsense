package adapter

import (
	"context"
	"fmt"
	"sync"

	"vulnsense/internal/domain"
)

// MemoryVulnerabilityRepository is an in-memory implementation of the VulnerabilityRepository interface.
// It is useful for testing and development. It is safe for concurrent use.
type MemoryVulnerabilityRepository struct {
	vulnerabilities map[string]domain.Vulnerability
	mu              sync.RWMutex
}

// NewMemoryVulnerabilityRepository creates a new, empty in-memory vulnerability repository.
func NewMemoryVulnerabilityRepository() *MemoryVulnerabilityRepository {
	return &MemoryVulnerabilityRepository{
		vulnerabilities: make(map[string]domain.Vulnerability),
	}
}

// Save stores a vulnerability in the repository. If a vulnerability with the same ID already exists, it will be overwritten.
func (r *MemoryVulnerabilityRepository) Save(ctx context.Context, vulnerability domain.Vulnerability) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if vulnerability.ID == "" {
		return fmt.Errorf("vulnerability ID cannot be empty")
	}
	r.vulnerabilities[vulnerability.ID] = vulnerability
	return nil
}

// FindByID retrieves a vulnerability by its ID. It returns an error if the vulnerability is not found.
func (r *MemoryVulnerabilityRepository) FindByID(ctx context.Context, id string) (domain.Vulnerability, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	vuln, ok := r.vulnerabilities[id]
	if !ok {
		return domain.Vulnerability{}, fmt.Errorf("vulnerability with ID '%s' not found", id)
	}
	return vuln, nil
}

// FindAll retrieves all vulnerabilities stored in the repository.
func (r *MemoryVulnerabilityRepository) FindAll(ctx context.Context) ([]domain.Vulnerability, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	all := make([]domain.Vulnerability, 0, len(r.vulnerabilities))
	for _, vuln := range r.vulnerabilities {
		all = append(all, vuln)
	}
	return all, nil
}
