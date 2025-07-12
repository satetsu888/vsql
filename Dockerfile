# Build stage
FROM golang:1.24-alpine AS builder

# Install build dependencies (including gcc for CGO)
RUN apk add --no-cache git gcc musl-dev

# Set working directory
WORKDIR /build

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the binary with CGO enabled (required for pg_query_go)
RUN CGO_ENABLED=1 GOOS=linux go build -a -o vsql .

# Final stage
FROM alpine:3.19

# Install bash for entrypoint script
RUN apk add --no-cache bash

# Create vsql user
RUN addgroup -g 1000 vsql && \
    adduser -D -u 1000 -G vsql vsql

# Set working directory
WORKDIR /app

# Copy binary from builder
COPY --from=builder /build/vsql /app/vsql

# Copy entrypoint script
COPY docker-entrypoint.sh /app/docker-entrypoint.sh
RUN chmod +x /app/docker-entrypoint.sh

# Create seed directory
RUN mkdir -p /seed && chown vsql:vsql /seed

# Switch to non-root user
USER vsql

# Expose PostgreSQL port
EXPOSE 5432

# Set entrypoint
ENTRYPOINT ["/app/docker-entrypoint.sh"]

# Default command (can be overridden)
CMD []