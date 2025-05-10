# ACME Integration Guide for LocalCA

This guide covers how to set up and use the ACME protocol with your LocalCA deployment to automate certificate management across your internal network.

## What is ACME?

The Automated Certificate Management Environment (ACME) protocol is an industry-standard protocol for automating the issuance and renewal of SSL/TLS certificates. It's the same protocol used by Let's Encrypt for public certificates, but LocalCA implements it for your private network.

## Benefits of Using ACME

- **Automation**: Certificates are automatically renewed before expiry
- **Reduced Manual Work**: No need to manually generate, distribute, and install certificates
- **Standardization**: Works with many existing tools and platforms
- **Security**: Improves security by ensuring up-to-date certificates

## Setup Guide

### 1. Deploy LocalCA with ACME Support

Update your `docker-compose.yml` to include the ACME configuration:

```yaml
services:
  localca:
    # ... existing configuration ...
    environment:
      # ... existing environment variables ...
      - ACME_ENABLED=true
      - ACME_BASE_URL=https://ca.your-domain.local
    ports:
      - "80:80"      # Required for HTTP-01 challenges
      - "443:443"    # Required for TLS-ALPN-01 challenges (optional)
```

### 2. Enable ACME in the Web Interface

1. Start LocalCA
2. Navigate to the web interface (e.g., `https://ca.your-domain.local:8443`)
3. Click on "ACME Protocol Settings"
4. Click "Enable ACME Protocol"
5. Enter the base URL where your LocalCA is accessible from your internal network
6. Click "Enable"

### 3. Install the CA Certificate on Client Servers

Before clients can use ACME, they need to trust your CA:

**Debian/Ubuntu:**
```bash
# Download the CA certificate
curl -k https://ca.your-domain.local:8443/download/ca -o /usr/local/share/ca-certificates/localca.crt
sudo update-ca-certificates
```

**CentOS/RHEL:**
```bash
# Download the CA certificate
curl -k https://ca.your-domain.local:8443/download/ca -o /etc/pki/ca-trust/source/anchors/localca.crt
sudo update-ca-trust
```

### 4. Configure ACME Clients

#### Using Certbot

