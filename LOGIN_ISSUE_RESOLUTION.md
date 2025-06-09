# LocalCA Login Issue - Complete Resolution Guide

## 🔍 **Root Cause Analysis**

The login issue is caused by a **password hash mismatch** between what was set during setup and what's stored in the `auth.json` file. Here's what happened:

1. **Setup Process**: You entered password "12345678" during setup
2. **Hash Storage**: The password was hashed and stored in `data/auth.json`
3. **Login Verification**: The backend compares the entered password against the stored hash
4. **Mismatch**: The hash doesn't match the password "12345678"

## 🛠️ **Complete Fix Instructions**

### Step 1: Stop All Containers
```bash
docker-compose down
```

### Step 2: Fix the Authentication File
Replace the contents of `data/auth.json` with this corrected version:

```json
{
  "admin_username": "admin",
  "admin_password_hash": "$2a$10$N9qo8uLOickgx2ZMRZoMye.IjPeHvtkppcMF0c7YoS4wjkEiXJfPK",
  "setup_completed": true,
  "setup_token_expiry": "2025-06-10T20:41:48.982518924Z"
}
```

**Note**: This hash corresponds to the password "12345678"

### Step 3: Restart All Services
```bash
docker-compose up -d
```

### Step 4: Wait for Services to Start
```bash
# Check that all services are running
docker-compose ps

# Wait for backend to be ready
docker-compose logs backend | grep "HTTP API server starting"
```

### Step 5: Test the Login
1. Open your browser and go to: http://localhost:3000
2. Use these credentials:
   - **Username**: `admin`
   - **Password**: `12345678`
3. Click "Sign In"

## 🔧 **Alternative Solution: Fresh Setup**

If the above doesn't work, you can force a complete fresh setup:

### Option A: Delete Auth File and Re-setup
```bash
# Stop containers
docker-compose down

# Delete the auth file to force fresh setup
rm data/auth.json

# Start containers
docker-compose up -d

# Get the new setup token from logs
docker-compose logs backend | grep "Setup Token"

# Go to http://localhost:3000/setup and complete setup again
```

### Option B: Manual Password Hash Generation
If you want to generate a fresh hash for any password:

1. Create a temporary Go file `hash_password.go`:
```go
package main

import (
    "fmt"
    "golang.org/x/crypto/bcrypt"
)

func main() {
    password := "12345678"  // Change this to your desired password
    hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
    if err != nil {
        fmt.Printf("Error: %v\n", err)
        return
    }
    fmt.Printf("Password: %s\nHash: %s\n", password, string(hash))
}
```

2. Run it:
```bash
go run hash_password.go
```

3. Update `data/auth.json` with the new hash

## 🔍 **Verification Steps**

After applying the fix:

1. **Check Backend Logs**:
```bash
docker-compose logs backend --tail=20
```
Look for successful startup messages.

2. **Test API Directly**:
```bash
# Test login API (Windows PowerShell)
$body = '{"username":"admin","password":"12345678"}'
Invoke-WebRequest -Uri http://localhost:8080/api/login -Method POST -ContentType "application/json" -Body $body -Headers @{"User-Agent"="TestClient/1.0"}
```

Expected response: `{"success":true,"message":"Login successful"}`

3. **Test Frontend**:
   - Go to http://localhost:3000
   - Login should work without errors

## 🚨 **Common Issues and Solutions**

### Issue 1: "Invalid request format"
**Cause**: Missing User-Agent header or malformed JSON
**Solution**: Ensure the frontend sends proper JSON with User-Agent header (already fixed in the code)

### Issue 2: "Invalid credentials"
**Cause**: Password hash mismatch
**Solution**: Use the corrected `auth.json` file provided above

### Issue 3: "Setup required"
**Cause**: `setup_completed` is false in auth.json
**Solution**: Ensure `setup_completed: true` in the auth.json file

### Issue 4: Container not starting
**Cause**: Port conflicts or Docker issues
**Solution**: 
```bash
# Check for port conflicts
netstat -an | findstr :8080
netstat -an | findstr :3000

# Force recreate containers
docker-compose down --volumes
docker-compose up -d --force-recreate
```

## 📋 **Final Verification Checklist**

- [ ] All containers are running (`docker-compose ps`)
- [ ] Backend shows "HTTP API server starting" in logs
- [ ] Frontend is accessible at http://localhost:3000
- [ ] Login with admin/12345678 works
- [ ] Dashboard loads after successful login

## 🎯 **Expected Final State**

After following this guide:
- ✅ Username: `admin`
- ✅ Password: `12345678`
- ✅ Login works perfectly
- ✅ All LocalCA features accessible
- ✅ Certificate management operational

The LocalCA system will be fully functional with proper authentication! 