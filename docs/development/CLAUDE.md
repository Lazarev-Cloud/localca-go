# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Essential Commands

### Development

```bash
# Go backend (starts on :8080)
go run main.go

# Next.js frontend (starts on :3000)
npm run dev

# Docker development
docker-compose up -d

# Local development scripts
./tools/run-dev.sh    # Linux/macOS
tools/run-dev.bat     # Windows
```

### Testing

```bash
# Run all Go tests with coverage
./tools/run-tests.sh  # Linux/macOS
tools/run-tests.bat   # Windows

# Go tests only
go test -v -cover ./pkg/...

# Frontend tests (Jest)
npm test

# Run Docker tests
./tools/run-tests-docker.sh
```

### Build & Lint

```bash
# Build Go backend
go build -o localca-go

# Frontend build
npm run build
npm run lint

# Docker build
docker build -t localca-go-backend .
```

## Architecture Overview

LocalCA is a dual-service application: Go backend with Gin + Next.js frontend.

### Go Backend Structure

- **main.go**: Entry point, starts HTTP server (:8080), HTTPS server (:8443), and ACME server (:8555)
- **pkg/certificates/**: Core CA and certificate operations (create, renew, revoke)
- **pkg/handlers/**: HTTP handlers for both web UI and API endpoints
- **pkg/storage/**: File-based certificate storage management
- **pkg/acme/**: ACME protocol implementation for automated certificate issuance
- **pkg/config/**: Environment-based configuration management
- **pkg/email/**: Certificate expiration notifications

### Next.js Frontend Structure

- **app/**: Next.js 13+ app router structure
- **app/api/**: API route handlers that proxy to Go backend
- **components/**: React components using ShadcnUI + Tailwind
- **hooks/**: Custom React hooks for API calls and state management

### Key Integration Points

- Frontend proxies API calls through `/api/proxy/[...path]` to Go backend
- Authentication via session cookies managed by Go backend
- CSRF protection for all non-API routes
- Setup flow: `/setup` → `/login` → dashboard

### Storage Structure

- **data/**: Certificate storage directory
  - `ca.pem`, `ca-key.pem`: Root CA certificate and private key
  - `certificates/`: Individual certificate files
  - `revoked/`: Revoked certificates for CRL
- Configuration via environment variables or `cakey.txt` file

### Security Features

- CSRF tokens for web forms
- Security headers middleware
- TLS configuration with secure ciphers
- Session-based authentication
- Input validation and sanitization

## Development Notes

### Environment Variables

Key variables: `CA_KEY_FILE`, `DATA_DIR`, `LISTEN_ADDR`, `TLS_ENABLED`, `NEXT_PUBLIC_API_URL`

### Package Dependencies

- Go: Gin framework, golang.org/x/crypto for cryptographic operations
- Frontend: Next.js 15, ShadcnUI components, Tailwind CSS, React Hook Form with Zod validation

### Test Structure

- Go: Unit tests for each package in `*_test.go` files
- Frontend: Jest configuration with jsdom environment
- Docker: Automated build testing in CI/CD pipeline

### Dual-License Model

Free for personal use, paid license required for commercial use.