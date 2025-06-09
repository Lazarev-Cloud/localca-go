# LocalCA Setup Guide

## Overview
This guide walks you through the complete setup process for LocalCA, from initial installation to user login.

## Setup Process

### Step 1: Reset System for Fresh Setup
The system has been reset to require initial setup. The auth.json file now contains:
```json
{
  "admin_username": "admin",
  "admin_password_hash": "",
  "setup_completed": false,
  "setup_token": "RESET_SETUP_TOKEN_2025",
  "setup_token_expiry": "2025-06-10T23:59:59.000Z"
}
```

### Step 2: Access Setup Page
1. **Start the containers** (if not already running):
   ```bash
   .\restart-containers.bat
   ```

2. **Visit the setup page**:
   - Open browser to: http://localhost:3000/setup
   - The system will automatically detect setup is required

### Step 3: Complete Setup
1. **Fill in the setup form**:
   - **Username**: `admin` (pre-filled)
   - **Password**: Choose your password (e.g., `12345678`)
   - **Confirm Password**: Repeat the password
   - **Setup Token**: `RESET_SETUP_TOKEN_2025` (should auto-fill)

2. **Submit the form**
   - Click "Complete Setup"
   - System will hash your password and save configuration

### Step 4: Login
1. **Redirect to login**: After setup, you'll be redirected to login page
2. **Use your credentials**:
   - **Username**: `admin`
   - **Password**: The password you set during setup

## Testing the Setup Process

### Manual Testing
1. **Test setup endpoint**:
   ```bash
   curl http://localhost:8080/api/setup
   ```

2. **Complete setup via API**:
   ```bash
   curl -X POST http://localhost:8080/api/setup \
     -H "Content-Type: application/json" \
     -d '{"username":"admin","password":"12345678","setup_token":"RESET_SETUP_TOKEN_2025"}'
   ```

3. **Test login**:
   ```bash
   curl -X POST http://localhost:8080/api/login \
     -H "Content-Type: application/json" \
     -d '{"username":"admin","password":"12345678"}'
   ```

### Automated Testing
Run the test script:
```bash
.\test_setup.bat
```

## Debugging

### Backend Logs
The enhanced setup process now provides detailed logging:

```
=== SETUP REQUEST DEBUG ===
Method: POST
Content-Type: application/json
Setup request - Username: admin, Password length: 8, Token: RESET_SETUP_TOKEN_2025
Current auth config - Setup completed: false, Token: RESET_SETUP_TOKEN_2025
Validating setup token - Provided: RESET_SETUP_TOKEN_2025, Expected: RESET_SETUP_TOKEN_2025
Setup token validated successfully
Completing setup for user: admin

=== COMPLETE SETUP DEBUG ===
Username: admin
Password length: 8
Password hashed successfully
Password hash: $2a$10$[hash]
Loaded auth config - Current setup completed: false
Updated config - Username: admin, Setup completed: true
Auth config saved successfully
Verification - Username: admin, Setup completed: true
Verification - Password hash: $2a$10$[hash]
✅ Password hash verification SUCCESS
Setup completed successfully for user: admin
```

### Login Logs
After setup, login attempts will show:

```
=== LOGIN REQUEST DEBUG ===
Method: POST
Content-Type: application/json
Successfully bound JSON data
Final parsed data - Username: 'admin', Password length: 8
Processing login for username: admin
Validating credentials for user: admin
Expected username: admin
Password hash in config: $2a$10$[hash]
Credentials validated successfully for user: admin
Login successful for user: admin, session: [token]
```

## Troubleshooting

### Setup Token Issues
- **Invalid token**: Check that you're using `RESET_SETUP_TOKEN_2025`
- **Expired token**: Token expires on 2025-06-10T23:59:59.000Z
- **Already completed**: If setup is already done, reset auth.json

### Password Issues
- **Hash mismatch**: The system now properly hashes passwords during setup
- **Login fails**: Ensure you're using the exact password set during setup
- **Case sensitivity**: Passwords are case-sensitive

### Frontend Issues
- **Setup page not loading**: Check that containers are running
- **API errors**: Check backend logs for detailed error messages
- **Redirect issues**: Clear browser cache and cookies

## Settings and User Management

### After Login
Once logged in, you can access:

1. **Dashboard**: http://localhost:3000/
   - View certificates
   - System statistics
   - CA information

2. **Settings**: http://localhost:3000/settings
   - Email configuration
   - System settings
   - User preferences

3. **Certificate Management**:
   - Create new certificates
   - Renew existing certificates
   - Revoke certificates
   - Download CA certificate

### User Parameters
The system stores user configuration in:
- **Session data**: Temporary login sessions
- **Auth config**: Username and password hash
- **Settings**: Email, notifications, system preferences

## Security Notes

### Password Security
- Passwords are hashed using bcrypt with default cost (10)
- Session tokens are cryptographically secure (32 bytes)
- Session files have restricted permissions (0600)

### Setup Token Security
- Setup tokens expire after 24 hours
- Tokens are cleared after successful setup
- Only one setup process allowed per installation

### Session Management
- Sessions expire after 8 hours of inactivity
- Session cleanup runs automatically
- Secure cookie settings (HttpOnly, SameSite)

## Next Steps

1. **Complete setup** using the web interface or API
2. **Test login** with your credentials
3. **Configure settings** in the settings page
4. **Create certificates** for your services
5. **Set up email notifications** (optional)

The system is now ready for production use with proper authentication and security measures in place. 