@echo off
REM Set environment variables for development
set CA_NAME=LocalCA
set ORGANIZATION=LocalCA Organization
set COUNTRY=US
set DATA_DIR=./data
set LISTEN_ADDR=:8080
set CA_KEY=development-password

REM Run the application
go run main.go 