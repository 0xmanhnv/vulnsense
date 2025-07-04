# Go Types Explained: Concrete Types vs Interface Types

## Overview
This document explains the fundamental difference between Concrete Types and Interface Types in Go programming, with practical examples from the VulnSense project. Understanding this distinction is crucial for writing maintainable, testable, and flexible Go code.

## What Are Concrete Types?

**Concrete Types** are types that have actual implementation and memory layout. They represent real data structures that can be instantiated and used directly.

### Examples of Concrete Types

#### 1. Struct Types
```go
// VulnSense examples
type User struct {
    ID    int
    Name  string
    Email string
}

type FetchVulnerabilitiesUseCase struct {
    providers []FeedProvider
    repo      VulnerabilityRepository
    logger    *slog.Logger
}

type Vulnerability struct {
    ID          string
    Title       string
    Severity    Severity
    Description string
}
```

#### 2. Basic Types
```go
type UserID int
type Email string
type Severity string

const (
    SeverityLow    Severity = "LOW"
    SeverityMedium Severity = "MEDIUM"
    SeverityHigh   Severity = "HIGH"
)
```

#### 3. Function Types
```go
type HandlerFunc func(http.ResponseWriter, *http.Request)
type ValidationFunc func(input string) error
```

#### 4. Collection Types
```go
type Users []User
type UserMap map[string]User
type VulnerabilityList []*Vulnerability
```

### Characteristics of Concrete Types
- ✅ **Have actual implementation**
- ✅ **Have specific memory layout**
- ✅ **Can be instantiated directly**
- ✅ **Have methods with actual code**
- ✅ **Are returned from functions**

## What Are Interface Types?

**Interface Types** define contracts (method signatures) but have no implementation. They specify what methods a type must have to satisfy the interface.

### Examples of Interface Types

#### 1. VulnSense Interface Examples
```go
// internal/usecase/interfaces.go
type VulnerabilityRepository interface {
    Save(ctx context.Context, vuln domain.Vulnerability) error
    FindByID(ctx context.Context, id string) (domain.Vulnerability, error)
    FindAll(ctx context.Context) ([]domain.Vulnerability, error)
}

type FeedProvider interface {
    Fetch(ctx context.Context) ([]domain.Vulnerability, error)
}

type Alerter interface {
    Notify(ctx context.Context, message string) error
}

// pkg/feed/interfaces.go
type Fetcher interface {
    Fetch(ctx context.Context) ([]*Vulnerability, error)
}
```

#### 2. Standard Library Examples
```go
type Reader interface {
    Read([]byte) (int, error)
}

type Writer interface {
    Write([]byte) (int, error)
}

type Stringer interface {
    String() string
}
```

### Characteristics of Interface Types
- ✅ **Define method signatures only**
- ✅ **Have no implementation**
- ✅ **Cannot be instantiated directly**
- ✅ **Are satisfied implicitly**
- ✅ **Are accepted as function parameters**

## Key Differences

| Aspect | Concrete Types | Interface Types |
|--------|---------------|-----------------|
| **Implementation** | Have actual code | Just method signatures |
| **Memory** | Have specific layout | No memory layout |
| **Instantiation** | Can create instances | Cannot instantiate |
| **Usage** | Return values, internal logic | Function parameters, dependencies |
| **Flexibility** | Specific behavior | Abstract behavior |

## Go Best Practice: "Accept Interfaces, Return Concrete Types"

This is one of the most important principles in Go programming.

### ✅ Good Examples

```go
// Accept interface, return concrete type
func ProcessVulnerabilities(fetcher Fetcher) []Vulnerability {
    //                         ↑              ↑
    //                   interface      concrete type
    vulns, err := fetcher.Fetch(context.Background())
    if err != nil {
        return nil
    }
    return processVulns(vulns)
}

// Accept interface for dependency injection
func NewFetchVulnerabilitiesUseCase(
    providers []FeedProvider,  // ← Interface (flexible)
    repo VulnerabilityRepository,  // ← Interface (flexible)
    logger *slog.Logger,  // ← Concrete type (specific)
) *FetchVulnerabilitiesUseCase {  // ← Return concrete type
    return &FetchVulnerabilitiesUseCase{
        providers: providers,
        repo:      repo,
        logger:    logger,
    }
}
```

### ❌ Bad Examples

