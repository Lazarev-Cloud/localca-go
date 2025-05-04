# Dockerfile
FROM golang:1.22-alpine AS builder

WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Install required tools
RUN apk add --no-cache git

# Fix go.sum entries before proceeding
RUN go mod tidy && go mod download

# Copy source code
COPY . .

# Build the application with security flags
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-w -s" -o localca-go .

# Use a smaller base image for the final container
FROM alpine:3.18

# Install required packages
RUN apk add --no-cache \
    openssl \
    ca-certificates \
    tzdata \
    curl \
    bash 

# Create a non-root user
RUN addgroup -S localca && adduser -S -G localca localca

# Create directories with appropriate permissions
RUN mkdir -p /app/certs/ca && \
    mkdir -p /app/data && \
    chown -R localca:localca /app

WORKDIR /app

# Copy the binary from builder
COPY --from=builder /app/localca-go /app/

# Copy templates and static files
COPY --from=builder /app/templates /app/templates
COPY --from=builder /app/static /app/static

# Create directory for certificates
RUN mkdir -p /app/certs/ca && chown -R localca:localca /app/certs

# Set environment variables
ENV GIN_MODE=release

# Switch to non-root user
USER localca

# Expose ports
EXPOSE 8080
EXPOSE 8443

# Command to run
CMD ["/app/localca-go"]