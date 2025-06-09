Write-Host "Restarting LocalCA containers to apply login fixes..." -ForegroundColor Green
Write-Host ""

Write-Host "Stopping containers..." -ForegroundColor Yellow
docker-compose down

Write-Host ""
Write-Host "Starting containers..." -ForegroundColor Yellow
docker-compose up -d

Write-Host ""
Write-Host "Checking container status..." -ForegroundColor Yellow
docker-compose ps

Write-Host ""
Write-Host "Checking backend logs for new debugging info..." -ForegroundColor Yellow
docker-compose logs backend --tail=10

Write-Host ""
Write-Host "========================================" -ForegroundColor Green
Write-Host "   Login fixes have been applied!" -ForegroundColor Green
Write-Host "========================================" -ForegroundColor Green
Write-Host ""
Write-Host "Now try logging in at http://localhost:3000" -ForegroundColor Cyan
Write-Host "Username: admin" -ForegroundColor White
Write-Host "Password: 12345678" -ForegroundColor White
Write-Host ""
Write-Host "The backend now has enhanced debugging and" -ForegroundColor Gray
Write-Host "will accept both JSON and form data." -ForegroundColor Gray
Write-Host ""
Read-Host "Press Enter to continue" 