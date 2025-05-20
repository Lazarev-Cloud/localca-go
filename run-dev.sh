#!/bin/bash
# Set environment variables for development
export CA_NAME=LocalCA
export ORGANIZATION="LocalCA Organization"
export COUNTRY=US
export DATA_DIR=./data
export LISTEN_ADDR=:8080

# Run the application
go run main.go 