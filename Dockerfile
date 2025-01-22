FROM golang:1.22-alpine

WORKDIR /app

# Install necessary build tools and runtime dependencies
RUN apk add --no-cache gcc musl-dev bash postgresql-client 

# Install golang-migrate 
RUN go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@v4.17.0

# Copy go.mod and go.sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the source code
COPY . .

RUN echo "Project structure:" && \
    find . -type f -print

# Build the application with verbose output
RUN CGO_ENABLED=0 GOOS=linux go build -v -o ./main ./cmd/main.go

RUN echo "Project structure:" && \
    find . -type f -print

# Expose port 8080
EXPOSE 8080

# Copy entrypoint script
COPY scripts/entrypoint.sh /
RUN chmod +x /entrypoint.sh

ENTRYPOINT ["/entrypoint.sh"]