[Certbot](https://certbot.eff.org/) is a popular ACME client that works well with LocalCA:

```bash
# Install Certbot
sudo apt-get install certbot

# Request a certificate
sudo certbot certonly --standalone \
  --server https://ca.your-domain.local/acme/directory \
  --domain your-service.your-domain.local

# Test automatic renewal
sudo certbot renew --dry-run
```

#### Using acme.sh

[acme.sh](https://github.com/acmesh-official/acme.sh) is a lightweight ACME client with minimal dependencies:

```bash
# Install acme.sh
curl https://get.acme.sh | sh -s email=your-email@example.com

# Download and use the CA certificate
curl -k https://ca.your-domain.local:8443/download/ca -o ca.pem
export ACME_CA_CERT=`pwd`/ca.pem

# Issue a certificate with HTTP validation
~/.acme.sh/acme.sh --issue \
  --server https://ca.your-domain.local/acme/directory \
  -d your-service.your-domain.local \
  -w /var/www/html
```

#### Using Traefik

[Traefik](https://traefik.io/) is a modern reverse proxy that can automatically obtain and renew certificates:

```yaml
# docker-compose.yml
version: '3'

services:
  traefik:
    image: traefik:v2.5
    command:
      - "--providers.docker=true"
      - "--entrypoints.web.address=:80"
      - "--entrypoints.websecure.address=:443"
      - "--certificatesresolvers.localca.acme.httpchallenge=true"
      - "--certificatesresolvers.localca.acme.httpchallenge.entrypoint=web"
      - "--certificatesresolvers.localca.acme.caserver=https://ca.your-domain.local/acme/directory"
      - "--certificatesresolvers.localca.acme.email=your-email@example.com"
      - "--certificatesresolvers.localca.acme.storage=/etc/traefik/acme.json"
    volumes:
      - "/var/run/docker.sock:/var/run/docker.sock:ro"
      - "./ca.pem:/etc/ssl/certs/localca.pem:ro"
      - "./acme.json:/etc/traefik/acme.json"
    environment:
      - "LEGO_CA_CERTIFICATES=/etc/ssl/certs/localca.pem"
      - "SSL_CERT_FILE=/etc/ssl/certs/localca.pem"
    ports:
      - "80:80"
      - "443:443"
```

#### Using Caddy

[Caddy](https://caddyserver.com/) is a web server with automatic HTTPS:

```
# Caddyfile
{
  acme_ca https://ca.your-domain.local/acme/directory
  acme_ca_root /path/to/ca.pem
}

your-service.your-domain.local {
  respond "Hello, World!"
}
```

## ACME Challenges Explained

LocalCA supports multiple types of challenges to verify domain ownership:

### HTTP-01 Challenge

This challenge requires the ability to serve a specific file at a well-known URL on your domain. The ACME client will:

1. Receive a token from the ACME server
2. Create a file at `/.well-known/acme-challenge/{token}` containing a key authorization value
3. LocalCA will send an HTTP request to this URL to verify control of the domain

Requirements:
- Port 80 must be accessible from LocalCA to the target server
- The server must be able to serve files under the `.well-known/acme-challenge/` path

### DNS-01 Challenge

This challenge requires the ability to add a TXT record to your DNS zone. The ACME client will:

1. Receive a token from the ACME server
2. Create a TXT record for `_acme-challenge.your-domain.local` with the key authorization digest
3. LocalCA will query this DNS record to verify control of the domain

Requirements:
- DNS server must be accessible from LocalCA
- You need the ability to add TXT records to your DNS zones

### TLS-ALPN-01 Challenge

This challenge uses TLS extensions to verify domain control. The ACME client will:

1. Receive a token from the ACME server
2. Configure a temporary TLS certificate with a special ACME validation extension
3. LocalCA will connect to the domain on port 443 using the "acme-tls/1" ALPN protocol

Requirements:
- Port 443 must be accessible from LocalCA to the target server
- The server must support ALPN and be able to present a custom certificate

## Troubleshooting

### Certificate Issuance Failures

1. **Check Challenge Accessibility**
   - For HTTP-01: Verify that LocalCA can reach the target server on port 80
   - For DNS-01: Verify that the DNS record is visible from LocalCA
   - For TLS-ALPN-01: Verify that LocalCA can reach the target server on port 443

2. **Check CA Trust**
   - Ensure the client server trusts the LocalCA root certificate
   - Verify with: `curl -v --cacert /path/to/ca.pem https://ca.your-domain.local/acme/directory`

3. **Check Logs**
   - Review LocalCA logs for errors: `docker-compose logs localca`
   - Review ACME client logs for detailed error messages

### Certificate Renewal Issues

1. **Automatic Renewal Not Working**
   - Verify that the renewal cronjob or timer is active
   - For Certbot: `systemctl status certbot.timer`
   - For acme.sh: Check crontab with `crontab -l`

2. **Network Changes**
   - Ensure that network routes between LocalCA and clients haven't changed
   - Check firewall rules to ensure ports 80/443 are still open

## Best Practices

1. **Monitor Certificate Expiry**
   - Set up additional monitoring for certificate expiry as a safety net
   - Use tools like Prometheus with ssl_exporter or certcheck

2. **Backup Your CA**
   - Regularly backup your LocalCA data, especially the CA private key
   - Use the `make backup` command to create backups

3. **Test Renewals Periodically**
   - Periodically test your renewal process with dry runs
   - For Certbot: `certbot renew --dry-run`

4. **Use Staging Environment for Testing**
   - When testing new ACME clients, consider setting up a staging LocalCA instance
   - This prevents potential issues from affecting your production certificates

## Advanced Topics

### ACME Account Management

ACME clients create accounts on the ACME server. These accounts are used to track certificate requests and manage rate limits. You can view active ACME accounts in the LocalCA web interface under "ACME Protocol Settings".

### Custom Challenge Validation

In some network environments, you might need to implement custom challenge validation:

- **Split DNS**: If your internal DNS is separate from your external DNS, ensure that LocalCA can resolve internal domains correctly
- **Proxy Challenges**: In some cases, you might need to proxy challenge validation from a DMZ server to an internal server

### High Availability Setup

For high availability, you can:

1. Deploy multiple LocalCA instances
2. Share the CA private key securely between instances
3. Use a load balancer in front of the ACME endpoints
4. Ensure all instances have access to the same certificate storage

## Support and Resources

- **LocalCA Documentation**: https://github.com/Lazarev-Cloud/localca
- **ACME Protocol Specification**: [RFC 8555](https://datatracker.ietf.org/doc/html/rfc8555)
- **Certbot Documentation**: https://certbot.eff.org/docs/
- **acme.sh Documentation**: https://github.com/acmesh-official/acme.sh
