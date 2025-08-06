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
echo LocalCA API service started! 
echo API: http://localhost:8080
echo Health: http://localhost:8080/api/health
echo API Docs: http://localhost:8080/api/docs
echo.
pause 