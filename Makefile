# Makefile for LocalCA

.PHONY: all build run stop restart status logs backup renew-certs distribute-certs monitor

# Configuration
CONTAINER_NAME := localca
DATA_DIR := ./data
CERT_NAME := server.example.local
SERVERS := web1.example.local,web2.example.local

all: build run

# Build the Docker image
build:
	@echo "Building LocalCA..."
	docker-compose build

# Start the service
run:
	@echo "Starting LocalCA..."
	docker-compose up -d
	@echo "LocalCA is now running at: https://localhost:8443"

# Stop the service
stop:
	@echo "Stopping LocalCA..."
	docker-compose down

# Restart the service
restart: stop run

# Check service status
status:
	@echo "LocalCA status:"
	docker-compose ps

# View logs
logs:
	docker-compose logs -f

# Create a backup
backup:
	@echo "Backing up LocalCA data..."
	tar -czf localca_backup_$(shell date +"%Y%m%d_%H%M%S").tar.gz $(DATA_DIR)
	@echo "Backup created."

# Renew certificates
renew-certs:
	@echo "Renewing certificates..."
	./scripts/renew_certs.sh
	@echo "Certificate renewal completed."

# Distribute certificates to servers
distribute-certs:
	@echo "Distributing certificates to servers..."
	./scripts/distribute_certs.sh --cert $(CERT_NAME) --servers $(SERVERS)
	@echo "Certificate distribution completed."

# Monitor certificate expiration
monitor:
	@echo "Monitoring certificate expiration..."
	./scripts/monitor_certs.sh
	@echo "Monitoring completed."

# Generate a new client certificate
client-cert:
	@read -p "Enter client name: " CLIENT_NAME; \
	read -p "Enter client certificate password: " -s PASSWORD; \
	echo ""; \
	docker exec -it $(CONTAINER_NAME) /app/scripts/create_client_cert.sh "$$CLIENT_NAME" "$$PASSWORD"

# Set up a new server with the CA
setup-server:
	@read -p "Enter server hostname: " SERVER; \
	read -p "Enter SSH user: " SSH_USER; \
	./scripts/setup_server.sh "$$SERVER" "$$SSH_USER"

# Help
help:
	@echo "LocalCA Management Commands:"
	@echo "  make build            - Build the Docker image"
	@echo "  make run              - Start the service"
	@echo "  make stop             - Stop the service"
	@echo "  make restart          - Restart the service"
	@echo "  make status           - Check service status"
	@echo "  make logs             - View logs"
	@echo "  make backup           - Create a backup"
	@echo "  make renew-certs      - Renew certificates"
	@echo "  make distribute-certs - Distribute certificates to servers"
	@echo "  make monitor          - Monitor certificate expiration"
	@echo "  make client-cert      - Generate a new client certificate"
	@echo "  make setup-server     - Set up a new server with the CA"