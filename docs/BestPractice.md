# Certificate Management and Best Practices

This guide provides best practices for managing your certificates and using them effectively in your organization.

## Certificate Lifecycle Management

### Planning

1. **Certificate Inventory**: Maintain an inventory of all certificates, their purposes, and expiration dates
2. **Service Mapping**: Document which services use which certificates
3. **Naming Conventions**: Establish consistent naming conventions for certificates
4. **Expiration Strategy**: Define renewal procedures at least 30 days before expiration

### Creation

1. **Proper Subject Information**: Include accurate organization and domain information
2. **SAN Certificates**: Use Subject Alternative Names (SANs) instead of multiple separate certificates
3. **Appropriate Validity Period**: 1 year is recommended for server certificates, shorter for high-security services
4. **Strong Keys**: Use at least 2048-bit RSA keys (4096-bit for the CA)
5. **Proper Extensions**: Include appropriate key usage and extended key usage extensions

### Deployment

1. **Secure Key Handling**: Keep private keys protected with proper permissions (0600)
2. **Automated Deployment**: Use scripts to automate certificate deployment to servers
3. **Validation**: Verify certificate installation with tools like SSL Labs (for public servers)
4. **Documentation**: Document where each certificate is deployed

### Monitoring

1. **Expiration Monitoring**: Set up monitoring to alert about upcoming expirations
2. **Certificate Validation**: Regularly check that certificates are properly trusted
3. **CRL Updates**: Ensure CRLs are regularly updated and distributed

### Renewal

1. **Early Renewal**: Renew certificates at least 30 days before expiration
2. **Validation**: Validate renewed certificates before deployment
3. **Key Rotation**: Consider generating new keys during renewal for critical systems

### Revocation

1. **Prompt Revocation**: Immediately revoke compromised or no longer needed certificates
2. **CRL Distribution**: Ensure your CRL is regularly updated and accessible to all services
3. **Documentation**: Document why certificates were revoked

## Security Best Practices

### CA Security

1. **Offline Root CA**: For high-security environments, consider keeping your root CA offline
2. **Limited Access**: Restrict access to the CA management interface
3. **Strong Passwords**: Use strong passwords for the CA key
4. **Regular Backups**: Backup the CA key and certificates regularly
5. **Audit Logging**: Monitor and log all certificate issuance and revocation activities

### Server Certificate Security

1. **Private Key Protection**: Keep private keys secure with proper permissions
2. **No Key Sharing**: Never share private keys between different servers
3. **Regular Rotation**: Regularly rotate keys and certificates (annually or bi-annually)
4. **Secure Storage**: Store keys in secure storage like hardware security modules (HSMs) for critical systems

### Client Certificate Security

1. **Strong P12 Passwords**: Use strong passwords for P12 files
2. **User Education**: Train users on the importance of protecting their certificates
3. **Revocation Process**: Establish a clear process for reporting lost or compromised client certificates
4. **Limited Scope**: Limit the scope of client certificates to only what's necessary

## Certificate Deployment Strategies

### Web Servers

#### NGINX

```nginx
# In http block for global settings
http {
    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_prefer_server_ciphers on;
    ssl_ciphers EECDH+AESGCM:EDH+AESGCM;
    ssl_session_cache shared:SSL:10m;
    ssl_session_timeout 10m;
    
    # Enable OCSP stapling
    ssl_stapling off;  # Not needed for internal CAs
    
    # Common CA certificate for client auth
    ssl_client_certificate /path/to/ca.pem;
    
    # Certificate revocation checking
    ssl_crl /path/to/ca.crl;
    
    # Other http settings...
}

# Server block
server {
    listen 443 ssl;
    server_name your-server.local;
    
    ssl_certificate /path/to/server.bundle.crt;
    ssl_certificate_key /path/to/server.key;
    
    # For client certificate authentication
    ssl_verify_client on;
    ssl_verify_depth 1;
    
    # Pass client certificate info to applications
    proxy_set_header X-SSL-Client-DN $ssl_client_s_dn;
    proxy_set_header X-SSL-Client-Verify $ssl_client_verify;
    
    # Other server settings...
}
```

