# syntax=docker/dockerfile:1
FROM golang:1.21 AS builder
WORKDIR /app
COPY . .
RUN go build -o vulnsense ./cmd/vulnsense

FROM debian:bullseye-slim
WORKDIR /app
COPY --from=builder /app/vulnsense /usr/local/bin/vulnsense
COPY configs/app.yaml /etc/vulnsense/app.yaml

# Optional: Add CA certificates, logging dependencies
RUN apt-get update && apt-get install -y ca-certificates && rm -rf /var/lib/apt/lists/*

CMD ["vulnsense", "--config", "/etc/vulnsense/app.yaml"]