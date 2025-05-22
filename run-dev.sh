#!/bin/bash
# Set environment variables for development
export CA_NAME=LocalCA
export ORGANIZATION="LocalCA Organization"
export COUNTRY=US
export DATA_DIR=./data
export LISTEN_ADDR=:8080
export CA_KEY_FILE=cakey.txt
export ALLOW_LOCALHOST=true

# Run the application
go run main.go 