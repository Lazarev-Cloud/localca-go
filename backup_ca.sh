#!/bin/bash
# backup_ca.sh - Backup CA certificates and keys

# Configuration
BACKUP_DIR="/backup/localca"
DATA_DIR="/app/data"
RETENTION_DAYS=30
ENCRYPTION_KEY="path/to/encryption.key"  # Optional: Encryption key
BACKUP_REMOTE="user@backupserver:/backups/localca"  # Optional: Remote backup location

# Function to log messages
log() {
    echo "[$(date +"%Y-%m-%d %H:%M:%S")] $1"
}

# Create backup directory if it doesn't exist
mkdir -p "$BACKUP_DIR"

# Generate backup filename with timestamp
TIMESTAMP=$(date +"%Y%m%d_%H%M%S")
BACKUP_FILE="$BACKUP_DIR/localca_backup_$TIMESTAMP.tar.gz"

log "Starting LocalCA backup to $BACKUP_FILE"

# Create tar archive of the data directory
tar -czf "$BACKUP_FILE" -C "$(dirname "$DATA_DIR")" "$(basename "$DATA_DIR")"

# Encrypt the backup if encryption key is provided
if [ -f "$ENCRYPTION_KEY" ]; then
    log "Encrypting backup..."
    openssl enc -aes-256-cbc -salt -in "$BACKUP_FILE" -out "$BACKUP_FILE.enc" -pass file:"$ENCRYPTION_KEY"
    
    # Replace original with encrypted file
    if [ -f "$BACKUP_FILE.enc" ]; then
        rm "$BACKUP_FILE"
        BACKUP_FILE="$BACKUP_FILE.enc"
        log "Backup encrypted successfully"
    else
        log "Encryption failed, keeping unencrypted backup"
    fi
fi

# Calculate checksum
sha256sum "$BACKUP_FILE" > "$BACKUP_FILE.sha256"
log "Generated checksum: $(cat "$BACKUP_FILE.sha256")"

# Copy to remote location if specified
if [ -n "$BACKUP_REMOTE" ]; then
    log "Copying backup to remote location: $BACKUP_REMOTE"
    rsync -avz "$BACKUP_FILE" "$BACKUP_FILE.sha256" "$BACKUP_REMOTE/"
    
    if [ $? -eq 0 ]; then
        log "Remote backup completed successfully"
    else
        log "Remote backup failed"
    fi
fi

# Clean up old backups
log "Cleaning up backups older than $RETENTION_DAYS days"
find "$BACKUP_DIR" -name "localca_backup_*.tar.gz*" -type f -mtime +$RETENTION_DAYS -delete
find "$BACKUP_DIR" -name "localca_backup_*.sha256" -type f -mtime +$RETENTION_DAYS -delete

log "Backup completed successfully"