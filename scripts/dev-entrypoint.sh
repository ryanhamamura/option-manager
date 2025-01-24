#!/bin/bash
set -e

# Development-specific environment setup
export CGO_ENABLED=0
export GO111MODULE=on

# Function to run database migrations
run_migrations() {
    echo "Running database migrations..."
    export POSTGRESQL_URL="postgres://${DB_USER}:${DB_PASSWORD}@postgres:5432/${DB_NAME}?sslmode=disable"
    
    # Add a longer timeout for development
    until pg_isready -h postgres -p 5432 -U ${DB_USER} -t 30; do
        echo "Waiting for PostgreSQL to be ready..."
        sleep 2
    done
    
    migrate -path /app/migrations -database "${POSTGRESQL_URL}" up
}

# Function to install/update Go dependencies
update_deps() {
    echo "Updating Go dependencies..."
    go mod tidy
    go mod verify
}

# Development utilities
case "$1" in
    "migrate")
        run_migrations
        ;;
    "deps")
        update_deps
        ;;
    "air")
        run_migrations
        air
        ;;
    "debug")
        run_migrations
        dlv debug --listen=:2345 --headless=true --api-version=2 --accept-multiclient ./cmd/main.go
        ;;
    *)
        # Default: run migrations and then execute the provided command or default to go run
        run_migrations
        if [ $# -eq 0 ]; then
            go run cmd/main.go
        else
            exec "$@"
        fi
        ;;
esac
