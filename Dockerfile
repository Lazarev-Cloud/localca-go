FROM golang:1.23-alpine AS builder

WORKDIR /app

# Copy go.mod and go.sum first for better caching
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the application
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o localca-go

# Use a smaller image for the final container
FROM alpine:latest

# Install CA certificates for HTTPS connections
RUN apk --no-cache add ca-certificates

WORKDIR /app

# Copy the binary from the builder stage
COPY --from=builder /app/localca-go .

# Copy static files
COPY --from=builder /app/static /app/static
COPY --from=builder /app/templates /app/templates

# Create data directory
RUN mkdir -p /app/data

# Expose ports
EXPOSE 8080 8443 8555

# Set environment variables
ENV LOCALCA_DATA_DIR=/app/data

# Run the application
CMD ["./localca-go"]