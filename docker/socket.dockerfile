# Build stage
FROM golang:1.23-alpine AS builder

# Add necessary build tools
RUN apk add --no-cache git

# Set working directory
WORKDIR /build

# Copy go mod files
COPY service_socket/go.mod service_socket/go.sum ./
RUN go mod download

# Copy source code
COPY service_socket/ .

# Build the application with security flags
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o main ./cmd/main

# Final stage
FROM alpine:3.19

# Set working directory
WORKDIR /app

# Copy binary from builder
COPY --from=builder /build/main .

# Use root user
USER root

# Add metadata labels
LABEL maintainer="Poker Socket" \
      version="1.0" \
      description="Poker Socket"

# Expose API port
EXPOSE 7070:7070

# Run the application
CMD ["/app/main"]