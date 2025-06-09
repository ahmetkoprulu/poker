# Build stage
FROM golang:1.23-alpine AS builder

# Add necessary build tools
RUN apk add --no-cache git

# Set working directory
WORKDIR /build

# Copy go mod files
COPY service_consumers/go.mod service_consumers/go.sum ./
RUN go mod download

# Copy source code
COPY service_consumers/ .

# Build the application with security flags
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o main ./cmd/service_consumers

# Final stage
FROM alpine:3.19

# Set working directory
WORKDIR /app

# Copy binary from builder
COPY --from=builder /build/main .

# Use root user
USER root

# Add metadata labels
LABEL maintainer="Poker service_consumers" \
      version="1.0" \
      description="Poker service_consumers"

# Expose service_consumers port
EXPOSE 7000:8000

# Run the application
CMD ["/app/main"]