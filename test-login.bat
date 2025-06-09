@echo off
echo Testing LocalCA Login Functionality...
echo.

echo 1. Testing backend health...
curl -s http://localhost:8080/health
echo.

echo 2. Testing setup status...
curl -s http://localhost:8080/api/setup
echo.

echo 3. Testing login with JSON...
curl -s -X POST http://localhost:8080/api/login ^
  -H "Content-Type: application/json" ^
  -d "{\"username\":\"admin\",\"password\":\"12345678\"}"
echo.

echo 4. Testing login with form data...
curl -s -X POST http://localhost:8080/api/login ^
  -H "Content-Type: application/x-www-form-urlencoded" ^
  -d "username=admin&password=12345678"
echo.

echo 5. Testing frontend proxy...
curl -s -X POST http://localhost:3000/api/proxy/api/login ^
  -H "Content-Type: application/json" ^
  -d "{\"username\":\"admin\",\"password\":\"12345678\"}"
echo.

echo Test completed!
pause 