#### Apache

```apache
# Global SSL settings
SSLProtocol all -SSLv3 -TLSv1 -TLSv1.1
SSLHonorCipherOrder on
SSLCipherSuite EECDH+AESGCM:EDH+AESGCM
SSLSessionCache shmcb:/var/cache/mod_ssl/scache(512000)
SSLSessionTimeout 300

# Virtual host
<VirtualHost *:443>
    ServerName your-server.local
    
    SSLEngine on
    SSLCertificateFile /path/to/server.crt
    SSLCertificateKeyFile /path/to/server.key
    SSLCACertificateFile /path/to/ca.pem
    
    # Certificate revocation checking
    SSLCARevocationPath /path/to/
    SSLCARevocationFile /path/to/ca.crl
    SSLCARevocationCheck chain
    
    # For client certificate authentication
    SSLVerifyClient require
    SSLVerifyDepth 1
    
    # Pass client certificate info to applications
    RequestHeader set X-SSL-Client-DN "%{SSL_CLIENT_S_DN}s"
    RequestHeader set X-SSL-Client-Verify "%{SSL_CLIENT_VERIFY}s"
    
    # Other virtual host settings...
</VirtualHost>
```

### Kubernetes

For Kubernetes environments, you can use cert-manager to automate certificate management. Here's a simple setup using LocalCA as a custom CA:

1. Create a Secret with your CA certificate and key:

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: localca-root
  namespace: cert-manager
type: Opaque
data:
  tls.crt: <base64-encoded-ca-cert>
  tls.key: <base64-encoded-ca-key>
```

2. Create a ClusterIssuer that uses your CA:

```yaml
apiVersion: cert-manager.io/v1
kind: ClusterIssuer
metadata:
  name: localca-issuer
spec:
  ca:
    secretName: localca-root
```

3. Request certificates in your deployments:

```yaml
apiVersion: cert-manager.io/v1
kind: Certificate
metadata:
  name: example-com-cert
  namespace: default
spec:
  secretName: example-com-tls
  duration: 8760h # 1 year
  renewBefore: 720h # 30 days
  subject:
    organizations:
      - Your Organization
  commonName: example.local
  dnsNames:
    - example.local
    - www.example.local
  issuerRef:
    name: localca-issuer
    kind: ClusterIssuer
```

### Java Applications

For Java applications, you need to import certificates into the Java keystore:

```bash
# Import CA certificate to Java truststore
keytool -import -trustcacerts -alias localca -file ca.pem -keystore $JAVA_HOME/lib/security/cacerts -storepass changeit

# Import server certificate and key (from P12) to keystore
keytool -importkeystore -srckeystore server.p12 -srcstoretype PKCS12 -srcstorepass password -destkeystore server.keystore -deststoretype JKS -deststorepass password
```

Java application configuration:

```
# For server certificate
javax.net.ssl.keyStore=/path/to/server.keystore
javax.net.ssl.keyStorePassword=password

# For client certificate trust
javax.net.ssl.trustStore=/path/to/truststore
javax.net.ssl.trustStorePassword=password
```

### Docker/Containerized Environments

For Docker environments, mount certificates as volumes:

```yaml
version: '3'
services:
  webapp:
    image: your-app:latest
    volumes:
      - ./certs/server.crt:/app/certs/server.crt:ro
      - ./certs/server.key:/app/certs/server.key:ro
      - ./certs/ca.pem:/app/certs/ca.pem:ro
    environment:
      - SSL_CERT_PATH=/app/certs/server.crt
      - SSL_KEY_PATH=/app/certs/server.key
      - CA_CERT_PATH=/app/certs/ca.pem
```

## Automation

### Renewal Automation

Create a script to automatically renew certificates approaching expiration:

```bash
#!/bin/bash

# Configuration
LOCALCA_URL="https://localca:8443"
CERT_DIR="/etc/ssl/certs"
DAYS_BEFORE_EXPIRY=30

