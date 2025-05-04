#!/bin/bash

# This script prepares the environment for LocalCA

# Create data directory if it doesn't exist
mkdir -p data/ca

# Generate a strong random password for the CA
if [ ! -f cakey.txt ]; then
    # Generate a secure random password
    PASSWORD=$(openssl rand -base64 24)
    echo "Generating CA key password..."
    echo "$PASSWORD" > cakey.txt
    chmod 600 cakey.txt
    echo "CA key password saved to cakey.txt"
else
    echo "Using existing CA key password"
fi

# Give appropriate permissions
chmod -R 755 data
chmod 600 cakey.txt

echo "Setup complete! You can now run: docker-compose up -d"