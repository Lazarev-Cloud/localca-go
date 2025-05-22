# Development Tools

This directory contains various development and utility scripts for the LocalCA project.

## Scripts

### Development Scripts
- `run-dev.sh` / `run-dev.bat` - Start development environment
- `run-docker.sh` / `run-docker.bat` - Run with Docker

### Testing Scripts
- `run-tests.sh` / `run-tests.bat` - Run Go tests with coverage
- `run-tests-docker.sh` / `run-tests-docker.bat` - Run comprehensive Docker test suite
- `comprehensive_test.sh` - Full application testing
- `simple-validation.sh` - Basic validation tests
- `test_application.sh` - Application-specific tests

### Security Scripts
- `run-security-scan.sh` / `run-security-scan.bat` - Security vulnerability scanning

### Build & Maintenance Scripts
- `fix-workflows.sh` - Fix GitHub Actions workflows
- `validate-workflows.sh` - Validate workflow configurations
- `syft-install.sh` - Install Syft for SBOM generation

## Usage

Make sure scripts are executable:
```bash
chmod +x tools/*.sh
```

Run from project root:
```bash
# Example: Run tests
./tools/run-tests.sh

# Example: Start development environment
./tools/run-dev.sh
```