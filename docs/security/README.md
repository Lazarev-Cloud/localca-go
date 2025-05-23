# Security & Analysis

This directory contains security analysis files and configurations.

## Files

### SBOM (Software Bill of Materials)
- `SPDX.json` - SPDX format SBOM
- `sbom.cyclonedx.json` - CycloneDX format SBOM
- `sbom.spdx.json` - SPDX format SBOM (alternative)

### Code Quality & Analysis
- `codecov.yml` - Code coverage configuration
- `eslint-report.json` - ESLint analysis results
- `sonar-project.properties` - SonarQube project configuration

## Security Practices

This project implements several security measures:

1. **Input Validation** - All user inputs are properly validated and sanitized
2. **Path Traversal Protection** - Secure path handling to prevent directory traversal
3. **Session Management** - Secure session handling with proper expiration
4. **CSRF Protection** - Cross-Site Request Forgery protection on all forms
5. **Security Headers** - Comprehensive security headers implementation
6. **Command Injection Prevention** - Secure handling of external commands

## Security Scanning

Regular security scans are performed using:
- Static code analysis
- Dependency vulnerability scanning
- Container image scanning
- SBOM generation for supply chain security

See `../tools/run-security-scan.sh` for automated security scanning.