```go
// BAD: Accept concrete type (inflexible)
func ProcessVulnerabilities(fetcher RSSFeedAdapter) []Vulnerability {
    //                         ↑
    //                   concrete type (too specific)
    // This function can only work with RSSFeedAdapter
    // Cannot use JSONFeedAdapter, XMLFeedAdapter, etc.
}

// BAD: Return interface (unclear)
func GetUser(id int) (UserInterface, error) {
    //                   ↑
    //              interface (vague)
    // Caller doesn't know what concrete type they're getting
    return &User{ID: id, Name: "John"}, nil
}
```

## Why This Principle Matters

### 1. Flexibility
```go
// With interfaces, we can use any implementation
type Fetcher interface {
    Fetch() ([]Vulnerability, error)
}

// All of these can be used interchangeably
var fetcher Fetcher = &RSSFeedAdapter{}
var fetcher Fetcher = &JSONFeedAdapter{}
var fetcher Fetcher = &XMLFeedAdapter{}
var fetcher Fetcher = &DatabaseFeedAdapter{}
```

### 2. Testability
```go
// Easy to mock for testing
type MockFetcher struct {
    vulnerabilities []Vulnerability
    error          error
}

func (m *MockFetcher) Fetch() ([]Vulnerability, error) {
    return m.vulnerabilities, m.error
}

// Test with mock
func TestProcessVulnerabilities(t *testing.T) {
    mockFetcher := &MockFetcher{
        vulnerabilities: []Vulnerability{
            {ID: "CVE-2023-1234", Title: "Test Vuln"},
        },
    }
    
    result := ProcessVulnerabilities(mockFetcher)
    assert.Equal(t, 1, len(result))
}
```

### 3. Decoupling
```go
// Use case doesn't depend on specific implementations
type FetchVulnerabilitiesUseCase struct {
    providers []FeedProvider  // ← Interface (any feed provider)
    repo      VulnerabilityRepository  // ← Interface (any repository)
    logger    *slog.Logger    // ← Concrete type (specific logger)
}

// Can inject any implementation:
// - PostgresRepository, MemoryRepository, RedisRepository
// - RSSFeedProvider, JSONFeedProvider, APIFeedProvider
```

## Practical Examples from VulnSense

### 1. Repository Pattern

#### Interface Definition (in usecase layer)
```go
// internal/usecase/interfaces.go
type VulnerabilityRepository interface {
    Save(ctx context.Context, vuln domain.Vulnerability) error
    FindByID(ctx context.Context, id string) (domain.Vulnerability, error)
    FindAll(ctx context.Context) ([]domain.Vulnerability, error)
}
```

#### Concrete Implementation (in adapter layer)
```go
// internal/adapter/postgres_repo.go
type PostgresVulnerabilityRepository struct {  // ← Concrete Type
    pool *pgxpool.Pool
}

func (r *PostgresVulnerabilityRepository) Save(ctx context.Context, vuln domain.Vulnerability) error {
    // PostgreSQL-specific implementation
    query := `INSERT INTO vulnerabilities (id, title, severity) VALUES ($1, $2, $3)`
    _, err := r.pool.Exec(ctx, query, vuln.ID, vuln.Title, vuln.Severity)
    return err
}

// internal/adapter/memory_repo.go
type MemoryVulnerabilityRepository struct {  // ← Another Concrete Type
    vulnerabilities map[string]domain.Vulnerability
    mu              sync.RWMutex
}

func (r *MemoryVulnerabilityRepository) Save(ctx context.Context, vuln domain.Vulnerability) error {
    // Memory-specific implementation
    r.mu.Lock()
    defer r.mu.Unlock()
    r.vulnerabilities[vuln.ID] = vuln
    return nil
}
```

### 2. Feed Provider Pattern

#### Interface Definition
```go
// internal/usecase/interfaces.go
type FeedProvider interface {
    Fetch(ctx context.Context) ([]domain.Vulnerability, error)
}
```

#### Concrete Implementations
```go
// internal/adapter/rss_feed.go
type RSSFeedAdapter struct {  // ← Concrete Type
    name     string
    url      string
    parser   *gofeed.Parser
    logger   *slog.Logger
}

func (r *RSSFeedAdapter) Fetch(ctx context.Context) ([]domain.Vulnerability, error) {
    // RSS-specific implementation
    feed, err := r.parser.ParseURL(r.url)
    if err != nil {
        return nil, err
    }
    return r.convertToVulnerabilities(feed.Items), nil
}

// internal/adapter/json_feed.go
type JSONFeedAdapter struct {  // ← Another Concrete Type
    apiURL string
    client *http.Client
}

func (j *JSONFeedAdapter) Fetch(ctx context.Context) ([]domain.Vulnerability, error) {
    // JSON API-specific implementation
    resp, err := j.client.Get(j.apiURL)
    if err != nil {
        return nil, err
    }
    // Parse JSON and convert to vulnerabilities
}
```

