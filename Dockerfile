FROM golang:1.23.8-alpine3.20 AS builder

WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Fix go.sum entries before proceeding
RUN go mod tidy && go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o localca-go .

# Use a smaller base image for the final container
FROM alpine:3.20

# Install required packages
RUN apk add --no-cache \
    openssl \
    ca-certificates \
    tzdata \
    curl \
    bash

# Create app directory
WORKDIR /app

# Copy the binary from builder
COPY --from=builder /app/localca-go /app/

# Copy templates and static files
COPY --from=builder /app/templates /app/templates
COPY --from=builder /app/static /app/static

# Create directory for certificates
RUN mkdir -p /app/certs/ca

# Set environment variables
ENV GIN_MODE=release

# Expose port
EXPOSE 8080
EXPOSE 8443

# Command to run
CMD ["/app/localca-go"]