# LocalCA Deployment Guide

This guide will walk you through setting up LocalCA, a self-hosted Certificate Authority for your local network.

## Prerequisites

- Docker and Docker Compose installed on your server
- Basic knowledge of SSL/TLS certificates
- A dedicated machine or VM to run the service (recommended)

## Initial Setup

### Step 1: Clone the Repository

```bash
git clone https://github.com/Lazarev-Cloud/localca-go.git
cd localca-go
```

### Step 2: Create a Password for Your CA

The CA private key needs to be protected with a strong password. Create a file named `cakey.txt` in the project root directory:

```bash
echo "your-secure-password" > cakey.txt
chmod 600 cakey.txt  # Restrict permissions
```

Replace "your-secure-password" with a strong, unique password. This password will be used to encrypt and protect your CA private key.

### Step 3: Configure the Service

Review and modify the `docker-compose.yml` file to set up your CA properly:

- **CA_NAME**: Set to your desired CA name (e.g., `ca.yourdomain.local`)
- **ORGANIZATION**: Set to your organization name
- **COUNTRY**: Set to your two-letter country code
- **TLS_ENABLED**: Keep as `true` for secure HTTPS access
- **EMAIL_NOTIFY**: Set to `true` if you want email notifications about expiring certificates

### Step 4: Build and Start the Service

```bash
docker-compose up -d
```

This will build the Docker image and start the LocalCA service in the background.

### Step 5: Access the Web Interface

Open your browser and navigate to:

```
https://your-server-ip:8443
```

You'll see a security warning because the browser doesn't trust the CA certificate yet. This is expected and normal. Proceed to the website.

### Step 6: Download and Trust the CA Certificate

1. From the web interface, click the **Download CA** button
2. Install the CA certificate on your devices following the instructions below

## Installing the CA Certificate

### On Windows

1. Double-click the downloaded `ca.pem` file
2. Click "Install Certificate"
3. Select "Local Machine" and click "Next"
4. Select "Place all certificates in the following store"
5. Click "Browse" and select "Trusted Root Certification Authorities"
6. Click "Next" and then "Finish"

### On macOS

1. Double-click the downloaded `ca.pem` file
2. This will open Keychain Access
3. Locate the certificate in your login keychain
4. Double-click it, expand "Trust," and set "When using this certificate" to "Always Trust"
5. Close the window and enter your password to save changes

### On Linux

1. Copy the CA certificate to the trust store:

```bash
sudo cp ca.pem /usr/local/share/ca-certificates/localca.crt
sudo update-ca-certificates
```

### In Firefox (uses its own certificate store)

1. Go to Preferences > Privacy & Security > Certificates > View Certificates
2. Go to the "Authorities" tab
3. Click "Import" and select the downloaded `ca.pem` file
4. Check "Trust this CA to identify websites" and click "OK"

### In Chrome/Edge on Windows

These browsers use the Windows certificate store, so follow the Windows instructions above.

### In Chrome/Edge on macOS

These browsers use the macOS certificate store, so follow the macOS instructions above.

### In Chrome/Edge on Linux

These browsers use the Linux certificate store, so follow the Linux instructions above.

### On Mobile Devices

#### Android

1. Go to Settings > Security > Install from Storage
2. Select the downloaded `ca.pem` file
3. Name the certificate and select "VPN and apps" for credential use

#### iOS

1. Email the CA certificate to yourself or make it available for download
2. Open the CA certificate file
3. Go to Settings > Profile Downloaded
4. Install the profile and trust it in Settings > General > About > Certificate Trust Settings

## Creating and Using Certificates

### Creating a Server Certificate

1. Log in to the LocalCA web interface
2. In the "Create Certificate" form:
   - Enter the server hostname in "Common Name" (e.g., `server.local`)
   - Add alternative names in "Additional Domain Names" if needed (comma-separated)
   - Ensure "Create client certificate" is NOT checked
   - Click "Create Certificate"
3. View the certificate details and download the files
4. Use the `.crt` and `.key` files in your web server configuration

### Creating a Client Certificate

1. Log in to the LocalCA web interface
2. In the "Create Certificate" form:
   - Enter a name in "Common Name" (e.g., `john.doe`)
   - Check "Create client certificate"
   - Enter a password for the P12 file
   - Click "Create Certificate"
3. Download the `.p12` file and import it into your browser or client application

### Certificate Renewal

1. Log in to the LocalCA web interface
2. Find the certificate in the list
3. Click the "Renew" button
4. The certificate will be renewed with a new expiration date
5. Re-deploy the renewed certificate to your services

### Certificate Revocation

1. Log in to the LocalCA web interface
2. Find the certificate in the list
3. Click the "Revoke" button
4. Download the updated CRL and distribute it to your services

## Web Server Configuration

### NGINX Configuration