### 3. Usage in Use Cases

```go
// internal/usecase/fetch_vulnerabilities.go
type FetchVulnerabilitiesUseCase struct {
    providers []FeedProvider  // ← Interface (flexible)
    repo      VulnerabilityRepository  // ← Interface (flexible)
    logger    *slog.Logger    // ← Concrete type (specific)
}

func (uc *FetchVulnerabilitiesUseCase) Execute(ctx context.Context) error {
    var allVulns []domain.Vulnerability
    
    // Can iterate over any FeedProvider implementation
    for _, provider := range uc.providers {
        vulns, err := provider.Fetch(ctx)  // ← Interface method
        if err != nil {
            uc.logger.Error("Failed to fetch from provider", "error", err)
            continue
        }
        allVulns = append(allVulns, vulns...)
    }
    
    // Can use any VulnerabilityRepository implementation
    for _, vuln := range allVulns {
        if err := uc.repo.Save(ctx, vuln); err != nil {  // ← Interface method
            uc.logger.Error("Failed to save vulnerability", "error", err)
        }
    }
    
    return nil
}
```

## When to Use Concrete Types

### 1. Return Values
```go
// ✅ Return concrete types for clarity
func GetUser(id int) (*User, error) {
    return &User{ID: id, Name: "John"}, nil
}

func CreateVulnerability(title string) (*Vulnerability, error) {
    return &Vulnerability{
        ID:    generateID(),
        Title: title,
    }, nil
}
```

### 2. Internal Implementation
```go
// ✅ Use concrete types for internal logic
func (uc *FetchVulnerabilitiesUseCase) Execute(ctx context.Context) error {
    var vulns []domain.Vulnerability  // ← Concrete type
    
    for _, provider := range uc.providers {
        providerVulns, err := provider.Fetch(ctx)
        if err != nil {
            continue
        }
        vulns = append(vulns, providerVulns...)  // ← Working with concrete types
    }
    
    return uc.repo.Save(ctx, vulns)
}
```

### 3. Data Structures
```go
// ✅ Struct fields are concrete types
type User struct {
    ID       int       // ← Concrete type
    Name     string    // ← Concrete type
    Email    string    // ← Concrete type
    Created  time.Time // ← Concrete type
}
```

## When to Use Interface Types

### 1. Function Parameters
```go
// ✅ Accept interfaces for flexibility
func ProcessData(reader io.Reader, writer io.Writer) error {
    data, err := io.ReadAll(reader)
    if err != nil {
        return err
    }
    _, err = writer.Write(data)
    return err
}

// Can be used with files, network connections, memory buffers, etc.
```

### 2. Dependencies
```go
// ✅ Use interfaces for dependencies
type UserService struct {
    repo   UserRepository    // ← Interface
    cache  Cache            // ← Interface
    logger Logger           // ← Interface
}
```

### 3. Method Sets
```go
// ✅ Define behavior contracts
type Validator interface {
    Validate(input string) error
}

type Processor interface {
    Process(data []byte) ([]byte, error)
}
```

## Common Patterns in VulnSense

### 1. Repository Pattern
```go
// Define interface in usecase layer
type VulnerabilityRepository interface {
    Save(ctx context.Context, vuln domain.Vulnerability) error
    FindByID(ctx context.Context, id string) (domain.Vulnerability, error)
}

// Implement in adapter layer
type PostgresVulnerabilityRepository struct {
    pool *pgxpool.Pool
}

func (r *PostgresVulnerabilityRepository) Save(ctx context.Context, vuln domain.Vulnerability) error {
    // Implementation
}
```

### 2. Adapter Pattern
```go
// External type (from gofeed library)
type gofeed.Feed struct {
    Title string
    Items []*gofeed.Item
}

// Internal interface
type FeedParser interface {
    Parse(url string) (*Feed, error)
}

// Adapter concrete type
type GofeedAdapter struct {
    parser *gofeed.Parser
}

func (g *GofeedAdapter) Parse(url string) (*Feed, error) {
    // Convert external type to internal type
}
```

### 3. Factory Pattern
```go
// Return concrete type from factory
func NewFetchVulnerabilitiesUseCase(
    providers []FeedProvider,  // ← Accept interfaces
    repo VulnerabilityRepository,  // ← Accept interfaces
    logger *slog.Logger,  // ← Accept concrete type
) *FetchVulnerabilitiesUseCase {  // ← Return concrete type
    return &FetchVulnerabilitiesUseCase{
        providers: providers,
        repo:      repo,
        logger:    logger,
    }
}
```

