package adapter

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"
	"vulnsense/pkg/feed"

	"github.com/mmcdole/gofeed"
)

// RSSFeedAdapter implements the feed.Fetcher interface using gofeed library
type RSSFeedAdapter struct {
	name        string
	url         string
	feedType    string // blog, advisory, vulnerability, etc.
	source      string
	client      *http.Client
	parser      *gofeed.Parser
	logger      *slog.Logger
	cvePattern  *regexp.Regexp
	prodPattern *regexp.Regexp
}

// NewRSSFeedAdapter creates a new RSS feed adapter
func NewRSSFeedAdapter(name, url, feedType, source string, client *http.Client, logger *slog.Logger) *RSSFeedAdapter {
	parser := gofeed.NewParser()
	parser.Client = client
	parser.UserAgent = "VulnSense/1.0 RSS Feed Reader"

	return &RSSFeedAdapter{
		name:        name,
		url:         url,
		feedType:    feedType,
		source:      source,
		client:      client,
		parser:      parser,
		logger:      logger.With("feed_name", name, "feed_type", "rss", "source", source),
		cvePattern:  regexp.MustCompile(`(?i)CVE-\d{4}-\d{4,}`),
		prodPattern: regexp.MustCompile(`(?i)\b(apache|nginx|mysql|postgresql|redis|docker|kubernetes|jenkins|wordpress|drupal|joomla|log4j|spring|struts|tomcat|openssl|openssh|php|python|java|node\.?js|react|angular|vue)\b`),
	}
}

// SourceName returns the name of the feed source
func (r *RSSFeedAdapter) SourceName() string {
	return r.name
}

// Fetch retrieves and parses RSS feed data using gofeed
func (r *RSSFeedAdapter) Fetch(ctx context.Context) ([]feed.Vulnerability, error) {
	r.logger.Info("Starting RSS feed fetch", "url", r.url)

	// Parse the feed
	gofeedFeed, err := r.parser.ParseURLWithContext(r.url, ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to parse RSS feed: %w", err)
	}

	vulnerabilities := make([]feed.Vulnerability, 0)
	for _, item := range gofeedFeed.Items {
		if vuln := r.convertItem(item); vuln != nil {
			vulnerabilities = append(vulnerabilities, *vuln)
		}
	}

	r.logger.Info("Successfully fetched RSS vulnerabilities", "count", len(vulnerabilities))
	return vulnerabilities, nil
}

// convertItem converts a gofeed.Item to a feed.Vulnerability
func (r *RSSFeedAdapter) convertItem(item *gofeed.Item) *feed.Vulnerability {
	// Skip items that don't seem to be security-related
	if !r.isSecurityRelated(item.Title, item.Description) {
		return nil
	}

	// Try to extract CVE IDs
	content := item.Title + " " + item.Description
	cveIDs := r.cvePattern.FindAllString(content, -1)
	for i, cve := range cveIDs {
		cveIDs[i] = strings.ToUpper(cve)
	}

	// Generate ID - prefer CVE ID if available, otherwise use GUID or generate one
	var id string
	if len(cveIDs) > 0 {
		id = cveIDs[0]
	} else if item.GUID != "" {
		id = fmt.Sprintf("RSS-%s-%s", r.source, r.shortHash(item.GUID))
	} else {
		id = fmt.Sprintf("RSS-%s-%s", r.source, r.shortHash(item.Link))
	}

	// Parse publication date
	pubDate := time.Now()
	if item.PublishedParsed != nil {
		pubDate = *item.PublishedParsed
	} else if item.UpdatedParsed != nil {
		pubDate = *item.UpdatedParsed
	}

	// Extract product information
	productName := r.extractProduct(content)

	// Determine severity based on feed type and keywords
	severity := r.determineSeverity(content, r.feedType)

	// Get description
	description := item.Description
	if description == "" && item.Content != "" {
		description = item.Content
	}

	vuln := &feed.Vulnerability{
		ID:               id,
		Title:            item.Title,
		Description:      description,
		Source:           r.source,
		Severity:         severity,
		ProductName:      productName,
		PublishedDate:    pubDate,
		LastModifiedDate: pubDate,
		References:       []string{item.Link},
		CVEIDs:           cveIDs,
		ExploitAvailable: r.hasExploit(content),
		Tags:             r.extractTags(content, r.feedType),
		Metadata: map[string]string{
			"feed_type": r.feedType,
			"feed_url":  r.url,
		},
	}

	return vuln
}

