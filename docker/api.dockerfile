# Build stage
FROM golang:1.23-alpine AS builder

# Add necessary build tools
RUN apk add --no-cache git

# Set working directory
WORKDIR /build

# Copy go mod files
COPY service_api/go.mod service_api/go.sum ./
RUN go mod download

# Copy source code
COPY service_api/ .

# Build the application with security flags
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o main ./cmd/api

# Final stage
FROM alpine:3.19

# Set working directory
WORKDIR /app

# Copy binary from builder
COPY --from=builder /build/main .

# Use root user
USER root

# Add metadata labels
LABEL maintainer="Poker Api" \
      version="1.0" \
      description="Poker Api"

# Expose API port
EXPOSE 7000:8000

# Run the application
CMD ["/app/main"]