```nginx
server {
    listen 443 ssl;
    server_name your-server.local;
    
    ssl_certificate /path/to/your-server.local.bundle.crt;
    ssl_certificate_key /path/to/your-server.local.key;
    
    # Enable certificate revocation checking
    ssl_crl /path/to/ca.crl;
    
    # Optional: Enable client certificate authentication
    ssl_client_certificate /path/to/ca.pem;
    ssl_verify_client on;
    
    # Strong SSL settings
    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_prefer_server_ciphers on;
    ssl_ciphers EECDH+AESGCM:EDH+AESGCM;
    ssl_session_timeout 1d;
    ssl_session_cache shared:SSL:50m;
    ssl_stapling off;  # OCSP stapling not needed for local CA
    
    # Rest of your server configuration...
}
```

### Apache Configuration

```apache
<VirtualHost *:443>
    ServerName your-server.local
    
    SSLEngine on
    SSLCertificateFile /path/to/your-server.local.crt
    SSLCertificateKeyFile /path/to/your-server.local.key
    SSLCACertificateFile /path/to/ca.pem
    
    # Enable certificate revocation checking
    SSLCARevocationPath /path/to/
    SSLCARevocationFile /path/to/ca.crl
    SSLVerifyDepth 1
    
    # Optional: Enable client certificate authentication
    SSLVerifyClient require
    
    # Strong SSL settings
    SSLProtocol all -SSLv3 -TLSv1 -TLSv1.1
    SSLHonorCipherOrder on
    SSLCipherSuite EECDH+AESGCM:EDH+AESGCM
    
    # Rest of your server configuration...
</VirtualHost>
```

## Automated Certificate Deployment

You can automate certificate deployment using simple scripts. Here's an example script that fetches a certificate from LocalCA and deploys it to an NGINX server:

```bash
#!/bin/bash

# Configuration
SERVER_NAME="your-server.local"
LOCALCA_URL="https://localca:8443"
NGINX_CONF_DIR="/etc/nginx"
CERT_DIR="/etc/ssl/certs"

# Download certificates
curl -k -o "$CERT_DIR/$SERVER_NAME.crt" "$LOCALCA_URL/download/$SERVER_NAME/crt"
curl -k -o "$CERT_DIR/$SERVER_NAME.key" "$LOCALCA_URL/download/$SERVER_NAME/key"
curl -k -o "$CERT_DIR/$SERVER_NAME.bundle.crt" "$LOCALCA_URL/download/$SERVER_NAME/bundle"
curl -k -o "$CERT_DIR/ca.crl" "$LOCALCA_URL/download/crl"

# Set proper permissions
chmod 644 "$CERT_DIR/$SERVER_NAME.crt"
chmod 644 "$CERT_DIR/$SERVER_NAME.bundle.crt"
chmod 600 "$CERT_DIR/$SERVER_NAME.key"
chmod 644 "$CERT_DIR/ca.crl"

# Reload NGINX
systemctl reload nginx
```

You can run this script via a cron job to regularly update certificates.

## Security Best Practices

1. **Protect the CA Key**: The CA private key is the most sensitive component. Never share it, and ensure it's properly encrypted.

2. **Regular Backups**: Regularly back up the `data` directory containing all certificates and the CA.

3. **Certificate Lifecycle Management**: Establish a process for certificate issuance, renewal, and revocation.

4. **Internal Network Only**: Never expose the LocalCA service to the public internet.

5. **Access Control**: Restrict access to the LocalCA web interface to administrators only.

6. **Certificate Visibility**: Maintain clear documentation of all issued certificates, their purposes, and expiration dates.

7. **Secure Web Interface**: Always use HTTPS (enabled by default) to access the web interface.

8. **Certificate Revocation**: Keep your CRLs up to date and ensure all services check the CRL.

## Troubleshooting

### Certificate Not Trusted

If your browser shows a certificate warning even after installing the CA:

1. Verify the CA certificate is installed in the correct certificate store
2. Restart your browser
3. Clear browser cache and SSL state
4. Ensure the hostname matches the certificate's Common Name or SAN

### Failed to Generate Certificate

If certificate generation fails:

1. Check the Docker logs: `docker-compose logs localca`
2. Ensure the CA password in `cakey.txt` is correct
3. Verify OpenSSL is working correctly in the container

### CRL Issues

If certificate revocation checking isn't working:

1. Ensure the CRL file is accessible to your web server
2. Verify the CRL is properly formatted: `openssl crl -in ca.crl -text -noout`
3. Make sure the CRL hasn't expired

### Container Won't Start

If the Docker container won't start:

1. Check logs: `docker-compose logs localca`
2. Verify port 8080 and 8443 aren't in use by another service
3. Ensure `cakey.txt` exists and is readable
4. Check for disk space issues