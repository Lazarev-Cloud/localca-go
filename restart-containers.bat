@echo off
echo Restarting LocalCA containers...

echo Stopping containers...
docker-compose down

echo Building and starting containers...
docker-compose up --build -d

echo Waiting for services to start...
timeout /t 10 /nobreak > nul

echo Checking container status...
docker-compose ps

echo.
echo Containers restarted! 
echo Frontend: http://localhost:3000
echo Backend: http://localhost:8080
echo.
echo Login with:
echo Username: admin
echo Password: 12345678
echo.
pause 