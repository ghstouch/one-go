# Build stage
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Install dependencies
RUN apk add --no-cache gcc musl-dev

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=1 GOOS=linux go build -o /app/bin/one ./cmd/server

# Runtime stage
FROM alpine:latest

WORKDIR /app

# Install runtime dependencies
RUN apk add --no-cache ca-certificates tzdata

# Copy binary from builder
COPY --from=builder /app/bin/one /app/one

# Copy web templates and static files
COPY --from=builder /app/web /app/web

# Create storage directory
RUN mkdir -p /app/storage

# Expose port
EXPOSE 8080

# Set environment variables
ENV GIN_MODE=release
ENV SERVER_PORT=8080
ENV SERVER_HOST=0.0.0.0

# Run the application
CMD ["/app/one"]
