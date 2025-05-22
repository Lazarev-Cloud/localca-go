# Changelog

All notable changes to the LocalCA project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- Comprehensive Cursor Rules for development guidance
- Enhanced documentation structure and organization
- Improved README with architecture overview and technology stack details
- Troubleshooting and debugging guide
- API endpoints documentation
- Development guide with coding standards

### Changed
- Updated project documentation structure
- Enhanced contribution guidelines with development setup
- Improved docs/README.md with comprehensive documentation index

### Security
- Enhanced security documentation and best practices

## [1.0.0] - 2024-01-XX

### Added
- Complete Certificate Authority (CA) implementation
- Web-based certificate management interface
- ACME protocol support for automated certificate issuance
- Support for both server and client certificates
- Subject Alternative Name (SAN) support
- PKCS#12 export for client certificates
- Certificate Revocation List (CRL) management
- Email notifications for expiring certificates
- Docker deployment with Docker Compose
- Comprehensive test suite
- Security hardening with CSRF protection and secure headers
- Next.js 15 frontend with ShadcnUI components
- Go backend with Gin web framework
- File-based storage system
- Authentication and session management
- Initial setup wizard
- Certificate lifecycle management (create, renew, revoke, delete)

### Security
- TLS 1.2+ support with secure cipher suites
- Input validation and sanitization
- Secure session management
- CSRF protection
- Security headers implementation
- Private key protection

### Infrastructure
- Multi-stage Docker builds
- GitHub Actions CI/CD pipeline
- Code coverage reporting with CodeCov
- Security scanning with SonarCloud
- SBOM (Software Bill of Materials) generation
- Supply chain security with SLSA Level 3

## Development Guidelines

### Version Numbering
- **Major version** (X.0.0): Breaking changes, major feature additions
- **Minor version** (X.Y.0): New features, backwards compatible
- **Patch version** (X.Y.Z): Bug fixes, security patches

### Release Process
1. Update version numbers in relevant files
2. Update CHANGELOG.md with release notes
3. Create release tag
4. Build and publish Docker images
5. Update documentation if needed

### Change Categories
- **Added**: New features
- **Changed**: Changes in existing functionality
- **Deprecated**: Soon-to-be removed features
- **Removed**: Removed features
- **Fixed**: Bug fixes
- **Security**: Security-related changes

## Migration Guides

### Upgrading from Pre-1.0 Versions
If upgrading from development versions:
1. Backup your certificate data directory
2. Review configuration changes
3. Update Docker Compose configuration if needed
4. Follow deployment guide for any breaking changes

## Support and Compatibility

### Supported Platforms
- Linux (x86_64, ARM64)
- macOS (Intel, Apple Silicon)
- Windows (x86_64)

### Dependencies
- Go 1.23+ (backend)
- Node.js 18+ (frontend development)
- Docker 20.10+ (containerized deployment)
- OpenSSL (certificate operations)

### Browser Compatibility
- Chrome 90+
- Firefox 88+
- Safari 14+
- Edge 90+

## Security Advisories

Security vulnerabilities will be documented here with:
- CVE numbers (if applicable)
- Severity level
- Affected versions
- Mitigation steps
- Fixed in version

## Contributing

When contributing changes:
1. Add entries to the [Unreleased] section
2. Use the appropriate change category
3. Include issue/PR references where applicable
4. Follow the existing format and style

## Links

- [Project Repository](https://github.com/Lazarev-Cloud/localca-go)
- [Documentation](docs/)
- [Security Policy](SECURITY.md)
- [License](LICENSE) 