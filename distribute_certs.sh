#!/bin/bash
# distribute_certs.sh - Distribute certificates to servers

# Configuration
API_URL="https://localhost:8443"
CERT_NAME="server.domain.local"
SERVERS=("web1.domain.local" "web2.domain.local" "web3.domain.local")
DEST_DIR="/etc/ssl/certs"
SSH_USER="admin"
SSH_KEY="~/.ssh/id_rsa"
SERVICE_TO_RELOAD="nginx"

# Function to log messages
log() {
    echo "[$(date +"%Y-%m-%d %H:%M:%S")] $1"
}

log "Starting certificate distribution for $CERT_NAME"

# Create temporary directory
TEMP_DIR=$(mktemp -d)
log "Created temporary directory: $TEMP_DIR"

# Download certificate files
log "Downloading certificate files from $API_URL"
curl -sk -o "$TEMP_DIR/$CERT_NAME.crt" "$API_URL/download/$CERT_NAME/crt"
curl -sk -o "$TEMP_DIR/$CERT_NAME.key" "$API_URL/download/$CERT_NAME/key"
curl -sk -o "$TEMP_DIR/$CERT_NAME.bundle.crt" "$API_URL/download/$CERT_NAME/bundle"
curl -sk -o "$TEMP_DIR/ca.crl" "$API_URL/download/crl"
curl -sk -o "$TEMP_DIR/ca.pem" "$API_URL/download/ca"

# Check if files were downloaded successfully
if [ ! -s "$TEMP_DIR/$CERT_NAME.crt" ] || [ ! -s "$TEMP_DIR/$CERT_NAME.key" ]; then
    log "Failed to download certificate files"
    rm -rf "$TEMP_DIR"
    exit 1
fi

# Distribute to servers
for SERVER in "${SERVERS[@]}"; do
    log "Deploying certificates to $SERVER..."
    
    # Check if server is reachable
    if ! ping -c 1 -W 2 "$SERVER" > /dev/null 2>&1; then
        log "Server $SERVER is not reachable, skipping"
        continue
    fi
    
    # Create directory on remote server if it doesn't exist
    ssh -i "$SSH_KEY" "${SSH_USER}@${SERVER}" "mkdir -p $DEST_DIR" || {
        log "Failed to create directory on $SERVER, skipping"
        continue
    }
    
    # Copy certificate files
    log "Copying certificate files to $SERVER:$DEST_DIR/"
    scp -i "$SSH_KEY" "$TEMP_DIR/$CERT_NAME.crt" "${SSH_USER}@${SERVER}:${DEST_DIR}/"
    scp -i "$SSH_KEY" "$TEMP_DIR/$CERT_NAME.key" "${SSH_USER}@${SERVER}:${DEST_DIR}/"
    scp -i "$SSH_KEY" "$TEMP_DIR/$CERT_NAME.bundle.crt" "${SSH_USER}@${SERVER}:${DEST_DIR}/"
    scp -i "$SSH_KEY" "$TEMP_DIR/ca.crl" "${SSH_USER}@${SERVER}:${DEST_DIR}/"
    scp -i "$SSH_KEY" "$TEMP_DIR/ca.pem" "${SSH_USER}@${SERVER}:${DEST_DIR}/"
    
    # Set proper permissions
    log "Setting permissions on certificate files"
    ssh -i "$SSH_KEY" "${SSH_USER}@${SERVER}" "chmod 644 ${DEST_DIR}/${CERT_NAME}.crt"
    ssh -i "$SSH_KEY" "${SSH_USER}@${SERVER}" "chmod 644 ${DEST_DIR}/${CERT_NAME}.bundle.crt"
    ssh -i "$SSH_KEY" "${SSH_USER}@${SERVER}" "chmod 600 ${DEST_DIR}/${CERT_NAME}.key"
    ssh -i "$SSH_KEY" "${SSH_USER}@${SERVER}" "chmod 644 ${DEST_DIR}/ca.crl"
    ssh -i "$SSH_KEY" "${SSH_USER}@${SERVER}" "chmod 644 ${DEST_DIR}/ca.pem"
    
    # Reload service
    log "Reloading $SERVICE_TO_RELOAD on $SERVER"
    ssh -i "$SSH_KEY" "${SSH_USER}@${SERVER}" "systemctl reload $SERVICE_TO_RELOAD || systemctl restart $SERVICE_TO_RELOAD" || {
        log "Failed to reload $SERVICE_TO_RELOAD on $SERVER"
    }
    
    log "Deployment to $SERVER completed successfully"
done

# Clean up
log "Cleaning up temporary files"
rm -rf "$TEMP_DIR"

log "Certificate distribution completed successfully"