# Get list of certificates
CERTS=$(curl -sk "${LOCALCA_URL}/" | grep -oP '(?<=<td><a href="/files\?name=)[^"]+')

for CERT in $CERTS; do
    # Get certificate expiry date
    EXPIRY=$(openssl x509 -in "${CERT_DIR}/${CERT}.crt" -noout -enddate | cut -d= -f2)
    EXPIRY_EPOCH=$(date -d "$EXPIRY" +%s)
    NOW_EPOCH=$(date +%s)
    DAYS_LEFT=$(( ($EXPIRY_EPOCH - $NOW_EPOCH) / 86400 ))
    
    if [ $DAYS_LEFT -lt $DAYS_BEFORE_EXPIRY ]; then
        echo "Certificate $CERT expires in $DAYS_LEFT days. Renewing..."
        
        # Renew certificate
        curl -sk -X POST "${LOCALCA_URL}/renew" \
            -d "name=${CERT}" \
            -d "csrf_token=TOKEN"  # You'll need to handle CSRF token
            
        # Download renewed certificate
        curl -sk -o "${CERT_DIR}/${CERT}.crt" "${LOCALCA_URL}/download/${CERT}/crt"
        curl -sk -o "${CERT_DIR}/${CERT}.bundle.crt" "${LOCALCA_URL}/download/${CERT}/bundle"
        
        # Set permissions
        chmod 644 "${CERT_DIR}/${CERT}.crt"
        chmod 644 "${CERT_DIR}/${CERT}.bundle.crt"
        
        # Restart services that use this certificate
        # This depends on your specific services
        if [ "$CERT" == "web-server" ]; then
            systemctl restart nginx
        elif [ "$CERT" == "mail-server" ]; then
            systemctl restart postfix
        fi
        
        echo "Certificate $CERT renewed successfully."
    else
        echo "Certificate $CERT has $DAYS_LEFT days before expiry. No action needed."
    fi
done
```

### Distribution Automation

Create a script to distribute certificates to multiple servers:

```bash
#!/bin/bash

# Configuration
LOCALCA_URL="https://localca:8443"
CERT_NAME="server.local"
SERVERS=("web1.local" "web2.local" "web3.local")
REMOTE_CERT_DIR="/etc/ssl/certs"
REMOTE_USER="admin"

# Download certificates
curl -sk -o "/tmp/${CERT_NAME}.crt" "${LOCALCA_URL}/download/${CERT_NAME}/crt"
curl -sk -o "/tmp/${CERT_NAME}.key" "${LOCALCA_URL}/download/${CERT_NAME}/key"
curl -sk -o "/tmp/${CERT_NAME}.bundle.crt" "${LOCALCA_URL}/download/${CERT_NAME}/bundle"
curl -sk -o "/tmp/ca.crl" "${LOCALCA_URL}/download/crl"

# Distribute to servers
for SERVER in "${SERVERS[@]}"; do
    echo "Deploying certificates to $SERVER..."
    
    # Create directory if it doesn't exist
    ssh ${REMOTE_USER}@${SERVER} "mkdir -p ${REMOTE_CERT_DIR}"
    
    # Copy certificates
    scp "/tmp/${CERT_NAME}.crt" "${REMOTE_USER}@${SERVER}:${REMOTE_CERT_DIR}/"
    scp "/tmp/${CERT_NAME}.key" "${REMOTE_USER}@${SERVER}:${REMOTE_CERT_DIR}/"
    scp "/tmp/${CERT_NAME}.bundle.crt" "${REMOTE_USER}@${SERVER}:${REMOTE_CERT_DIR}/"
    scp "/tmp/ca.crl" "${REMOTE_USER}@${SERVER}:${REMOTE_CERT_DIR}/"
    
    # Set permissions
    ssh ${REMOTE_USER}@${SERVER} "chmod 644 ${REMOTE_CERT_DIR}/${CERT_NAME}.crt"
    ssh ${REMOTE_USER}@${SERVER} "chmod 644 ${REMOTE_CERT_DIR}/${CERT_NAME}.bundle.crt"
    ssh ${REMOTE_USER}@${SERVER} "chmod 600 ${REMOTE_CERT_DIR}/${CERT_NAME}.key"
    ssh ${REMOTE_USER}@${SERVER} "chmod 644 ${REMOTE_CERT_DIR}/ca.crl"
    
    # Reload web server (example for NGINX)
    ssh ${REMOTE_USER}@${SERVER} "systemctl reload nginx"
    
    echo "Deployment to $SERVER completed."
