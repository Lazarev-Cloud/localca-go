#!/bin/bash
# deploy_localca.sh - Deploy LocalCA with Docker

# Configuration
LOCALCA_DIR="/opt/localca"
GIT_REPO="https://github.com/Lazarev-Cloud/localca-go.git"
DOMAIN="ca.example.local"
ORGANIZATION="YourOrganization"
COUNTRY="US"

# Function to log messages
log() {
    echo "[$(date +"%Y-%m-%d %H:%M:%S")] $1"
}

# Check if Docker and Docker Compose are installed
if ! command -v docker &> /dev/null; then
    log "Docker is not installed. Please install Docker first."
    exit 1
fi

if ! command -v docker-compose &> /dev/null; then
    log "Docker Compose is not installed. Please install Docker Compose first."
    exit 1
fi

# Create directory
mkdir -p "$LOCALCA_DIR"
cd "$LOCALCA_DIR" || exit 1

# Clone or update repository
if [ -d ".git" ]; then
    log "Updating existing repository..."
    git pull
else
    log "Cloning repository..."
    git clone "$GIT_REPO" .
fi

# Create or update .env file
log "Configuring environment..."
cat > .env << EOF
CA_NAME=$DOMAIN
ORGANIZATION=$ORGANIZATION
COUNTRY=$COUNTRY
TLS_ENABLED=true
EMAIL_NOTIFY=true
EOF

# Generate CA key password if it doesn't exist
if [ ! -f "cakey.txt" ]; then
    log "Generating CA key password..."
    openssl rand -base64 32 > cakey.txt
    chmod 600 cakey.txt
fi

# Create data directory
mkdir -p data/ca
chmod -R 755 data

# Build and start containers
log "Building and starting LocalCA..."
docker-compose down
docker-compose build --no-cache
docker-compose up -d

# Check if containers are running
if docker-compose ps | grep -q "localca.*Up"; then
    log "LocalCA is now running!"
    log "Access the web interface at: https://$DOMAIN:8443"
    log "If you're using this on a development machine, add the following to your /etc/hosts file:"
    log "127.0.0.1 $DOMAIN"
else
    log "Something went wrong. Check logs with: docker-compose logs"
    exit 1
fi

# Display CA certificate path
log "Once certificates are generated, you can find them in: $LOCALCA_DIR/data"
log "To install the CA certificate on your system:"
log "  - Download it from the web interface"
log "  - Or use the file at: $LOCALCA_DIR/data/ca.pem"

log "Deployment completed successfully!"