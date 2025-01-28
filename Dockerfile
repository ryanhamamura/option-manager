# Build stage
FROM golang:1.23-alpine AS builder

WORKDIR /build

# Install build dependencies
RUN apk add --no-cache gcc musl-dev

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o main ./cmd/main.go

# Final stage
FROM alpine:latest

WORKDIR /app

# Install runtime dependencies
RUN apk add --no-cache postgresql-client bash

# Install golang-migrate
RUN wget https://github.com/golang-migrate/migrate/releases/download/v4.17.0/migrate.linux-amd64.tar.gz \
    && tar -xf migrate.linux-amd64.tar.gz \
    && mv migrate /usr/local/bin/migrate \
    && rm migrate.linux-amd64.tar.gz

# Copy binary from builder
COPY --from=builder /build/main .

# Copy migrations and scripts
COPY templates/ ./templates/
COPY migrations/ ./migrations/
COPY scripts/entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh

EXPOSE 8080

ENTRYPOINT ["/entrypoint.sh"]
