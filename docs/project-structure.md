# VulnSense Project Structure

## Overview
This document explains the VulnSense project structure, following Go best practices and Clean Architecture principles. The project implements Hexagonal Architecture (Ports & Adapters) pattern for maintainable, testable, and scalable code.

## Directory Structure

```
vulnsense/
├── cmd/                          # 📱 Applications & Entry Points
│   ├── vulnsense/               # Main application
│   │   └── main.go             # Main entry point
│   ├── scheduler/               # Background job scheduler
│   │   └── main.go             # Scheduler entry point
│   └── test-rss/               # RSS testing utility
│       └── main.go             # Test tool
│
├── internal/                    # 🔒 Private Application Code
│   ├── domain/                 # 💎 Business Domain Layer
│   │   ├── asset.go           # Asset aggregate root
│   │   ├── asset_types.go     # Asset type enums
│   │   ├── asset_details.go   # Asset detail value objects
│   │   ├── asset_change.go    # Asset change events
│   │   ├── vulnerability.go   # Vulnerability aggregate root
│   │   └── finding.go         # Finding aggregate root
│   │
│   ├── usecase/               # 🎯 Business Logic Layer
│   │   ├── interfaces.go      # Port interfaces (dependencies)
│   │   ├── fetch_vulnerabilities.go # Vulnerability fetching use case
│   │   ├── fetch_assets.go    # Asset fetching use case
│   │   ├── match_and_record.go # Vulnerability matching use case
│   │   └── notify_findings.go # Notification use case
│   │
│   ├── adapter/               # 🔌 External Adapters Layer
│   │   ├── rss_feed.go        # RSS feed implementation
│   │   ├── feed_converter.go  # Type conversion between layers
│   │   ├── feeds.go           # Feed provider implementations
│   │   ├── postgres_repo.go   # PostgreSQL repository
│   │   ├── memory_repo.go     # In-memory repository (testing)
│   │   ├── splunk.go          # Splunk asset fetcher
│   │   └── telegram.go        # Telegram notification adapter
│   │
│   ├── config/                # ⚙️ Configuration Management
│   │   └── config.go          # Application configuration
│   │
│   ├── app/                   # 🏗️ Application Assembly
│   │   └── app.go             # Dependency injection & wiring
│   │
│   └── task/                  # 🔄 Background Job Management
│       ├── tasks.go           # Task definitions
│       ├── scheduler.go       # Task scheduler
│       └── processor.go       # Task processor
│
├── pkg/                       # 📦 Public Library Code
│   └── feed/                  # RSS feed public API
│       ├── interfaces.go      # Public interfaces
│       ├── types.go           # Public types
│       ├── jsonurl/           # JSON URL utilities
│       └── opencve/           # OpenCVE specific types
│
├── configs/                   # 📝 Configuration Files
│   ├── sources.yaml          # RSS feed sources
│   ├── app.yaml              # Application config
│   └── app.env.example       # Environment variables example
│
├── docs/                      # 📚 Documentation
│   ├── architecture_feed_package.md
│   ├── rss-feeds.md
│   ├── advanced-detection-strategies.md
│   ├── background_jobs.md
│   └── VulnSense_Spec.md
│
├── migrations/                # 🗄️ Database Migrations
│   └── *.sql                 # SQL migration files
│
├── docker-compose.yml         # 🐳 Docker setup
├── Dockerfile                # 🐳 Container image
├── Makefile                  # 🛠️ Build automation
├── go.mod                    # 📋 Go module definition
├── go.sum                    # 🔐 Dependency checksums
└── README.md                 # 📖 Project documentation
```

## Architecture Layers

### 1. Application Layer (`cmd/`)
**Purpose**: Application entry points and executables

#### Structure:
- **`cmd/vulnsense/`**: Main application server
- **`cmd/scheduler/`**: Background job scheduler
- **`cmd/test-rss/`**: RSS testing utility

#### Go Best Practice:
```go
// cmd/vulnsense/main.go
func main() {
    // Initialize configuration
    // Wire dependencies
    // Start server
}
```

**Rule**: Each executable gets its own directory under `cmd/`

### 2. Domain Layer (`internal/domain/`)
**Purpose**: Core business entities and rules

#### Key Components:
- **Aggregate Roots**: `Asset`, `Vulnerability`, `Finding`
- **Value Objects**: `OS`, `CloudInfo`, `CVSS`, `Product`
- **Enums**: `AssetType`, `Severity`, `FindingStatus`

#### Example:
```go
// internal/domain/vulnerability.go
type Vulnerability struct {
    ID          string
    Title       string
    Severity    Severity
    // Business logic methods
}

func (v *Vulnerability) IsCritical() bool {
    return v.Severity == SeverityCritical
}
```

