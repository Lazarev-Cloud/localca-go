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
Write-Host "Checking API logs for new debugging info..." -ForegroundColor Yellow
docker-compose logs localca --tail=10

Write-Host ""
Write-Host "========================================" -ForegroundColor Green
Write-Host "   LocalCA API service started!" -ForegroundColor Green
Write-Host "========================================" -ForegroundColor Green
Write-Host ""
Write-Host "API is now available at http://localhost:8080" -ForegroundColor Cyan
Write-Host "Health check: http://localhost:8080/api/health" -ForegroundColor White
Write-Host "API docs: http://localhost:8080/api/docs" -ForegroundColor White
Write-Host ""
Write-Host "The API now has enhanced debugging and" -ForegroundColor Gray
Write-Host "comprehensive RESTful endpoints." -ForegroundColor Gray
Write-Host ""
Read-Host "Press Enter to continue" 