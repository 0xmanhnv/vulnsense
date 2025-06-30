# Background Job Processing Architecture

This document outlines the architecture for handling background tasks in VulnSense using `asynq`.

## 1. Overview

We use the `hibiken/asynq` library, which leverages Redis as a message broker, to create a robust and scalable background job processing system. This architecture separates the *scheduling* of tasks from the *execution* of tasks.

- **Scheduler**: A process that enqueues tasks at specific times (e.g., via a cron schedule).
- **Worker**: A process that dequeues tasks from the queue and executes the corresponding logic.
- **Redis**: Acts as the queue, storing tasks durably.

This design allows for:
- **Durability**: Tasks are not lost if the application restarts.
- **Scalability**: We can run multiple worker processes to handle high loads.
- **Separation of Concerns**: The main application (e.g., an API) is not blocked by long-running tasks.

## 2. Core Components

### `internal/task`
This package contains all logic related to background jobs.

- **`tasks.go`**: Defines the task "payloads" and "types". A task type is a unique string identifier (e.g., `"task:fetch:vulnerabilities"`). Payloads are structs that carry the data needed to execute the task.
- **`processor.go`**: Defines the `Processor` (our Worker). It initializes an `asynq.Server`, maps task types to their handler functions, and starts listening for jobs on the queue.
- **`scheduler.go`**: Defines the `Scheduler`. It initializes an `asynq.Scheduler` and registers periodic tasks with their cron schedules.
- **`logger_adapter.go`**: An adapter to make our standard `slog.Logger` compatible with the logger interface required by `asynq`.

### Entrypoints
The application now has two separate main entrypoints:

- **`cmd/vulnsense/main.go`**: This is the entrypoint for the **worker** process. Its sole responsibility is to start the task processor that listens for and executes jobs from the queue.
- **`cmd/scheduler/main.go`**: This is the entrypoint for the **scheduler** process. Its sole responsibility is to start the scheduler that enqueues periodic jobs.

### Docker
The `Dockerfile` is configured to build both the `vulnsense` (worker) and `scheduler` binaries. The `docker-compose.yml` then runs these as two separate services, each calling its respective binary. This provides better isolation and scalability.

## 3. How to Add a New Periodic Task

Let's say you want to create a new task that cleans up old findings every week.

**Step 1: Define the Task (in `tasks.go`)**
```go
// 1. Add new task type
const (
    // ... other types
    TypeCleanupOldFindings = "task:cleanup:findings"
)

// 2. Define the payload struct
type CleanupOldFindingsPayload struct {
    MaxAgeInDays int
}

// 3. Create a task constructor
func NewCleanupOldFindingsTask(maxAge int) (*asynq.Task, error) {
    payload, err := json.Marshal(CleanupOldFindingsPayload{MaxAgeInDays: maxAge})
    if err != nil {
        return nil, err
    }
    return asynq.NewTask(TypeCleanupOldFindings, payload), nil
}
```

**Step 2: Create the Handler (in `tasks.go` or a new file)**
```go
// This would need a new use case, e.g., CleanupFindingsUseCase
func HandleCleanupOldFindingsTask(ctx context.Context, t *asynq.Task, cleanupUseCase *usecase.CleanupUseCase) error {
    var p CleanupOldFindingsPayload
    if err := json.Unmarshal(t.Payload(), &p); err != nil {
        return fmt.Errorf("failed to unmarshal payload: %w", err)
    }
    
    return cleanupUseCase.Execute(ctx, p.MaxAgeInDays)
}
```

**Step 3: Register the Handler (in `processor.go`)**
```go
// Inside Processor.Start()
mux.HandleFunc(
    TypeCleanupOldFindings,
    func(ctx context.Context, t *asynq.Task) error {
        // Assume app.CleanupUseCase() exists
        return HandleCleanupOldFindingsTask(ctx, t, app.CleanupUseCase())
    },
)
```

**Step 4: Schedule the Task (in `scheduler.go`)**
```go
// Inside Scheduler.Start()
task, err := NewCleanupOldFindingsTask(90) // Clean findings older than 90 days
if err != nil {
    return err
}
// Cron spec for running every Sunday at 3 AM
entryID, err := s.scheduler.Register("0 3 * * SUN", task)
```

By following this pattern, you can add complex, scheduled background jobs to the application in a structured and maintainable way. 