done

# Clean up
rm "/tmp/${CERT_NAME}.crt" "/tmp/${CERT_NAME}.key" "/tmp/${CERT_NAME}.bundle.crt" "/tmp/ca.crl"

echo "Certificate distribution completed."
```

## Monitoring and Alerting

### Set Up Certificate Expiry Monitoring

You can use monitoring tools like Prometheus, Nagios, or simple scripts to monitor certificate expiration:

#### Prometheus/Blackbox Exporter

For Prometheus with Blackbox Exporter:

```yaml
scrape_configs:
  - job_name: 'ssl_cert_check'
    metrics_path: /probe
    params:
      module: [http_2xx]  # Use the HTTP probe
    static_configs:
      - targets:
        - https://server1.local
        - https://server2.local
    relabel_configs:
      - source_labels: [__address__]
        target_label: __param_target
      - source_labels: [__param_target]
        target_label: instance
      - target_label: __address__
        replacement: blackbox_exporter:9115  # Blackbox exporter's address
```

Alert rule:

```yaml
groups:
- name: ssl_cert_alerts
  rules:
  - alert: SSLCertExpiringSoon
    expr: probe_ssl_earliest_cert_expiry - time() < 86400 * 30
    for: 1h
    labels:
      severity: warning
    annotations:
      summary: "SSL certificate expiring soon for {{ $labels.instance }}"
      description: "SSL certificate for {{ $labels.instance }} will expire in less than 30 days"
```

#### Nagios/Check_MK

Nagios check command:

```
define command {
    command_name check_ssl_cert
    command_line $USER1$/check_ssl_cert -H $ARG1$ -w 30 -c 15
}

define service {
    use                  generic-service
    host_name            server1
    service_description  SSL Certificate
    check_command        check_ssl_cert!server1.local
}
```

#### Simple Shell Script for Cron

```bash
#!/bin/bash

# Configuration
CERT_DIR="/etc/ssl/certs"
WARN_DAYS=30
ALERT_EMAIL="admin@example.com"

