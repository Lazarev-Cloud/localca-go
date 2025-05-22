# LocalCA Documentation

This directory contains comprehensive documentation for the LocalCA project, covering deployment, security, testing, and operational procedures.

## Quick Start

- [README.md](../README.md) - Main project overview and quick start guide
- [DEPLOYMENT.md](DEPLOYMENT.md) - Detailed deployment and configuration guide
- [BestPractice.md](BestPractice.md) - Security best practices and guidelines

## Development Documentation

- [README-TESTING.md](README-TESTING.md) - Testing documentation and procedures
- [WORKFLOW_OPTIMIZATION_COMPLETE.md](WORKFLOW_OPTIMIZATION_COMPLETE.md) - Development workflow optimizations
- [CICD_OPTIMIZATION_SUMMARY.md](CICD_OPTIMIZATION_SUMMARY.md) - CI/CD pipeline documentation

## Troubleshooting and Fixes

- [FIXES_SUMMARY.md](FIXES_SUMMARY.md) - Summary of common fixes and solutions
- [DOCKER-FIXES.md](DOCKER-FIXES.md) - Docker-specific troubleshooting

## Architecture and Development Guidelines

For detailed development guidelines and architectural information, see the [Cursor Rules](../.cursor/rules/) directory:

- **Project Overview**: Understanding the LocalCA architecture
- **Backend Architecture**: Go backend structure and patterns
- **Frontend Architecture**: Next.js frontend organization
- **Certificate Lifecycle**: Certificate management workflows
- **ACME Protocol**: ACME implementation details
- **Security Model**: Security architecture and practices
- **Development Guide**: Coding standards and best practices
- **API Endpoints**: API documentation and data flow
- **Troubleshooting**: Debugging and problem resolution

## Getting Help

1. **Common Issues**: Check the troubleshooting documentation first
2. **Development Questions**: Refer to the development guides in `.cursor/rules/`
3. **Security Concerns**: Review the security best practices
4. **Deployment Issues**: Consult the deployment guide

## Contributing to Documentation

When contributing to the project, please:

1. Update relevant documentation for any changes
2. Add new documentation for new features
3. Follow the existing documentation structure
4. Include code examples where appropriate
5. Update this index when adding new documentation files

## Documentation Structure

```
docs/
├── README.md                           # This file - documentation index
├── DEPLOYMENT.md                       # Production deployment guide
├── BestPractice.md                     # Security and operational best practices
├── README-TESTING.md                   # Testing procedures and guidelines
├── WORKFLOW_OPTIMIZATION_COMPLETE.md   # Development workflow documentation
├── CICD_OPTIMIZATION_SUMMARY.md        # CI/CD pipeline documentation
├── FIXES_SUMMARY.md                    # Common fixes and solutions
└── DOCKER-FIXES.md                     # Docker troubleshooting

.cursor/rules/                          # Development guidelines (Cursor IDE)
├── 01-project-overview.mdc             # Project architecture overview
├── 02-backend-architecture.mdc         # Go backend structure
├── 03-frontend-architecture.mdc        # Next.js frontend structure
├── 04-certificate-lifecycle.mdc        # Certificate management
├── 05-deployment-configuration.mdc     # Deployment configuration
├── 06-acme-protocol.mdc                # ACME protocol implementation
├── 07-security-model.mdc               # Security architecture
├── 08-project-workflow.mdc             # Development workflow
├── 09-development-guide.mdc            # Coding standards and practices
├── 10-api-endpoints.mdc                # API documentation
└── 11-troubleshooting.mdc              # Debugging and troubleshooting
```