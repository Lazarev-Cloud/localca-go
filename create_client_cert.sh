#!/bin/bash
# create_client_cert.sh - Create a client certificate from command line

# Check arguments
if [ $# -lt 2 ]; then
    echo "Usage: $0 <client_name> <password>"
    exit 1
fi

CLIENT_NAME="$1"
PASSWORD="$2"

# Function to log messages
log() {
    echo "[$(date +"%Y-%m-%d %H:%M:%S")] $1"
}

# Send request to the API
log "Creating client certificate for: $CLIENT_NAME"

# Get CSRF token
CSRF_TOKEN=$(curl -sk "https://localhost:8443" | grep -oP '(?<=name="csrf_token" value=")[^"]+')

if [ -z "$CSRF_TOKEN" ]; then
    log "Failed to get CSRF token"
    exit 1
fi

# Create client certificate
RESPONSE=$(curl -sk -X POST "https://localhost:8443/" \
    -d "cn=$CLIENT_NAME" \
    -d "client=on" \
    -d "password=$PASSWORD" \
    -d "csrf_token=$CSRF_TOKEN")

# Check if successful
if echo "$RESPONSE" | grep -q "Certificate created successfully" || echo "$RESPONSE" | grep -q "Certificates"; then
    log "Certificate created successfully"
    
    # Download the certificate
    curl -sk -o "$CLIENT_NAME.p12" "https://localhost:8443/download/$CLIENT_NAME/p12"
    
    if [ -s "$CLIENT_NAME.p12" ]; then
        log "Certificate downloaded to: $CLIENT_NAME.p12"
        log "You can import this certificate into your browser or device"
        log "Use the password you provided when prompted"
    else
        log "Failed to download certificate"
        exit 1
    fi
else
    log "Failed to create certificate"
    echo "$RESPONSE" | grep -o "<div class=\"alert alert-danger\">[^<]*</div>" | sed 's/<[^>]*>//g'
    exit 1
fi