**Rules**:
- No external dependencies
- Pure business logic
- Rich domain models with behavior

### 3. Use Case Layer (`internal/usecase/`)
**Purpose**: Business logic orchestration and port definitions

#### Key Components:
- **Interfaces**: Port definitions for dependencies
- **Use Cases**: Business logic orchestration
- **DTOs**: Data transfer objects

#### Example:
```go
// internal/usecase/interfaces.go
type VulnerabilityRepository interface {
    Save(ctx context.Context, vuln domain.Vulnerability) error
    FindByID(ctx context.Context, id string) (domain.Vulnerability, error)
}

// internal/usecase/fetch_vulnerabilities.go
type FetchVulnerabilitiesUseCase struct {
    repo VulnerabilityRepository
    feeds []FeedProvider
}
```

**Rules**:
- Define interfaces here (Dependency Inversion)
- Orchestrate domain objects
- No direct external dependencies

### 4. Adapter Layer (`internal/adapter/`)
**Purpose**: External system integrations (implementations of ports)

#### Key Components:
- **Repositories**: Database implementations
- **External Services**: RSS feeds, Splunk, Telegram
- **Converters**: Type conversion between layers

#### Example:
```go
// internal/adapter/postgres_repo.go
type PostgresVulnerabilityRepository struct {
    pool *pgxpool.Pool
}

func (r *PostgresVulnerabilityRepository) Save(ctx context.Context, vuln domain.Vulnerability) error {
    // PostgreSQL implementation
}
```

**Rules**:
- Implement interfaces from usecase layer
- Handle external system specifics
- Convert between external and internal types

### 5. Configuration Layer (`internal/config/`)
**Purpose**: Application configuration management

#### Components:
- **Config struct**: Application configuration
- **Environment loading**: Viper integration
- **Validation**: Configuration validation

### 6. Application Assembly (`internal/app/`)
**Purpose**: Dependency injection and wiring

#### Example:
```go
// internal/app/app.go
type App struct {
    fetchVulns *usecase.FetchVulnerabilitiesUseCase
    // ... other dependencies
}

func New(cfg *config.Config) (*App, error) {
    // Wire all dependencies
    // Return assembled app
}
```

### 7. Public Library (`pkg/`)
**Purpose**: Public API that can be imported by external projects

#### Components:
- **Interfaces**: Public contracts
- **Types**: Public data structures
- **Utilities**: Reusable components

#### Example:
```go
// pkg/feed/interfaces.go
type Fetcher interface {
    Fetch(ctx context.Context) ([]*Vulnerability, error)
}
```

**Rules**:
- Can be imported by external projects
- Keep API stable and minimal
- No internal dependencies

## Go Best Practices Implementation

### 1. Interface Design
```go
// ✅ DO: Accept interfaces, return concrete types
func ProcessVulnerabilities(fetcher feed.Fetcher) []domain.Vulnerability {
    // implementation
}

// ✅ DO: Define interfaces where they're used
// usecase/interfaces.go defines VulnerabilityRepository
// adapter/postgres_repo.go implements it
```

### 2. Package Organization
```go
// ✅ DO: Group related functionality
// internal/domain/asset.go - Asset entity
// internal/domain/asset_types.go - Asset enums
// internal/domain/asset_details.go - Asset value objects
```

### 3. Dependency Direction
```
┌─────────────────┐
│   cmd/vulnsense │
└─────────────────┘
         │
         ▼
┌─────────────────┐
│  internal/app   │
└─────────────────┘
         │
         ▼
┌─────────────────┐
│ internal/usecase│◄─────────────────┐
└─────────────────┘                  │
         │                           │
         ▼                           │
┌─────────────────┐                  │
│ internal/adapter│──────────────────┘
└─────────────────┘
         │
         ▼
┌─────────────────┐
│   pkg/feed      │
└─────────────────┘
```

### 4. Access Control
- **`internal/`**: Only accessible within the project
- **`pkg/`**: Can be imported by external projects
- **`cmd/`**: Executable applications

## Development Guidelines

### 1. Adding New Features
1. **Domain**: Add entities/value objects to `internal/domain/`
2. **Use Case**: Add business logic to `internal/usecase/`
3. **Adapter**: Add external integrations to `internal/adapter/`
4. **Assembly**: Wire dependencies in `internal/app/`

### 2. Interface Placement
```go
// ✅ DO: Define interfaces in usecase layer
// internal/usecase/interfaces.go
type NotificationService interface {
    SendAlert(ctx context.Context, message string) error
}

// ✅ DO: Implement in adapter layer
// internal/adapter/telegram.go
type TelegramNotifier struct {}
func (t *TelegramNotifier) SendAlert(ctx context.Context, message string) error {
    // implementation
}
```