// isSecurityRelated checks if the item is security-related
func (r *RSSFeedAdapter) isSecurityRelated(title, description string) bool {
	securityKeywords := []string{
		"vulnerability", "exploit", "cve", "security", "patch", "fix", "bug",
		"rce", "xss", "sql injection", "csrf", "lfi", "rfi", "dos", "ddos",
		"buffer overflow", "privilege escalation", "authentication bypass",
		"directory traversal", "remote code execution", "cross-site scripting",
		"injection", "backdoor", "malware", "trojan", "ransomware",
	}

	content := strings.ToLower(title + " " + description)
	for _, keyword := range securityKeywords {
		if strings.Contains(content, keyword) {
			return true
		}
	}

	// Always include if it's from a vulnerability or advisory feed
	if r.feedType == "vulnerability" || r.feedType == "advisory" || r.feedType == "exploit" {
		return true
	}

	return false
}

// extractProduct tries to extract product information from text
func (r *RSSFeedAdapter) extractProduct(content string) string {
	contentLower := strings.ToLower(content)

	// Try to find known products
	matches := r.prodPattern.FindAllString(contentLower, -1)
	if len(matches) > 0 {
		return matches[0]
	}

	return "Unknown"
}

// determineSeverity determines severity based on content and feed type
func (r *RSSFeedAdapter) determineSeverity(content, feedType string) string {
	contentLower := strings.ToLower(content)

	// Critical keywords
	if strings.Contains(contentLower, "critical") || strings.Contains(contentLower, "rce") ||
		strings.Contains(contentLower, "remote code execution") {
		return "CRITICAL"
	}

	// High keywords
	if strings.Contains(contentLower, "high") || strings.Contains(contentLower, "privilege escalation") ||
		strings.Contains(contentLower, "authentication bypass") {
		return "HIGH"
	}

	// Medium keywords
	if strings.Contains(contentLower, "medium") || strings.Contains(contentLower, "xss") ||
		strings.Contains(contentLower, "sql injection") {
		return "MEDIUM"
	}

	// Low keywords
	if strings.Contains(contentLower, "low") || strings.Contains(contentLower, "information disclosure") {
		return "LOW"
	}

	// Default based on feed type
	switch feedType {
	case "exploit":
		return "HIGH"
	case "vulnerability", "advisory":
		return "MEDIUM"
	default:
		return "MEDIUM"
	}
}

// hasExploit checks if the item mentions available exploits
func (r *RSSFeedAdapter) hasExploit(content string) bool {
	contentLower := strings.ToLower(content)
	exploitKeywords := []string{
		"exploit", "poc", "proof of concept", "metasploit", "exploit-db",
		"publicly available", "in the wild", "weaponized",
		"proof-of-concept", "working exploit", "exploit code",
	}

	for _, keyword := range exploitKeywords {
		if strings.Contains(contentLower, keyword) {
			return true
		}
	}

	return false
}

// extractTags extracts relevant tags from the content
func (r *RSSFeedAdapter) extractTags(content, feedType string) []string {
	contentLower := strings.ToLower(content)
	tags := make([]string, 0)

	tagKeywords := map[string]string{
		"web":       "web application",
		"database":  "database",
		"network":   "network",
		"crypto":    "cryptographic",
		"auth":      "authentication",
		"mobile":    "mobile",
		"cloud":     "cloud",
		"container": "container",
		"api":       "api",
		"windows":   "windows",
		"linux":     "linux",
		"macos":     "macos",
	}

	for tag, keyword := range tagKeywords {
		if strings.Contains(contentLower, keyword) {
			tags = append(tags, tag)
		}
	}

	// Add feed type as tag
	tags = append(tags, feedType)

	return tags
}

// shortHash generates a short hash for IDs
func (r *RSSFeedAdapter) shortHash(input string) string {
	if len(input) < 8 {
		return input
	}

	// Use a simple hash of the string
	hash := 0
	for _, char := range input {
		hash = int(char) + ((hash << 5) - hash)
	}

	return strconv.FormatInt(int64(hash), 16)[:8]
}
