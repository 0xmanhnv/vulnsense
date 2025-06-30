# syntax=docker/dockerfile:1
FROM golang:1.24 AS builder
WORKDIR /app

COPY . .

# Build the worker binary
RUN go build -o /vulnsense ./cmd/vulnsense

# Build the scheduler binary
RUN go build -o /scheduler ./cmd/scheduler

FROM debian:bullseye-slim
WORKDIR /app

# Copy both binaries from the builder stage
COPY --from=builder /vulnsense /usr/local/bin/vulnsense
COPY --from=builder /scheduler /usr/local/bin/scheduler

# Copy configuration (optional, can be mounted via volumes)
COPY configs/app.yaml /etc/vulnsense/app.yaml

# Install dependencies needed by the application
RUN apt-get update && apt-get install -y ca-certificates && rm -rf /var/lib/apt/lists/*

# Default command runs the worker
CMD ["vulnsense"]