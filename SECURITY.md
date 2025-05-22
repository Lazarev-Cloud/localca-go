# Security Policy

## Supported Versions

We actively support the following versions of LocalCA with security updates:

| Version | Supported          |
| ------- | ------------------ |
| 1.0.x   | :white_check_mark: |
| < 1.0   | :x:                |

## Reporting a Vulnerability

We take security vulnerabilities seriously. If you discover a security vulnerability in LocalCA, please report it responsibly.

### How to Report

**Please do NOT report security vulnerabilities through public GitHub issues.**

Instead, please report security vulnerabilities by:

1. **Email**: Send details to [security@lazarev.cloud](mailto:security@lazarev.cloud)
2. **Subject Line**: Include "LocalCA Security Vulnerability" in the subject
3. **Include**: 
   - Description of the vulnerability
   - Steps to reproduce
   - Potential impact
   - Suggested fix (if any)

### What to Expect

- **Acknowledgment**: We will acknowledge receipt within 48 hours
- **Initial Assessment**: We will provide an initial assessment within 5 business days
- **Updates**: We will keep you informed of our progress
- **Resolution**: We aim to resolve critical vulnerabilities within 30 days
- **Credit**: We will credit you in the security advisory (unless you prefer to remain anonymous)

### Responsible Disclosure

We follow responsible disclosure practices:

1. **Investigation**: We investigate and validate the report
2. **Fix Development**: We develop and test a fix
3. **Coordinated Release**: We coordinate the release with the reporter
4. **Public Disclosure**: We publish a security advisory after the fix is released

## Security Best Practices

### For Users

#### Installation Security
- Always download LocalCA from official sources
- Verify checksums and signatures when available
- Use the latest supported version
- Keep your system and dependencies updated

#### Deployment Security
- **Network Isolation**: Deploy LocalCA in a secure, isolated network
- **Access Control**: Limit access to authorized personnel only
- **HTTPS**: Always use HTTPS in production environments
- **Firewall**: Configure firewalls to restrict access to necessary ports only
- **Monitoring**: Implement logging and monitoring for security events

#### Certificate Authority Security
- **Private Key Protection**: Secure the CA private key with strong passwords
- **Key Storage**: Store CA keys in secure, encrypted storage
- **Access Logging**: Monitor and log all CA operations
- **Backup Security**: Encrypt and secure CA backups
- **Key Rotation**: Plan for CA key rotation procedures

#### Operational Security
- **Regular Updates**: Keep LocalCA updated with security patches
- **Security Scanning**: Regularly scan for vulnerabilities
- **Access Reviews**: Regularly review user access and permissions
- **Incident Response**: Have an incident response plan for security events

### For Developers

#### Code Security
- **Input Validation**: Validate all user inputs
- **Output Encoding**: Properly encode outputs to prevent injection
- **Authentication**: Implement strong authentication mechanisms
- **Authorization**: Use principle of least privilege
- **Cryptography**: Use established cryptographic libraries and practices

#### Development Practices
- **Security Reviews**: Conduct security reviews for all changes
- **Dependency Management**: Keep dependencies updated and scan for vulnerabilities
- **Secret Management**: Never commit secrets to version control
- **Testing**: Include security testing in the development process

## Security Features

### Built-in Security Measures

#### Cryptographic Security
- **TLS 1.2+**: Minimum TLS version with secure cipher suites
- **Strong Algorithms**: RSA 2048+ bit keys, SHA-256+ hashing
- **Secure Random**: Cryptographically secure random number generation
- **Key Protection**: Private keys are encrypted and securely stored

#### Application Security
- **CSRF Protection**: Cross-Site Request Forgery protection
- **Input Validation**: Comprehensive input validation and sanitization
- **Security Headers**: Security headers including CSP, HSTS, X-Frame-Options
- **Session Security**: Secure session management with HTTP-only cookies
- **Rate Limiting**: Protection against brute force attacks

#### Infrastructure Security
- **Container Security**: Minimal container images with security scanning
- **Supply Chain**: SLSA Level 3 build security with signed attestations
- **SBOM**: Software Bill of Materials for transparency
- **Vulnerability Scanning**: Automated vulnerability scanning in CI/CD

### Security Configuration

#### Environment Variables
Sensitive configuration should use environment variables:
- `CA_KEY`: CA private key password
- `SMTP_PASSWORD`: Email server password
- Database credentials (if applicable)

#### File Permissions
Ensure proper file permissions:
- CA private key: 600 (owner read/write only)
- Configuration files: 644 (owner read/write, group/other read)
- Data directory: 700 (owner access only)

#### Network Security
- Use reverse proxies with proper SSL termination
- Implement network segmentation
- Configure firewalls to allow only necessary traffic
- Use VPNs for remote access

## Threat Model

### Assets
- **CA Private Key**: Most critical asset requiring maximum protection
- **Certificate Database**: Contains all issued certificates and metadata
- **User Credentials**: Administrative access credentials
- **Configuration Data**: System configuration and secrets

### Threats
- **Unauthorized CA Access**: Compromise of CA private key
- **Certificate Misuse**: Unauthorized certificate issuance
- **Data Breach**: Exposure of certificate data or user information
- **Service Disruption**: Denial of service attacks
- **Supply Chain Attacks**: Compromise of dependencies or build process

### Mitigations
- Strong access controls and authentication
- Encryption of sensitive data at rest and in transit
- Regular security audits and vulnerability assessments
- Monitoring and alerting for suspicious activities
- Incident response procedures

## Compliance and Standards

### Standards Compliance
- **RFC 5280**: Internet X.509 Public Key Infrastructure Certificate and CRL Profile
- **RFC 8555**: Automatic Certificate Management Environment (ACME)
- **FIPS 140-2**: Cryptographic module standards (where applicable)

### Security Frameworks
- **NIST Cybersecurity Framework**: Risk management approach
- **ISO 27001**: Information security management
- **OWASP**: Web application security best practices

## Security Audits

### Internal Audits
- Regular code reviews with security focus
- Automated security scanning in CI/CD pipeline
- Dependency vulnerability scanning
- Container image security scanning

### External Audits
- We welcome security researchers and bug bounty hunters
- Consider third-party security audits for major releases
- Participate in responsible disclosure programs

## Incident Response

### Security Incident Classification
- **Critical**: Immediate threat to CA integrity or widespread impact
- **High**: Significant security vulnerability with potential for exploitation
- **Medium**: Security issue with limited impact or difficult exploitation
- **Low**: Minor security concern or theoretical vulnerability

### Response Timeline
- **Critical**: Immediate response, fix within 24-48 hours
- **High**: Response within 24 hours, fix within 1 week
- **Medium**: Response within 1 week, fix within 1 month
- **Low**: Response within 1 month, fix in next release cycle

## Contact Information

- **Security Email**: [security@lazarev.cloud](mailto:security@lazarev.cloud)
- **General Contact**: [contact@lazarev.cloud](mailto:contact@lazarev.cloud)
- **Project Repository**: [https://github.com/Lazarev-Cloud/localca-go](https://github.com/Lazarev-Cloud/localca-go)

## Acknowledgments

We thank the security research community for their contributions to making LocalCA more secure. Security researchers who responsibly disclose vulnerabilities will be acknowledged in our security advisories (unless they prefer to remain anonymous).

---

**Note**: This security policy is subject to change. Please check back regularly for updates. 