## Testing Strategies

### 1. Mocking Interfaces
```go
// Easy to create mocks for interfaces
type MockVulnerabilityRepository struct {
    vulnerabilities []domain.Vulnerability
    saveError      error
}

func (m *MockVulnerabilityRepository) Save(ctx context.Context, vuln domain.Vulnerability) error {
    if m.saveError != nil {
        return m.saveError
    }
    m.vulnerabilities = append(m.vulnerabilities, vuln)
    return nil
}

func (m *MockVulnerabilityRepository) FindByID(ctx context.Context, id string) (domain.Vulnerability, error) {
    for _, vuln := range m.vulnerabilities {
        if vuln.ID == id {
            return vuln, nil
        }
    }
    return domain.Vulnerability{}, errors.New("not found")
}
```

### 2. Test Setup
```go
func TestFetchVulnerabilities(t *testing.T) {
    // Create mocks
    mockRepo := &MockVulnerabilityRepository{}
    mockProvider := &MockFeedProvider{
        vulnerabilities: []domain.Vulnerability{
            {ID: "CVE-2023-1234", Title: "Test Vuln"},
        },
    }
    
    // Create use case with mocks
    useCase := NewFetchVulnerabilitiesUseCase(
        []FeedProvider{mockProvider},
        mockRepo,
        slog.Default(),
    )
    
    // Test
    err := useCase.Execute(context.Background())
    assert.NoError(t, err)
    assert.Equal(t, 1, len(mockRepo.vulnerabilities))
}
```

## Performance Considerations

### 1. Interface Overhead
```go
// Interfaces have slight overhead due to dynamic dispatch
type Processor interface {
    Process(data []byte) []byte
}

// Direct call (faster)
var processor ConcreteProcessor
result := processor.Process(data)

// Interface call (slightly slower, but more flexible)
var processor Processor = &ConcreteProcessor{}
result := processor.Process(data)
```

### 2. When Performance Matters
```go
// For hot paths, consider concrete types
func (uc *FetchVulnerabilitiesUseCase) processBatch(vulns []domain.Vulnerability) {
    // Use concrete types for tight loops
    for i := 0; i < len(vulns); i++ {
        vuln := vulns[i]  // ← Concrete type, no interface overhead
        // Process vulnerability
    }
}
```

## Best Practices Summary

### ✅ Do This
1. **Accept interfaces** in function parameters
2. **Return concrete types** from functions
3. **Define interfaces** in the package that uses them
4. **Implement interfaces** in adapter/implementation packages
5. **Use small, focused interfaces** (single responsibility)
6. **Prefer composition** over large interfaces

### ❌ Don't Do This
1. **Don't accept concrete types** unless necessary
2. **Don't return interfaces** unless you have multiple implementations
3. **Don't define interfaces** just because you can
4. **Don't create huge interfaces** with many methods
5. **Don't use interfaces** for performance-critical code paths

## Real-World Example: HTTP Handler

```go
// ✅ Good: Accept interface, return concrete type
func HandleVulnerabilities(fetcher Fetcher) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        vulns, err := fetcher.Fetch(r.Context())
        if err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }
        
        // Return concrete type (JSON response)
        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(vulns)
    }
}

// Can be used with any Fetcher implementation
handler := HandleVulnerabilities(&RSSFeedAdapter{})
handler = HandleVulnerabilities(&JSONFeedAdapter{})
handler = HandleVulnerabilities(&DatabaseFeedAdapter{})
```

## Conclusion

Understanding the difference between concrete types and interface types is fundamental to writing good Go code. The key principles are:

1. **Interfaces define contracts** - what something can do
2. **Concrete types provide implementation** - how something does it
3. **Accept interfaces for flexibility** - your functions can work with any implementation
4. **Return concrete types for clarity** - callers know exactly what they're getting

This pattern makes your code more testable, maintainable, and flexible while keeping it clear and easy to understand.

In VulnSense, we use this pattern extensively:
- **Interfaces** for repositories, feed providers, and external services
- **Concrete types** for domain entities, use cases, and return values
- **Dependency injection** to wire everything together

This approach allows us to easily swap implementations, write comprehensive tests, and maintain clean architecture boundaries.

---

*Remember: In Go, interfaces are satisfied implicitly - you don't need to explicitly declare that a type implements an interface. If it has the required methods, it automatically satisfies the interface.* 