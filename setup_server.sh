#!/bin/bash
# setup_server.sh - Set up a server with the CA certificate

# Check arguments
if [ $# -lt 2 ]; then
    echo "Usage: $0 <server_hostname> <ssh_user>"
    exit 1
fi

SERVER="$1"
SSH_USER="$2"
API_URL="https://localhost:8443"

# Function to log messages
log() {
    echo "[$(date +"%Y-%m-%d %H:%M:%S")] $1"
}

log "Setting up server: $SERVER"

# Check if server is reachable
if ! ping -c 1 -W 2 "$SERVER" > /dev/null 2>&1; then
    log "Server $SERVER is not reachable"
    exit 1
fi

# Create temporary directory
TEMP_DIR=$(mktemp -d)
log "Created temporary directory: $TEMP_DIR"

# Download CA certificate
log "Downloading CA certificate"
curl -sk -o "$TEMP_DIR/ca.pem" "$API_URL/download/ca"

if [ ! -s "$TEMP_DIR/ca.pem" ]; then
    log "Failed to download CA certificate"
    rm -rf "$TEMP_DIR"
    exit 1
fi

# Copy CA certificate to server
log "Copying CA certificate to server"
scp "$TEMP_DIR/ca.pem" "${SSH_USER}@${SERVER}:/tmp/ca.pem"

# Set up CA certificate on server
log "Setting up CA certificate on server"
ssh "${SSH_USER}@${SERVER}" << 'EOF'
    # Detect distribution
    if [ -f /etc/debian_version ]; then
        # Debian/Ubuntu
        sudo mkdir -p /usr/local/share/ca-certificates
        sudo cp /tmp/ca.pem /usr/local/share/ca-certificates/localca.crt
        sudo update-ca-certificates
    elif [ -f /etc/redhat-release ]; then
        # CentOS/RHEL
        sudo mkdir -p /etc/pki/ca-trust/source/anchors
        sudo cp /tmp/ca.pem /etc/pki/ca-trust/source/anchors/localca.pem
        sudo update-ca-trust extract
    elif [ -f /etc/arch-release ]; then
        # Arch Linux
        sudo mkdir -p /etc/ca-certificates/trust-source/anchors
        sudo cp /tmp/ca.pem /etc/ca-certificates/trust-source/anchors/localca.pem
        sudo trust extract-compat
    else
        echo "Unknown distribution, installing to /etc/ssl/certs"
        sudo mkdir -p /etc/ssl/certs
        sudo cp /tmp/ca.pem /etc/ssl/certs/localca.pem
    fi
    
    # Remove temp file
    rm /tmp/ca.pem
    
    echo "CA certificate installed successfully"
EOF

# Clean up
log "Cleaning up temporary files"
rm -rf "$TEMP_DIR"

log "Server setup completed successfully"