# LocalCA: Self-Hosted Certificate Authority

[![Docker](https://img.shields.io/badge/Docker-Enabled-blue.svg)](https://docker.com)
[![Go](https://img.shields.io/badge/Go-1.21+-00ADD8.svg)](https://golang.org)
[![Next.js](https://img.shields.io/badge/Next.js-14.0+-000000.svg)](https://nextjs.org)
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

### Web Interface
- ✅ **Modern Dashboard**: React-based responsive web interface
- ✅ **Real-time Statistics**: Live system and certificate statistics
- ✅ **Certificate Management**: Create, view, renew, and revoke certificates
- ✅ **Advanced Filtering**: Filter certificates by type, status, and search
- ✅ **Dark/Light Theme**: Configurable UI themes
- ✅ **Mobile Responsive**: Works on all device sizes

### Security & Authentication
- ✅ **Secure Authentication**: Session-based authentication with CSRF protection
- ✅ **Initial Setup**: Secure setup process with time-limited tokens
- ✅ **Password Protection**: CA private key protection
- ✅ **Session Management**: Secure session handling
- ✅ **Security Headers**: Comprehensive security headers

### Storage & Performance
- ✅ **Enhanced Storage**: Multi-backend storage (File, PostgreSQL, S3)
- ✅ **Caching Layer**: Redis/KeyDB caching for improved performance
- ✅ **Audit Logging**: Comprehensive audit trail
- ✅ **Backup Support**: Automated backup capabilities
- ✅ **Data Encryption**: Encrypted sensitive data storage

### API & Automation
- ✅ **REST API**: Complete API for all certificate operations
- ✅ **ACME Protocol**: Automated certificate issuance (experimental)
- ✅ **Email Notifications**: Certificate expiration alerts
- ✅ **JSON Logging**: Structured logging for monitoring

## 🚀 Quick Start

### Prerequisites

- **Docker & Docker Compose** (recommended)
- **Go 1.21+** (for local development)
- **Node.js 18+** (for frontend development)

### Docker Deployment (Recommended)

1. **Clone the repository**:
```bash
git clone https://github.com/your-username/localca-go.git
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
- **MinIO Console**: http://localhost:9001

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
│   ├── certificates/        # Certificate management pages
│   ├── create/             # Certificate creation page
│   ├── login/              # Authentication pages
│   ├── settings/           # Settings page
│   └── setup/              # Initial setup page
├── components/              # React components
│   ├── ui/                 # Base UI components (shadcn/ui)
│   ├── certificate-*.tsx  # Certificate-related components
│   ├── system-status.tsx  # System statistics dashboard
│   └── dashboard-*.tsx    # Dashboard components
├── hooks/                   # React hooks
│   ├── use-api.ts          # API interaction hook
│   ├── use-certificates.ts # Certificate management hook
│   └── use-*.ts           # Other utility hooks
├── lib/                     # Utility libraries
│   ├── config.ts           # Frontend configuration
│   └── utils.ts            # Utility functions
├── pkg/                     # Go backend packages
│   ├── acme/               # ACME protocol implementation
│   ├── cache/              # Caching layer
│   ├── certificates/       # Certificate operations
│   ├── config/             # Configuration management
│   ├── email/              # Email notifications
│   ├── handlers/           # HTTP handlers
│   ├── logging/            # Structured logging
│   ├── security/           # Security utilities
│   └── storage/            # Storage backends
├── docs/                    # Documentation
│   ├── deployment/         # Deployment guides
│   ├── development/        # Development documentation
│   ├── guides/             # User guides
│   └── security/           # Security documentation
├── tools/                   # Utility scripts
├── static/                  # Static assets
├── templates/               # HTML templates
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
| `CA_NAME` | Certificate Authority name | "LocalCA" | ✅ Working |
| `CA_KEY_PASSWORD` | CA private key password | *required* | ✅ Working |
| `ORGANIZATION` | Organization name | "LocalCA Organization" | ✅ Working |
| `COUNTRY` | Country code | "US" | ✅ Working |
| `DATA_DIR` | Data storage directory | "./data" | ✅ Working |
| `LISTEN_ADDR` | HTTP server address | ":8080" | ✅ Working |
| `TLS_ENABLED` | Enable HTTPS | "false" | ✅ Working |
| `EMAIL_NOTIFY` | Enable email notifications | "false" | ✅ Working |
| `SMTP_*` | SMTP configuration | *various* | ✅ Working |
| `DATABASE_ENABLED` | Enable PostgreSQL storage | "false" | ✅ Working |
| `S3_ENABLED` | Enable S3 storage | "false" | ✅ Working |
| `CACHE_ENABLED` | Enable caching | "false" | ✅ Working |
| `LOG_FORMAT` | Logging format (json/text) | "text" | ✅ Working |
| `LOG_LEVEL` | Logging level | "info" | ✅ Working |
| `NEXT_PUBLIC_API_URL` | Frontend API URL | "http://localhost:8080" | ✅ Working |

### Docker Environment

For Docker deployments, copy `.env.example` to `.env` and modify as needed:

```bash
cp .env.example .env
# Edit .env with your preferred settings
```

## 🎯 Function Status

### ✅ Fully Working Features

1. **Certificate Management**
   - CA creation and management
   - Server certificate generation
   - Client certificate generation
   - Certificate revocation with CRL
   - Certificate renewal
   - PKCS#12 export

2. **Web Interface**
   - Dashboard with real-time statistics
   - Certificate listing and filtering
   - Certificate creation wizard
   - Settings management
   - Authentication system

3. **API Endpoints**
   - `/api/certificates` - Certificate CRUD operations
   - `/api/ca-info` - CA information
   - `/api/statistics` - System statistics
   - `/api/settings` - Configuration management
   - `/api/auth/*` - Authentication endpoints

4. **Security**
   - CSRF protection
   - Session management
   - Password hashing
   - Secure headers

5. **Storage & Performance**
   - File-based storage
   - PostgreSQL integration
   - S3/MinIO object storage
   - Redis/KeyDB caching
   - Audit logging

### 🚧 Partially Working Features

1. **ACME Protocol**
   - Basic ACME server implementation
   - Needs testing with real ACME clients
   - Some challenges may not work properly

2. **Email Notifications**
   - SMTP configuration working
   - Template system needs improvement
   - Scheduling system needs enhancement

### 🔄 Recently Fixed

1. **API Proxy Issues**
   - Fixed 404 errors in API proxy endpoints
   - Added missing statistics endpoint
   - Improved error handling

2. **Frontend Data Loading**
   - Replaced hardcoded mock data with real API calls
   - Fixed system status component
   - Improved certificate statistics

3. **File Organization**
   - Moved documentation to `docs/` directory
   - Cleaned up temporary files
   - Updated gitignore patterns

## 🧪 Testing

### Run All Tests

```bash
# Backend tests
go test ./...

# Frontend tests
npm test

# Docker-based testing
./tools/run-tests-docker.sh
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
```

## 📈 Monitoring & Management

### Health Checks

```bash
# Check deployment status
./tools/deployment-status.sh

# Check service health
curl http://localhost:8080/api/auth/status
curl http://localhost:3000/api/health
```

### Performance Metrics

The application provides performance metrics through:

- **System Statistics API**: Real-time system metrics
- **Cache Statistics**: Cache hit rates and performance
- **Storage Metrics**: Storage usage and performance
- **Certificate Statistics**: Certificate counts and status

### Logging

Structured logging is available in JSON format:

```bash
# View backend logs
docker-compose logs backend

# View all service logs
docker-compose logs -f
```

## 🔐 Security

### Security Features

- **Authentication**: Session-based with secure cookies
- **CSRF Protection**: Built-in CSRF token validation
- **Security Headers**: Comprehensive security headers
- **Input Validation**: Server-side input validation
- **Rate Limiting**: Built-in rate limiting (ACME)
- **Audit Logging**: Complete audit trail

### Security Best Practices

1. **Change default passwords** before production use
2. **Enable HTTPS** for production deployments
3. **Use strong CA key passwords**
4. **Regular certificate rotation**
5. **Monitor audit logs**
6. **Keep software updated**

## 📚 Documentation

- **[Deployment Guide](docs/deployment/SETUP_DATABASE_S3.md)**: Enhanced storage setup
- **[Development Guide](docs/development/CACHING.md)**: Caching and performance
- **[Security Guide](docs/security/SECURITY.md)**: Security best practices
- **[Security Review](docs/security/SECURITY_REVIEW.md)**: Security assessment
- **[Changelog](docs/CHANGELOG.md)**: Version history

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests if applicable
5. Submit a pull request

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🆘 Support

- **Issues**: [GitHub Issues](https://github.com/your-username/localca-go/issues)
- **Documentation**: [docs/](docs/)
- **Security**: See [SECURITY.md](docs/security/SECURITY.md) for security policy

## 🎉 Acknowledgments

- Built with [Go](https://golang.org/) and [Gin](https://gin-gonic.com/)
- Frontend powered by [Next.js](https://nextjs.org/) and [React](https://reactjs.org/)
- UI components from [shadcn/ui](https://ui.shadcn.com/)
- Enhanced storage with [PostgreSQL](https://postgresql.org/) and [MinIO](https://min.io/)
- Caching with [KeyDB](https://keydb.dev/)

---

**LocalCA** - Self-hosted Certificate Authority for modern applications 🔒
