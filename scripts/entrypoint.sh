#!/bin/bash 
set -e 

echo "Starting Options Manager..."

# Debug: List contents of current directory
echo "Contents of current directory:"
ls -la 

# Wait for postgres 
until pg_isready -h postgres -p 5432 -U ${DB_USER}; do
  echo "Waiting for PostgreSQL to be ready..."
  sleep 2
done 

echo "Running database migrations..."
export POSTGRESQL_URL="postgres://${DB_USER}:${DB_PASSWORD}@postgres:5432/${DB_NAME}?sslmode=disable"
migrate -path ./migrations -database "${POSTGRESQL_URL}" up

echo "Starting the application..."
exec "./main"

