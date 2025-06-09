# LocalCA: Self-Hosted Certificate Authority

[![Docker](https://img.shields.io/badge/Docker-Enabled-blue.svg)](https://docker.com)
[![Go](https://img.shields.io/badge/Go-1.23+-00ADD8.svg)](https://golang.org)
[![Next.js](https://img.shields.io/badge/Next.js-15.0+-000000.svg)](https://nextjs.org)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)

LocalCA is a complete solution for managing a private Certificate Authority within a local network. The project allows you to generate, manage, and deploy SSL/TLS certificates for internal services and clients with a modern web-based interface.

## ✨ Features

### Core Certificate Management
- ✅ **CA Management**: Create and manage your own Certificate Authority
- ✅ **Server Certificates**: Generate SSL/TLS certificates for web servers
- ✅ **Client Certificates**: Create certificates for client authentication
- ✅ **Certificate Revocation**: Revoke compromised certificates with CRL support
- ✅ **Certificate Renewal**: Renew certificates before expiration
- ✅ **PKCS#12 Export**: Export client certificates with private keys

### Enhanced Storage & Performance
- ✅ **Multi-Backend Storage**: File, PostgreSQL, and S3/MinIO support
- ✅ **Caching Layer**: Redis/KeyDB caching for improved performance
- ✅ **Audit Logging**: Comprehensive audit trail for compliance
- ✅ **Backup Support**: Automated backup and recovery capabilities
- ✅ **Data Encryption**: Encrypted sensitive data storage

### Web Interface & API
- ✅ **Modern Dashboard**: React-based responsive web interface with real-time statistics
- ✅ **Certificate Management**: Create, view, renew, and revoke certificates
- ✅ **Advanced Filtering**: Filter certificates by type, status, and search
- ✅ **Dark/Light Theme**: Configurable UI themes with system preference detection
- ✅ **Mobile Responsive**: Works on all device sizes
- ✅ **REST API**: Complete API for all certificate operations

### Security & Authentication
- ✅ **Secure Authentication**: Session-based authentication with CSRF protection
- ✅ **Initial Setup**: Secure setup process with time-limited tokens
- ✅ **Password Protection**: CA private key protection with secure storage
- ✅ **Session Management**: Secure session handling with HTTP-only cookies
- ✅ **Security Headers**: Comprehensive security headers and middleware

### Automation & Integration
- ✅ **ACME Protocol**: Automated certificate issuance (experimental)
- ✅ **Email Notifications**: Certificate expiration alerts
- ✅ **JSON Logging**: Structured logging for monitoring and alerting
- ✅ **Health Checks**: Service health monitoring and status endpoints

## 🚀 Quick Start

### Prerequisites

- **Docker & Docker Compose** (recommended)
- **Go 1.23+** (for local development)
- **Node.js 18+** (for frontend development)

### Docker Deployment (Recommended)

1. **Clone the repository**:
```bash
git clone https://github.com/Lazarev-Cloud/localca-go.git
cd localca-go
```

2. **Start with Docker Compose**:
```bash
# For production with enhanced storage
docker-compose up -d

# For development
docker-compose -f docker-compose.dev.yml up -d
```

3. **Access the application**:
- **Frontend**: http://localhost:3000
- **Backend API**: http://localhost:8080
- **MinIO Console**: http://localhost:9001 (admin/password123)

### Standalone Deployment

1. **Build the backend**:
```bash
go build -o localca-go
```

2. **Build the frontend**:
```bash
npm install
npm run build
```

3. **Run the application**:
```bash
./localca-go
```

## 📁 Project Structure

```
localca-go/
├── app/                      # Next.js app directory
│   ├── api/                 # API routes and proxy
│   │   ├── proxy/[...path]/ # Dynamic API proxy to backend
│   │   ├── ca-info/         # CA information endpoint
│   │   ├── certificates/    # Certificate management endpoints
│   │   ├── login/           # Authentication endpoints
│   │   └── setup/           # Initial setup endpoints
│   ├── certificates/        # Certificate management pages
│   │   └── [id]/           # Individual certificate details
│   ├── create/             # Certificate creation wizard
│   ├── login/              # Authentication pages
│   ├── settings/           # Settings and configuration
│   └── setup/              # Initial application setup
├── components/              # React components
│   ├── ui/                 # Base UI components (shadcn/ui)
│   ├── certificate-*.tsx  # Certificate-related components
│   ├── dashboard-*.tsx    # Dashboard components
│   └── system-status.tsx  # System monitoring components
├── hooks/                   # React hooks
│   ├── use-api.ts          # Generic API client hook
│   ├── use-certificates.ts # Certificate management hook
│   ├── use-auth.ts         # Authentication hook
│   └── use-*.ts           # Other utility hooks
├── lib/                     # Utility libraries
│   ├── config.ts           # Frontend configuration
│   └── utils.ts            # Utility functions
├── pkg/                     # Go backend packages
│   ├── acme/               # ACME protocol implementation
│   ├── cache/              # Redis/KeyDB caching layer
│   ├── certificates/       # Certificate operations
│   ├── config/             # Configuration management
│   ├── database/           # PostgreSQL integration
│   ├── email/              # Email notifications
│   ├── handlers/           # HTTP handlers and routing
│   ├── logging/            # Structured logging
│   ├── s3storage/          # S3/MinIO object storage
│   ├── security/           # Security utilities
│   └── storage/            # Storage backends and interfaces
├── docs/                    # Documentation
│   ├── deployment/         # Deployment guides
│   ├── development/        # Development documentation
│   └── security/           # Security documentation
├── .cursor/                 # Cursor AI rules and configuration
│   └── rules/              # Comprehensive project rules
├── tools/                   # Utility scripts and tools
├── docker-compose.yml       # Production Docker setup
├── Dockerfile              # Backend container
├── Dockerfile.frontend     # Frontend container
└── main.go                 # Application entry point
```

## 🔧 Configuration

### Environment Variables

The application is configured through environment variables:

| Variable | Description | Default | Status |
|----------|-------------|---------|--------|
| **Core Configuration** |
| `CA_NAME` | Certificate Authority name | "LocalCA" | ✅ Working |
| `CA_KEY_PASSWORD` | CA private key password | *required* | ✅ Working |
| `ORGANIZATION` | Organization name | "LocalCA Organization" | ✅ Working |
| `COUNTRY` | Country code | "US" | ✅ Working |
| `DATA_DIR` | Data storage directory | "./data" | ✅ Working |
| `LISTEN_ADDR` | HTTP server address | ":8080" | ✅ Working |
| **Security Configuration** |
| `TLS_ENABLED` | Enable HTTPS | "false" | ✅ Working |
| `SESSION_SECRET` | Session encryption key | *auto-generated* | ✅ Working |
| **Enhanced Storage** |
| `DATABASE_ENABLED` | Enable PostgreSQL storage | "false" | ✅ Working |
| `DATABASE_URL` | PostgreSQL connection string | *optional* | ✅ Working |
| `S3_ENABLED` | Enable S3/MinIO storage | "false" | ✅ Working |
| `S3_ENDPOINT` | S3 endpoint URL | *optional* | ✅ Working |
| `CACHE_ENABLED` | Enable Redis/KeyDB caching | "false" | ✅ Working |
| `REDIS_URL` | Redis connection URL | *optional* | ✅ Working |
| **Notifications** |
| `EMAIL_NOTIFY` | Enable email notifications | "false" | ✅ Working |
| `SMTP_HOST` | SMTP server hostname | *optional* | ✅ Working |
| `SMTP_PORT` | SMTP server port | "587" | ✅ Working |
| **Logging** |
| `LOG_FORMAT` | Logging format (json/text) | "text" | ✅ Working |
| `LOG_LEVEL` | Logging level | "info" | ✅ Working |
| **Frontend** |
| `NEXT_PUBLIC_API_URL` | Frontend API URL | "http://localhost:8080" | ✅ Working |

### Docker Environment

For Docker deployments, copy `.env.example` to `.env` and modify as needed:

```bash
cp .env.example .env
# Edit .env with your preferred settings
```

Example `.env` configuration:
```bash
# Core Configuration
CA_NAME=MyLocalCA
CA_KEY_PASSWORD=secure-ca-password
ORGANIZATION=My Organization
COUNTRY=US

# Enhanced Storage
DATABASE_ENABLED=true
DATABASE_URL=postgres://localca:localca_password@postgres:5432/localca
S3_ENABLED=true
S3_ENDPOINT=http://minio:9000
S3_ACCESS_KEY=minioadmin
S3_SECRET_KEY=minioadmin
CACHE_ENABLED=true
REDIS_URL=redis://keydb:6379

# Email Notifications
EMAIL_NOTIFY=true
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USERNAME=your-email@gmail.com
SMTP_PASSWORD=your-app-password
```

## 🎯 Function Status

### ✅ Fully Working Features

#### 1. Certificate Management
- **CA Creation**: Automatic CA certificate generation with secure key storage
- **Server Certificates**: SSL/TLS certificates for web servers with SAN support
- **Client Certificates**: Client authentication certificates with PKCS#12 export
- **Certificate Revocation**: CRL generation and certificate revocation
- **Certificate Renewal**: Automated and manual certificate renewal
- **Certificate Validation**: X.509 certificate chain validation

#### 2. Enhanced Storage System
- **Multi-Backend Storage**: File, PostgreSQL, and S3/MinIO storage backends
- **Caching Layer**: Redis/KeyDB caching for improved performance
- **Audit Logging**: Comprehensive audit trail for compliance
- **Backup & Recovery**: Automated backup and disaster recovery
- **Health Monitoring**: Storage backend health checks and monitoring

#### 3. Web Interface
- **Modern Dashboard**: Real-time system statistics and certificate overview
- **Certificate Management**: Complete certificate lifecycle management
- **Advanced Search**: Filter and search certificates by multiple criteria
- **Responsive Design**: Mobile-first responsive design
- **Theme Support**: Dark/light theme with system preference detection

#### 4. API & Integration
- **REST API**: Complete RESTful API for all operations
- **API Proxy**: Next.js API proxy for seamless frontend integration
- **Authentication**: Secure session-based authentication
- **CSRF Protection**: Cross-site request forgery prevention
- **Rate Limiting**: Built-in rate limiting for security

#### 5. Security Features
- **Secure Authentication**: Password hashing with bcrypt
- **Session Management**: HTTP-only cookies with secure attributes
- **Security Headers**: Comprehensive HTTP security headers
- **Input Validation**: Server-side input validation and sanitization
- **TLS Configuration**: Modern TLS 1.2/1.3 configuration

#### 6. Monitoring & Logging
- **Structured Logging**: JSON and text logging formats
- **Performance Metrics**: System and application performance monitoring
- **Health Checks**: Service health and readiness endpoints
- **Error Tracking**: Comprehensive error logging and alerting

### 🚧 Experimental Features

#### 1. ACME Protocol
- **Basic ACME Server**: ACME protocol implementation for automated certificate issuance
- **HTTP-01 Challenge**: Web-based domain validation
- **Account Management**: ACME account creation and management
- **Order Processing**: Certificate order lifecycle management

*Note: ACME implementation is experimental and may require additional testing with real ACME clients.*

#### 2. Email Notifications
- **SMTP Integration**: Email notifications for certificate expiration
- **Template System**: HTML and text email templates
- **Batch Processing**: Efficient batch email processing

*Note: Email system is functional but templates and scheduling may need enhancement.*

### 🔄 Recently Enhanced

#### 1. Storage Architecture
- **Multi-Backend Support**: Added PostgreSQL and S3/MinIO storage backends
- **Caching Integration**: Implemented Redis/KeyDB caching layer
- **Performance Optimization**: Improved storage operation performance
- **Backup Capabilities**: Added automated backup and recovery features

#### 2. Frontend Improvements
- **Real-Time Data**: Replaced mock data with real API integration
- **Enhanced UI**: Improved user interface with modern components
- **Performance**: Optimized frontend performance and loading times
- **Error Handling**: Better error handling and user feedback

#### 3. Security Enhancements
- **Authentication System**: Improved session management and security
- **CSRF Protection**: Enhanced CSRF protection implementation
- **Input Validation**: Comprehensive input validation and sanitization
- **Security Headers**: Added comprehensive security headers

## 🧪 Testing

### Run All Tests

```bash
# Backend tests
go test ./...

# Frontend tests
npm test

# Docker-based testing
docker-compose -f docker-compose.test.yml up --abort-on-container-exit
```

### Enhanced Storage Tests

```bash
# Test enhanced storage features
./tools/test-enhanced-storage.sh

# Comprehensive system validation
./tools/comprehensive-enhanced-test.sh
```

### Application Tests

```bash
# Basic functionality tests
./tools/test_application.sh

# ACME protocol tests
./tools/test-acme.sh
```

## 📈 Monitoring & Management

### Health Checks

```bash
# Check deployment status
./tools/deployment-status.sh

# Check service health
curl http://localhost:8080/api/health
curl http://localhost:3000/api/health
```

### Performance Metrics

The application provides comprehensive performance metrics:

- **System Statistics**: Real-time system metrics via `/api/statistics`
- **Cache Performance**: Cache hit rates and performance metrics
- **Storage Metrics**: Storage usage and performance monitoring
- **Certificate Statistics**: Certificate counts and status overview

### Logging and Monitoring

Structured logging is available in multiple formats:

```bash
# View backend logs
docker-compose logs backend

# View frontend logs
docker-compose logs frontend

# View all service logs
docker-compose logs -f
```

## 🔐 Security

### Security Features

- **Authentication**: Session-based authentication with secure cookies
- **CSRF Protection**: Built-in CSRF token validation
- **Security Headers**: Comprehensive HTTP security headers
- **Input Validation**: Server-side input validation and sanitization
- **Rate Limiting**: Built-in rate limiting for API endpoints
- **Audit Logging**: Complete audit trail for all operations

### Security Best Practices

1. **Change default passwords** before production use
2. **Enable HTTPS** for production deployments using `TLS_ENABLED=true`
3. **Use strong CA key passwords** with `CA_KEY_PASSWORD`
4. **Regular certificate rotation** and monitoring
5. **Monitor audit logs** for security events
6. **Keep software updated** with latest security patches

### Security Configuration

```bash
# Enable TLS for production
TLS_ENABLED=true
TLS_CERT_FILE=/path/to/cert.pem
TLS_KEY_FILE=/path/to/key.pem

# Configure secure session settings
SESSION_SECRET=your-secure-session-secret
SESSION_TIMEOUT=3600

# Enable audit logging
AUDIT_ENABLED=true
AUDIT_LOG_FILE=/var/log/localca/audit.log
```

## 📚 Documentation

### Comprehensive Documentation
- **[Project Overview](.cursor/rules/01-project-overview.mdc)**: Complete project overview and architecture
- **[Backend Architecture](.cursor/rules/02-backend-architecture.mdc)**: Go backend implementation details
- **[Frontend Architecture](.cursor/rules/03-frontend-architecture.mdc)**: Next.js frontend implementation
- **[Enhanced Storage](.cursor/rules/12-enhanced-storage-caching.mdc)**: Multi-backend storage and caching
- **[API Integration](.cursor/rules/13-api-integration-patterns.mdc)**: API endpoints and integration patterns

### Deployment & Operations
- **[Deployment Guide](docs/deployment/SETUP_DATABASE_S3.md)**: Enhanced storage setup and deployment
- **[Development Guide](docs/development/CACHING.md)**: Caching and performance optimization
- **[Docker Setup](docs/DEPLOYMENT.md)**: Docker deployment and configuration

### Security & Compliance
- **[Security Guide](docs/security/SECURITY.md)**: Security best practices and guidelines
- **[Security Review](docs/security/SECURITY_REVIEW.md)**: Comprehensive security assessment
- **[Best Practices](docs/BestPractice.md)**: Operational best practices

### Development & Troubleshooting
- **[Development Guide](.cursor/rules/09-development-guide.mdc)**: Development standards and workflow
- **[Troubleshooting](docs/TROUBLESHOOTING.md)**: Common issues and solutions
- **[Changelog](docs/CHANGELOG.md)**: Version history and changes

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch: `git checkout -b feature/amazing-feature`
3. Make your changes following the development guidelines
4. Add tests if applicable
5. Commit your changes: `git commit -m 'Add amazing feature'`
6. Push to the branch: `git push origin feature/amazing-feature`
7. Submit a pull request

### Development Guidelines

- Follow the coding standards in [Development Guide](.cursor/rules/09-development-guide.mdc)
- Write tests for new features
- Update documentation as needed
- Ensure all tests pass before submitting PR

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🆘 Support

- **Issues**: [GitHub Issues](https://github.com/Lazarev-Cloud/localca-go/issues)
- **Documentation**: [docs/](docs/) and [.cursor/rules/](.cursor/rules/)
- **Security**: See [SECURITY.md](docs/security/SECURITY.md) for security policy

## 🎉 Acknowledgments

- Built with [Go](https://golang.org/) and [Gin](https://gin-gonic.com/)
- Frontend powered by [Next.js](https://nextjs.org/) and [React](https://reactjs.org/)
- UI components from [shadcn/ui](https://ui.shadcn.com/)
- Enhanced storage with [PostgreSQL](https://postgresql.org/) and [MinIO](https://min.io/)
- Caching with [KeyDB](https://keydb.dev/)
- Containerization with [Docker](https://docker.com/)

---

**LocalCA** - Self-hosted Certificate Authority for modern applications 🔒

*Secure, scalable, and easy to deploy certificate management solution.*
