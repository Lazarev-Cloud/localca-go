#!/bin/bash

# Test script for LocalCA application
echo "=== LocalCA Application Test Suite ==="
echo

# Test 1: Frontend accessibility
echo "1. Testing frontend accessibility..."
FRONTEND_STATUS=$(curl -s -o /dev/null -w "%{http_code}" http://localhost:3000)
if [ "$FRONTEND_STATUS" = "200" ]; then
    echo "✅ Frontend is accessible (HTTP $FRONTEND_STATUS)"
else
    echo "❌ Frontend is not accessible (HTTP $FRONTEND_STATUS)"
fi
echo

# Test 2: Backend accessibility
echo "2. Testing backend accessibility..."
BACKEND_STATUS=$(curl -s -o /dev/null -w "%{http_code}" http://localhost:8080/api/setup)
if [ "$BACKEND_STATUS" = "200" ]; then
    echo "✅ Backend is accessible (HTTP $BACKEND_STATUS)"
else
    echo "❌ Backend is not accessible (HTTP $BACKEND_STATUS)"
fi
echo

# Test 3: Setup status
echo "3. Checking setup status..."
SETUP_RESPONSE=$(curl -s http://localhost:8080/api/setup)
SETUP_COMPLETED=$(echo "$SETUP_RESPONSE" | jq -r '.data.setup_completed // false')
if [ "$SETUP_COMPLETED" = "true" ]; then
    echo "✅ Setup is completed"
else
    echo "⚠️  Setup is not completed"
    SETUP_TOKEN=$(echo "$SETUP_RESPONSE" | jq -r '.data.setup_token // ""')
    echo "   Setup token: $SETUP_TOKEN"
fi
echo

# Test 4: Authentication requirement
echo "4. Testing authentication requirement..."
AUTH_RESPONSE=$(curl -s http://localhost:8080/api/ca-info)
AUTH_MESSAGE=$(echo "$AUTH_RESPONSE" | jq -r '.message // ""')
if [[ "$AUTH_MESSAGE" == *"Authentication required"* ]]; then
    echo "✅ Authentication is properly enforced"
else
    echo "❌ Authentication is not enforced properly"
    echo "   Response: $AUTH_RESPONSE"
fi
echo

# Test 5: CORS headers
echo "5. Testing CORS headers..."
CORS_RESPONSE=$(curl -s -I -H "Origin: http://localhost:3000" http://localhost:8080/api/setup)
if echo "$CORS_RESPONSE" | grep -i "access-control-allow-origin" > /dev/null; then
    echo "✅ CORS headers are present"
else
    echo "❌ CORS headers are missing"
fi
echo

# Test 6: Frontend proxy functionality
echo "6. Testing frontend proxy functionality..."
PROXY_RESPONSE=$(curl -s http://localhost:3000/api/setup)
PROXY_SUCCESS=$(echo "$PROXY_RESPONSE" | jq -r '.success // false')
if [ "$PROXY_SUCCESS" = "true" ]; then
    echo "✅ Frontend proxy is working"
else
    echo "❌ Frontend proxy is not working properly"
    echo "   Response: $PROXY_RESPONSE"
fi
echo

# Test 7: Container health
echo "7. Checking container health..."
docker-compose ps --format "table {{.Name}}\t{{.Status}}" 2>/dev/null | grep -v "NAME"
echo

echo "=== Test Summary ==="
echo "Frontend URL: http://localhost:3000"
echo "Backend URL:  http://localhost:8080"
echo "Setup completed: $SETUP_COMPLETED"
echo
echo "🔍 To access the application, visit: http://localhost:3000"