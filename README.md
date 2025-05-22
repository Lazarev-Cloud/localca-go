# LocalCA - Self-Hosted Certificate Authority

[![codecov](https://codecov.io/gh/Lazarev-Cloud/localca-go/graph/badge.svg)](https://codecov.io/gh/Lazarev-Cloud/localca-go)
[![Coverage](https://sonarcloud.io/api/project_badges/measure?project=Lazarev-Cloud_localca-go&metric=coverage)](https://sonarcloud.io/summary/new_code?id=Lazarev-Cloud_localca-go)

LocalCA is a complete solution for running your own Certificate Authority (CA) within a local network environment. It allows you to generate, manage, and deploy SSL/TLS certificates for internal services and clients without relying on external certificate providers.

## Features

- **Complete Certificate Lifecycle Management**: Create, renew, revoke, and delete certificates
- **Web-Based Interface**: User-friendly dashboard for certificate operations
- **Self-Contained CA**: Generate and manage your own root Certificate Authority
- **Server and Client Certificates**: Support for both server (websites, services) and client (user authentication) certificates
- **SAN Support**: Create certificates with multiple domain names (Subject Alternative Names)
- **PKCS#12 Export**: Export client certificates in P12 format for easy browser/device import
- **Certificate Revocation**: Maintain a Certificate Revocation List (CRL)
- **Email Notifications**: Get alerts before certificates expire
- **HTTPS Support**: Secure access to the management interface itself
- **Docker Deployment**: Easy deployment with Docker and Docker Compose
- **ACME Protocol Support**: Automated certificate issuance compatible with standard ACME clients
- **Security Hardened**: Comprehensive security measures including CSRF protection, secure headers, and input validation

## Architecture Overview

LocalCA follows a modern client-server architecture with clear separation of concerns:

### Technology Stack

**Backend (Go)**:
- **Framework**: Gin web framework for HTTP handling
- **Cryptography**: Go's crypto/x509 and crypto/rsa packages
- **Storage**: File-based storage with JSON metadata
- **ACME**: Custom ACME protocol implementation
- **Security**: CSRF protection, secure headers, input validation

**Frontend (Next.js + React)**:
- **Framework**: Next.js 15 with App Router
- **UI Components**: ShadcnUI + Radix UI primitives
- **Styling**: Tailwind CSS with custom design system
- **Forms**: React Hook Form with Zod validation
- **State Management**: React hooks and context

**Infrastructure**:
- **Containerization**: Docker with multi-stage builds
- **Orchestration**: Docker Compose for development
- **Reverse Proxy**: Next.js rewrites for API integration
- **Security**: TLS 1.2+, secure cipher suites

### Service Architecture

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Frontend      │    │   Backend API   │    │   ACME Server   │
│   (Next.js)     │◄──►│   (Go/Gin)      │◄──►│   (Go)          │
│   Port 3000     │    │   Port 8080     │    │   Port 8555     │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                       │                       │
         │                       │                       │
         ▼                       ▼                       ▼
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Browser UI    │    │   Certificate   │    │   ACME Client   │
│   Management    │    │   Storage       │    │   Integration   │
└─────────────────┘    └─────────────────┘    └─────────────────┘
```

## Project Structure

```
├── app/              # Next.js 15 frontend application
│   ├── api/          # API routes and proxy endpoints
│   ├── certificates/ # Certificate management pages
│   ├── create/       # Certificate creation flow
│   ├── login/        # Authentication pages
│   ├── settings/     # Application settings
│   └── setup/        # Initial setup flow
├── build/            # Docker build configurations and test setups
├── components/       # React components (ShadcnUI + Tailwind)
│   ├── ui/           # Base UI components
│   ├── forms/        # Form components
│   └── certificates/ # Certificate-specific components
├── docs/             # Documentation files
├── hooks/            # Custom React hooks
├── pkg/              # Go backend packages
│   ├── acme/         # ACME protocol implementation
│   ├── certificates/ # Certificate management
│   ├── config/       # Configuration management
│   ├── handlers/     # HTTP request handlers
│   ├── security/     # Security utilities
│   └── storage/      # File-based storage
├── public/           # Static assets
├── security/         # Security analysis and SBOM files
├── static/           # Backend static files
├── templates/        # HTML templates for backend
├── tools/            # Development and utility scripts
└── examples/         # Example configurations
```

## Documentation

For comprehensive documentation, see the [docs/](docs/) directory:

- **[Quick Start Guide](docs/README.md)** - Documentation overview and quick links
- **[Deployment Guide](docs/DEPLOYMENT.md)** - Production deployment instructions
- **[Troubleshooting Guide](docs/TROUBLESHOOTING.md)** - Common issues and solutions
- **[Security Best Practices](docs/BestPractice.md)** - Security guidelines and recommendations
- **[Testing Guide](docs/README-TESTING.md)** - Testing procedures and guidelines
- **[Development Guidelines](.cursor/rules/)** - Detailed development documentation

## Security

Security is a top priority for LocalCA. Please review our [Security Policy](SECURITY.md) for:
- Vulnerability reporting procedures
- Security best practices
- Supported versions
- Contact information

## Changelog

See [CHANGELOG.md](CHANGELOG.md) for a detailed history of changes and releases.

## License

This software is licensed under a dual license model:

- **Free for personal, non-commercial self-hosting**
- **Paid license required for commercial, corporate, or organizational use**

See the [LICENSE](LICENSE) file for complete details.

Copyright (c) 2023-2025 lazarevtill (lazarev.cloud)

## Security Considerations

⚠️ **Important Warning**:
- This tool is for **internal use only**. Do not expose it to the public internet.
- Keep your CA private key secure. Anyone with access to it can issue trusted certificates for your domain.
- This project is intended for testing, development environments, and private networks, not for production public-facing services.

## Installation

### Prerequisites

- Go 1.22+ (for building from source)
- Docker and Docker Compose (for container deployment)
- OpenSSL

### Option 1: Docker Deployment (Recommended)

1. Clone the repository:
   ```bash
   git clone https://github.com/Lazarev-Cloud/localca-go.git
   cd localca-go
   ```

2. Start the service:
   ```bash
   docker-compose up -d
   ```

3. Access the web interface at http://localhost:3000

4. Complete the initial setup:
   - On first access, you'll be redirected to the setup page
   - Create an admin account with your desired username and password
   - The setup token will be automatically loaded from the backend

### Option 2: Building from Source

1. Clone the repository:
   ```bash
   git clone https://github.com/Lazarev-Cloud/localca-go.git
   cd localca-go
   ```

2. Download dependencies and build:
   ```bash
   go mod tidy
   go build -o localca-go
   ```

3. Run the application using the provided scripts:
   
   Windows:
   ```bash
   run-dev.bat
   ```
   
   Linux/macOS:
   ```bash
   chmod +x run-dev.sh
   ./run-dev.sh
   ```

5. In a separate terminal, start the frontend:
   ```bash
   npm install --legacy-peer-deps
   npm run dev
   ```

6. Access the web interface at http://localhost:3000

7. Complete the initial setup:
   - On first access, you'll be redirected to the setup page 
   - Create an admin account with your desired username and password
   - Follow the guided setup process

## Configuration

LocalCA is configured through environment variables:

| Variable | Description | Default |
|----------|-------------|---------|
| `CA_NAME` | Name of the Certificate Authority (FQDN) | "LocalCA" |
| `CA_KEY_FILE` | Path to the file containing the CA key password | *optional* |
| `CA_KEY` | Direct CA key password (alternative to CA_KEY_FILE) | "" |
| `ORGANIZATION` | Organization name for the CA | "LocalCA Organization" |
| `COUNTRY` | Country code for the CA | "US" |
| `DATA_DIR` | Path where certificates are stored | "./data" |
| `LISTEN_ADDR` | Address and port for the HTTP server | ":8080" |
| `EMAIL_NOTIFY` | Enable email notifications for expiring certificates | "false" |
| `SMTP_SERVER` | SMTP server for notifications | "" |
| `SMTP_PORT` | SMTP port | "25" |
| `SMTP_USER` | SMTP username | "" |
| `SMTP_PASSWORD` | SMTP password | "" |
| `SMTP_USE_TLS` | Use TLS for SMTP | "false" |
| `EMAIL_FROM` | From address for notification emails | "" |
| `EMAIL_TO` | Default recipient for notification emails | "" |
| `TLS_ENABLED` | Enable HTTPS for the web interface | "false" |
| `ALLOW_LOCALHOST` | Allow loading content from localhost in CSP headers | "false" |
| `NEXT_PUBLIC_API_URL` | URL for the frontend to connect to the backend | "http://localhost:8080" |

## Usage Guide

### Initial Setup

1. **First Access**: Navigate to the web interface at http://localhost:3000
2. **Complete Setup**: 
   - Create an admin account with username and password
   - The setup token is automatically loaded from the backend
   - Complete the guided setup process
3. **Login**: Use your created credentials to log into the dashboard
4. **Trust the CA**: Download the CA certificate and install it in your browser/OS trust store

### Creating Certificates

#### Server Certificates

1. Enter the hostname in the "Common Name" field (e.g., `server.local`)
2. Add any additional domain names separated by commas (e.g., `www.server.local, admin.server.local`)
3. Ensure the "Create client certificate" checkbox is **not** checked
4. Click "Create Certificate"

#### Client Certificates

1. Enter a name for the client in the "Common Name" field (e.g., `john.doe`)
2. Check the "Create client certificate" checkbox
3. Enter a password for the P12 file
4. Click "Create Certificate"
5. Download the P12 file and import it into your browser/client device

### Certificate Management

- **View Certificate Details**: Click on a certificate name in the list
- **Renew Certificate**: Click the "Renew" button next to a certificate
- **Revoke Certificate**: Click the "Revoke" button to add the certificate to the CRL
- **Delete Certificate**: Click the "Delete" button to remove a certificate

### Distribution and Installation

#### Installing the CA Certificate

- **Windows**:
  - Double-click the `ca.pem` file
  - Select "Install Certificate"
  - Choose "Local Machine" and place in "Trusted Root Certification Authorities"

- **macOS**:
  - Double-click the `ca.pem` file
  - Add to your keychain and set to "Always Trust"

- **Linux**:
  - Copy to `/usr/local/share/ca-certificates/`
  - Run `sudo update-ca-certificates`

- **Firefox** (uses its own certificate store):
  - Go to Preferences > Privacy & Security > Certificates > View Certificates
  - Import the CA certificate under "Authorities"

#### Using Certificates with Common Web Servers

**Nginx**:
```nginx
server {
    listen 443 ssl;
    server_name your-domain.local;
    
    ssl_certificate /path/to/your-domain.local.bundle.crt;
    ssl_certificate_key /path/to/your-domain.local.key;
    
    # Other SSL settings
    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_prefer_server_ciphers on;
}
```

**Apache**:
```apache
<VirtualHost *:443>
    ServerName your-domain.local
    
    SSLEngine on
    SSLCertificateFile /path/to/your-domain.local.crt
    SSLCertificateKeyFile /path/to/your-domain.local.key
    SSLCACertificateFile /path/to/ca.pem
    
    # Other settings...
</VirtualHost>
```

## Development

### Project Structure

```
localca-go/
├── main.go                    # Application entry point
├── pkg/                       # Core Go packages
│   ├── certificates/          # Certificate operations
│   ├── config/                # Configuration management
│   ├── email/                 # Email notification system
│   ├── handlers/              # HTTP request handlers
│   ├── storage/               # Certificate storage
│   └── acme/                  # ACME protocol implementation
├── app/                       # Next.js frontend
│   ├── page.tsx               # Main dashboard
│   ├── create/                # Certificate creation
│   ├── certificates/          # Certificate management
│   ├── settings/              # Application settings
│   └── api/                   # API routes
├── components/                # React components
├── Dockerfile                 # Docker build instructions
├── Dockerfile.frontend        # Frontend Docker build
└── docker-compose.yml         # Docker Compose configuration
```

### Local Development

1. Make your changes
2. Run with hot-reload using `air` (optional):
   ```bash
   go install github.com/cosmtrek/air@latest
   air
   ```
3. Or run directly:
   ```bash
   go run main.go
   ```

## Contributing

We welcome contributions to LocalCA! Please read our contribution guidelines before submitting changes.

### Development Setup

1. **Prerequisites**:
   - Go 1.23+ for backend development
   - Node.js 18+ for frontend development
   - Docker for containerized development

2. **Local Development**:
   ```bash
   # Clone the repository
   git clone https://github.com/Lazarev-Cloud/localca-go.git
   cd localca-go
   
   # Backend development
   go mod tidy
   go run main.go
   
   # Frontend development (separate terminal)
   npm install --legacy-peer-deps
   npm run dev
   ```

3. **Development Scripts**:
   - Linux/macOS: `./run-dev.sh`
   - Windows: `run-dev.bat`

### Contribution Process

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Write tests for your changes
4. Ensure all tests pass (`go test ./...` and `npm test`)
5. Commit your changes (`git commit -m 'Add some amazing feature'`)
6. Push to the branch (`git push origin feature/amazing-feature`)
7. Open a Pull Request

### Code Standards

- Follow Go best practices and conventions
- Use TypeScript for all frontend code
- Write comprehensive tests for new features
- Update documentation for user-facing changes
- Follow the existing code style and patterns

For detailed development guidelines, see our [development documentation](docs/).

## Acknowledgments

- [Go](https://golang.org/) - The Go Programming Language
- [Gin](https://gin-gonic.com/) - Web framework for Go
- [Next.js](https://nextjs.org/) - React framework
- [Tailwind CSS](https://tailwindcss.com/) - CSS framework
- [ShadcnUI](https://ui.shadcn.com/) - UI component library
- [OpenSSL](https://www.openssl.org/) - Cryptography and SSL/TLS toolkit

## Roadmap

- [x] ACME protocol support
- [x] Automated certificate renewal
- [x] API for programmatic certificate management
- [ ] OCSP responder
- [ ] Advanced audit logging
- [ ] Certificate transparency log
- [ ] Hardware security module (HSM) support

## Future Development

The following features and improvements are planned for future versions of LocalCA:

### Core Functionality Enhancements
- [ ] OCSP responder for real-time certificate validation
- [ ] Certificate transparency log for enhanced security
- [ ] Hardware security module (HSM) support for key protection
- [ ] ECC certificate support (currently only RSA)
- [ ] Wildcard certificate support
- [ ] Certificate templates for common use cases
- [ ] Support for external CSRs (Certificate Signing Requests)

### Security Improvements
- [ ] Advanced audit logging with searchable history
- [ ] Certificate usage analytics and reporting
- [ ] Role-based access control for multi-user environments
- [ ] Two-factor authentication for administrative access
- [ ] Enhanced key management with rotation policies

### User Experience
- [ ] Improved dashboard with visualization enhancements
- [ ] Batch operations for certificate management
- [ ] Dark mode support
- [ ] Mobile-responsive design
- [ ] Drag-and-drop certificate import
- [ ] Guided wizards for complex operations

### Operational Features
- [ ] Automated backup and restore functionality
- [ ] High availability deployment options
- [ ] Performance optimizations for large certificate stores
- [ ] Metrics and monitoring integration
- [ ] Centralized logging

### Integration Capabilities
- [ ] OpenAPI/Swagger documentation for REST API
- [ ] Webhook notifications for certificate events
- [ ] Integration with popular deployment tools (Ansible, Terraform)
- [ ] Support for cloud provider certificate services
- [ ] LDAP/Active Directory integration

### Email and Notifications
- [ ] Enhanced email templates with HTML formatting
- [ ] Configurable notification thresholds
- [ ] Additional notification channels (Slack, MS Teams, etc.)
- [ ] Calendar integration for expiration events

We welcome contributions to any of these areas. If you're interested in working on a feature, please open an issue to discuss implementation details before submitting a pull request.

## Running Tests

LocalCA-Go includes a comprehensive test suite for all packages. You can run the tests using the provided scripts.

### On Linux/macOS
```bash
./run-tests.sh
```

### On Windows
```batch
.\run-tests.bat
```

The test scripts will:
1. Run all package tests with coverage
2. Run the main package test
3. Test Docker build if Docker is available
4. Test Docker Compose configuration if available

## Docker Deployment

You can easily deploy LocalCA using Docker and Docker Compose.

### Building and Running with Docker Compose

```bash
# Build and start all services
docker-compose up -d

# View logs
docker-compose logs -f

# Stop all services
docker-compose down
```

### Accessing the Application

- Web UI: http://localhost:3000
- Backend API: http://localhost:8080
- ACME Server: http://localhost:8555

### Docker Volumes

The application uses a Docker volume named `localca-data` to persist certificate data. This ensures that your certificates and CA information are preserved even if the containers are removed.

### Custom Configuration

You can customize the application by setting environment variables in the docker-compose.yml file:

```yaml
environment:
  - LOCALCA_DATA_DIR=/app/data
  - LOCALCA_HOST=0.0.0.0
  - CA_NAME=My Custom CA
  - CA_KEY=mysecretpassword
  # Add more environment variables as needed
```

For a complete list of configuration options, see the Configuration section above.

## Supply Chain Security

This project implements SLSA (Supply chain Levels for Software Artifacts) Level 3 build security. The artifacts have cryptographically signed attestations that provide provenance and integrity guarantees.

### Verifying Binary Attestations

To verify the binary attestations, you can use the GitHub CLI:

```bash
# Install GitHub CLI if not already installed
# https://cli.github.com/manual/installation

# Verify the binary
gh attestation verify localca-go -R lazarev-cloud/localca-go
  
# Verify the SBOM
gh attestation verify localca-go -R lazarev-cloud/localca-go --predicate-type https://spdx.dev/Document/v2.3
```

### Verifying Container Attestations

To verify the container attestations:

```bash
# Login to GitHub Container Registry
docker login ghcr.io
  
# Verify backend container
gh attestation verify oci://ghcr.io/lazarev-cloud/localca-go/backend:latest -R lazarev-cloud/localca-go
  
# Verify frontend container
gh attestation verify oci://ghcr.io/lazarev-cloud/localca-go/frontend:latest -R lazarev-cloud/localca-go
```

These verifications ensure the software you're using was built from the source code in this repository using GitHub Actions secure builder workflows.

## Software Bill of Materials (SBOM)

LocalCA-Go includes a Software Bill of Materials (SBOM) in [SPDX format](https://spdx.dev/). The SBOM provides a comprehensive inventory of all components used in the software, including dependencies, licenses, and other metadata.

### SBOM Files

- `SPDX.json`: The main SBOM file in SPDX 2.3 format
- `merged-sbom.json`: Generated during CI/CD, combines backend and frontend component information
- `examples/spdx-example.json`: Example SPDX document showing the structure and format

### Automated SBOM Generation

The SBOM is automatically generated and updated through our CI/CD pipeline using the [anchore/sbom-action](https://github.com/anchore/sbom-action) GitHub Action. This ensures the SBOM stays current with the project's dependencies.

For more information about SPDX and examples, see the [examples/README.md](examples/README.md) file.

---

Created by [@lazarevtill](https://github.com/lazarevtill) - feel free to contact me!
