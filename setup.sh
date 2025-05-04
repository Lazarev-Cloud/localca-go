#!/bin/bash

# This script prepares the data directory for LocalCA

# Create data directory if it doesn't exist
mkdir -p data/ca

# Give appropriate permissions
chmod -R 755 data