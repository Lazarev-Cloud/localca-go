#!/bin/bash
# monitor_certs.sh - Monitor certificate expiration

# Configuration
API_URL="https://localhost:8443"
WARN_DAYS=30
CRITICAL_DAYS=7
ADMIN_EMAIL="admin@domain.local"
SLACK_WEBHOOK="" # Optional: Add Slack webhook URL for notifications

# Function to log messages
log() {
    echo "[$(date +"%Y-%m-%d %H:%M:%S")] $1"
}

# Function to send email alert
send_email() {
    local subject="$1"
    local body="$2"
    
    echo "$body" | mail -s "$subject" "$ADMIN_EMAIL"
}

# Function to send Slack notification
send_slack() {
    local message="$1"
    
    if [ -n "$SLACK_WEBHOOK" ]; then
        curl -s -X POST -H "Content-type: application/json" \
            --data "{\"text\":\"$message\"}" \
            "$SLACK_WEBHOOK"
    fi
}

log "Starting certificate monitoring"

# Get list of certificates
CERTS=$(curl -sk "${API_URL}" | grep -oP '(?<=<td><a href="/files\?name=)[^"]+')

if [ -z "$CERTS" ]; then
    log "No certificates found or failed to connect to $API_URL"
    exit 1
fi

# Check CA certificate first
CA_INFO=$(curl -sk "${API_URL}" | grep -A5 "CA Certificate")
CA_EXPIRY=$(echo "$CA_INFO" | grep -oP '(?<=<span class=")[^"]*(?=">)[^<]*(?=</span>)' | tail -1)

if echo "$CA_INFO" | grep -q "text-danger"; then
    log "CRITICAL: CA certificate is expired or will expire soon"
    send_email "CRITICAL: CA Certificate Expiration" "The CA certificate has expired or will expire soon. Please renew immediately."
    send_slack ":red_circle: *CRITICAL: CA Certificate Expiration* - The CA certificate has expired or will expire soon. Please renew immediately."
fi

# Monitor all certificates
WARN_COUNT=0
CRITICAL_COUNT=0

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
    fi
    
    # Get current time in seconds since epoch
    CURRENT_SECONDS=$(date +%s)
    
    # Calculate days until expiry
    DAYS_LEFT=$(( ($EXPIRY_SECONDS - $CURRENT_SECONDS) / 86400 ))
    
    # Check if already expired
    if [ $DAYS_LEFT -lt 0 ]; then
        log "CRITICAL: Certificate $CERT has expired $((DAYS_LEFT * -1)) days ago"
        send_email "CRITICAL: Certificate Expired - $CERT" "Certificate $CERT has expired $((DAYS_LEFT * -1)) days ago."
        send_slack ":red_circle: *CRITICAL: Certificate Expired* - Certificate $CERT has expired $((DAYS_LEFT * -1)) days ago."
        CRITICAL_COUNT=$((CRITICAL_COUNT + 1))
    # Check if critical expiry
    elif [ $DAYS_LEFT -lt $CRITICAL_DAYS ]; then
        log "CRITICAL: Certificate $CERT will expire in $DAYS_LEFT days"
        send_email "CRITICAL: Certificate Expiring Soon - $CERT" "Certificate $CERT will expire in $DAYS_LEFT days. Please renew immediately."
        send_slack ":red_circle: *CRITICAL: Certificate Expiring Soon* - Certificate $CERT will expire in $DAYS_LEFT days. Please renew immediately."
        CRITICAL_COUNT=$((CRITICAL_COUNT + 1))
    # Check if warning expiry
    elif [ $DAYS_LEFT -lt $WARN_DAYS ]; then
        log "WARNING: Certificate $CERT will expire in $DAYS_LEFT days"
        send_email "WARNING: Certificate Expiring - $CERT" "Certificate $CERT will expire in $DAYS_LEFT days. Please plan to renew soon."
        send_slack ":warning: *WARNING: Certificate Expiring* - Certificate $CERT will expire in $DAYS_LEFT days. Please plan to renew soon."
        WARN_COUNT=$((WARN_COUNT + 1))
    else
        log "Certificate $CERT is valid for $DAYS_LEFT days"
    fi
done

log "Certificate monitoring completed: $CRITICAL_COUNT critical, $WARN_COUNT warnings"