### 3. Type Conversion
```go
// ✅ DO: Convert between layers
// internal/adapter/feed_converter.go
func ConvertToDomainVulnerability(feedVuln feed.Vulnerability) domain.Vulnerability {
    // conversion logic
}
```

### 4. Testing Strategy
- **Unit Tests**: Domain layer (pure business logic)
- **Integration Tests**: Use case layer (with mocks)
- **End-to-End Tests**: Adapter layer (with real dependencies)

## Configuration Management

### 1. Environment Variables
```bash
# configs/app.env.example
APP_ENV=development
DB_HOST=localhost
DB_PORT=5432
```

### 2. YAML Configuration
```yaml
# configs/sources.yaml
- name: "BleepingComputer"
  url: "https://www.bleepingcomputer.com/feed/"
  type: "news"
  enabled: true
```

### 3. Configuration Loading
```go
// internal/config/config.go
func Load() (*Config, error) {
    // Load from environment
    // Load from YAML files
    // Validate configuration
}
```

## Background Jobs

### 1. Task Definitions
```go
// internal/task/tasks.go
const TypeFetchVulnerabilities = "task:fetch:vulnerabilities"

func NewFetchVulnerabilitiesTask() (*asynq.Task, error) {
    return asynq.NewTask(TypeFetchVulnerabilities, nil), nil
}
```

### 2. Scheduler
```go
// internal/task/scheduler.go
func (s *Scheduler) Start() error {
    // Register periodic tasks
    // Start scheduler
}
```

### 3. Processor
```go
// internal/task/processor.go
func (p *Processor) Start(app *app.App) error {
    // Register task handlers
    // Start worker
}
```

## Database Management

### 1. Migrations
```sql
-- migrations/001_create_vulnerabilities.sql
CREATE TABLE vulnerabilities (
    id VARCHAR(255) PRIMARY KEY,
    title TEXT NOT NULL,
    severity VARCHAR(50) NOT NULL
);
```

### 2. Repository Pattern
```go
// internal/adapter/postgres_repo.go
type PostgresVulnerabilityRepository struct {
    pool *pgxpool.Pool
}

func (r *PostgresVulnerabilityRepository) Save(ctx context.Context, vuln domain.Vulnerability) error {
    // SQL implementation
}
```

## Docker Configuration

### 1. Development Environment
```yaml
# docker-compose.dev.yml
version: '3.8'
services:
  postgres:
    image: postgres:15
    environment:
      POSTGRES_DB: vulnsense
      POSTGRES_USER: vulnsense
      POSTGRES_PASSWORD: password
```

### 2. Production Build
```dockerfile
# Dockerfile
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o main cmd/vulnsense/main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/main .
CMD ["./main"]
```

## Build Automation

### 1. Makefile
```makefile
# Makefile
.PHONY: build test run clean

build:
	go build -o bin/vulnsense cmd/vulnsense/main.go

test:
	go test ./...

run:
	go run cmd/vulnsense/main.go

clean:
	rm -rf bin/
```

## Key Benefits

### 1. **Maintainability**
- Clear separation of concerns
- Easy to understand and modify
- Consistent code organization

### 2. **Testability**
- Pure business logic in domain layer
- Dependency injection for easy mocking
- Clear interfaces for testing

### 3. **Scalability**
- Loosely coupled components
- Easy to add new features
- Pluggable architecture

### 4. **Team Collaboration**
- Clear boundaries between layers
- Consistent naming conventions
- Well-documented interfaces

## Common Patterns

### 1. Repository Pattern
```go
// Define in usecase layer
type VulnerabilityRepository interface {
    Save(ctx context.Context, vuln domain.Vulnerability) error
    FindByID(ctx context.Context, id string) (domain.Vulnerability, error)
}

// Implement in adapter layer
type PostgresVulnerabilityRepository struct {
    pool *pgxpool.Pool
}
```

### 2. Adapter Pattern
```go
// Convert between external and internal types
func ConvertToDomainVulnerability(feedVuln feed.Vulnerability) domain.Vulnerability {
    return domain.Vulnerability{
        ID:       feedVuln.ID,
        Title:    feedVuln.Title,
        Severity: convertSeverity(feedVuln.Severity),
    }
}
```

### 3. Factory Pattern
```go
// Create complex objects
func NewApp(cfg *config.Config) (*App, error) {
    // Initialize all dependencies
    // Wire them together
    // Return assembled app
}
```

---

This project structure follows Go best practices and Clean Architecture principles, creating a maintainable, testable, and scalable vulnerability management system. 