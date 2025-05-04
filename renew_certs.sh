#!/bin/bash
# renew_certs.sh - Automatically renew certificates that will expire soon

# Configuration
API_URL="https://localhost:8443"
EXPIRY_DAYS=30
LOG_FILE="renewal.log"

# Function to log messages
log() {
    echo "[$(date +"%Y-%m-%d %H:%M:%S")] $1" | tee -a "$LOG_FILE"
}

log "Starting automatic certificate renewal"

# Get list of certificates
CERTS=$(curl -sk "${API_URL}" | grep -oP '(?<=<td><a href="/files\?name=)[^"]+')

for CERT in $CERTS; do
    log "Checking certificate: $CERT"
    
    # Get certificate details
    DETAILS=$(curl -sk "${API_URL}/files?name=${CERT}")
    
    # Extract expiry date
    EXPIRY=$(echo "$DETAILS" | grep -oP '(?<=<span class="certificate-detail-label">Valid To:</span> )[^<]+')
    if [ -z "$EXPIRY" ]; then
        log "Failed to get expiry date for $CERT, skipping"
        continue
    fi
    
    # Convert expiry date to seconds since epoch
    EXPIRY_SECONDS=$(date -d "$EXPIRY" +%s 2>/dev/null)
    if [ $? -ne 0 ]; then
        log "Failed to parse expiry date for $CERT, skipping"
        continue
    }
    
    # Get current time in seconds since epoch
    CURRENT_SECONDS=$(date +%s)
    
    # Calculate days until expiry
    DAYS_LEFT=$(( ($EXPIRY_SECONDS - $CURRENT_SECONDS) / 86400 ))
    
    log "Certificate $CERT expires in $DAYS_LEFT days"
    
    # Check if certificate needs renewal
    if [ $DAYS_LEFT -lt $EXPIRY_DAYS ]; then
        log "Renewing certificate $CERT..."
        
        # Get CSRF token
        CSRF_TOKEN=$(curl -sk "${API_URL}" | grep -oP '(?<=name="csrf_token" value=")[^"]+')
        
        # Renew certificate
        RESPONSE=$(curl -sk -X POST "${API_URL}/renew" \
            -d "name=${CERT}" \
            -d "csrf_token=${CSRF_TOKEN}")
        
        if echo "$RESPONSE" | grep -q "success\":true"; then
            log "Certificate $CERT renewed successfully"
        else
            ERROR=$(echo "$RESPONSE" | grep -oP '(?<="message":")(.*?)(?=")' || echo "Unknown error")
            log "Failed to renew certificate $CERT: $ERROR"
        fi
    else
        log "Certificate $CERT still valid, skipping renewal"
    fi
done

log "Automatic certificate renewal completed"