# Check all certificates
for CERT in "${CERT_DIR}"/*.crt; do
    # Skip CA certificate and bundle certificates
    if [[ "$CERT" == *"ca.crt"* || "$CERT" == *"bundle"* ]]; then
        continue
    fi
    
    # Get certificate expiry date
    EXPIRY=$(openssl x509 -in "$CERT" -noout -enddate | cut -d= -f2)
    EXPIRY_EPOCH=$(date -d "$EXPIRY" +%s)
    NOW_EPOCH=$(date +%s)
    DAYS_LEFT=$(( ($EXPIRY_EPOCH - $NOW_EPOCH) / 86400 ))
    CERT_NAME=$(basename "$CERT")
    
    if [ $DAYS_LEFT -lt $WARN_DAYS ]; then
        # Send alert email
        echo "Certificate $CERT_NAME expires in $DAYS_LEFT days" | \
        mail -s "Certificate Expiration Warning: $CERT_NAME" "$ALERT_EMAIL"
    fi
done
```

## Advanced Topics

### Certificate Pinning

For critical applications, consider implementing certificate pinning:

#### Android Example

```java
// Network security configuration (res/xml/network_security_config.xml)
<?xml version="1.0" encoding="utf-8"?>
<network-security-config>
    <domain-config>
        <domain includeSubdomains="true">example.com</domain>
        <pin-set>
            <!-- Pin the CA certificate -->
            <pin digest="SHA-256">base64EncodedPinOfYourCACert==</pin>
        </pin-set>
    </domain-config>
</network-security-config>
```

#### iOS Example (Swift)

```swift
let trustManager = ServerTrustManager(evaluators: [
    "example.com": PinnedCertificatesTrustEvaluator(
        certificates: [ourBundledCACertificate],
        acceptSelfSignedCertificates: false,
        performDefaultValidation: true,
        validateHost: true
    )
])

let session = Session(serverTrustManager: trustManager)
```

#### NGINX HPKP Example (for public sites)

```nginx
add_header Public-Key-Pins 'pin-sha256="base64EncodedPin1"; pin-sha256="base64EncodedPin2"; max-age=5184000; includeSubDomains';
```

### Mutual TLS Authentication (mTLS)

For securing service-to-service communication:

#### NGINX Server Configuration

```nginx
server {
    listen 443 ssl;
    
    ssl_certificate /path/to/server.crt;
    ssl_certificate_key /path/to/server.key;
    ssl_client_certificate /path/to/ca.pem;
    ssl_verify_client on;
    
    location / {
        if ($ssl_client_verify != "SUCCESS") {
            return 403;
        }
        
        # Optional: restrict to specific client certificates
        if ($ssl_client_s_dn !~ "CN=allowed-client") {
            return 403;
        }
        
        proxy_pass http://backend;
    }
}
```

#### Client Configuration (curl example)

```bash
curl --cert client.crt --key client.key --cacert ca.pem https://server.local
```

#### Client Configuration (Python example)

```python
import requests

response = requests.get(
    'https://server.local',
    cert=('/path/to/client.crt', '/path/to/client.key'),
    verify='/path/to/ca.pem'
)
```

### Certificate Transparency for Internal PKI

While public CAs use Certificate Transparency (CT) logs, you can implement a simple log for your internal PKI:

```sql
CREATE TABLE certificate_log (
    id SERIAL PRIMARY KEY,
    common_name VARCHAR(255) NOT NULL,
    serial_number VARCHAR(64) NOT NULL,
    issuer VARCHAR(255) NOT NULL,
    valid_from TIMESTAMP NOT NULL,
    valid_to TIMESTAMP NOT NULL,
    subject_alternative_names TEXT,
    issued_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    issued_by VARCHAR(255),
    certificate_data TEXT,
    UNIQUE (serial_number)
);
```

Integrate this with your LocalCA to log all certificate issuance.

## Disaster Recovery

### CA Compromise

If your CA is compromised:

1. **Immediate Actions**:
   - Disconnect the CA server from the network
   - Revoke all certificates issued by the CA
   - Notify all users and systems
   
2. **Recovery**:
   - Create a new CA with a new key pair
   - Issue new certificates to all systems
   - Distribute the new CA certificate to all clients
   - Update all configurations to use new certificates

### Backup and Restore

Regularly back up your CA data:

```bash
# Backup
tar -czf localca-backup-$(date +%Y%m%d).tar.gz /path/to/data

# Encrypt the backup
gpg -e -r admin@example.com localca-backup-$(date +%Y%m%d).tar.gz

# Store in a secure location
rsync -av localca-backup-*tar.gz.gpg backup-server:/secure-backups/
```

Restore procedure:

1. Stop the LocalCA service:
```bash
docker-compose down
```

2. Restore the data:
```bash
gpg -d localca-backup-20250101.tar.gz.gpg | tar -xzf - -C /path/to/restore
```

3. Restart the service:
```bash
docker-compose up -d
```

## Reference Architecture

For larger environments, consider this reference architecture:

1. **Offline Root CA**:
   - Air-gapped system for highest security
   - Used only to create and sign intermediate CAs
   - Very long validity period (10+ years)

2. **Intermediate CAs**:
   - Online but secured systems
   - Different intermediates for different purposes:
     - Server certificates
     - Client certificates
     - Code signing
   - Medium validity period (5 years)

3. **Certificate Management System** (LocalCA):
   - Web interface for certificate issuance and revocation
   - Automated certificate deployment
   - Certificate lifecycle management

4. **Validation Infrastructure**:
   - CRL distribution points
   - OCSP responders (for public-facing services)
   - Certificate validation services

This architecture provides a balance of security and operational efficiency for medium to large organizations.