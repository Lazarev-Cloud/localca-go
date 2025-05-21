# cleanup.ps1
# Script to remove unnecessary files from the project

Write-Host "Starting cleanup of unnecessary files..."

# Build artifacts and temporary files
if (Test-Path ".next") {
    Write-Host "Removing .next directory (build artifacts)..."
    Remove-Item -Path ".next" -Recurse -Force
}

# Node modules (can be reinstalled)
if (Test-Path "node_modules") {
    Write-Host "Removing node_modules directory (can be reinstalled)..."
    Remove-Item -Path "node_modules" -Recurse -Force
}

# Coverage and test reports
if (Test-Path "coverage") {
    Write-Host "Removing coverage file..."
    Remove-Item -Path "coverage" -Force
}

if (Test-Path "coverage.out") {
    Write-Host "Removing coverage.out file..."
    Remove-Item -Path "coverage.out" -Force
}

# Binary/executable files (can be rebuilt)
if (Test-Path "localca-go") {
    Write-Host "Removing localca-go binary (can be rebuilt)..."
    Remove-Item -Path "localca-go" -Force
}

if (Test-Path "localca-go.exe") {
    Write-Host "Removing localca-go.exe binary (can be rebuilt)..."
    Remove-Item -Path "localca-go.exe" -Force
}

# Unused handlers file (duplicate of coverage.out)
if (Test-Path "handlers") {
    Write-Host "Removing handlers file (duplicate coverage info)..."
    Remove-Item -Path "handlers" -Force
}

# Security reports (can be regenerated)
if (Test-Path "security-reports") {
    Write-Host "Removing security-reports directory (can be regenerated)..."
    Remove-Item -Path "security-reports" -Recurse -Force
}

# This file appears to be just a test/placeholder
if (Test-Path "cakey.txt") {
    Write-Host "Removing cakey.txt (placeholder file)..."
    Remove-Item -Path "cakey.txt" -Force
}

# Cursor artifacts
if (Test-Path ".cursor") {
    Write-Host "Removing .cursor directory (IDE-specific artifacts)..."
    Remove-Item -Path ".cursor" -Recurse -Force
}

# Check if __mocks__ directory is needed (used for testing)
if (Test-Path "__mocks__") {
    Write-Host "Checking __mocks__ directory..."
    $mockFiles = Get-ChildItem -Path "__mocks__" -Recurse | Measure-Object
    Write-Host "Found $($mockFiles.Count) files in __mocks__"
    
    # Keep this directory as it's small and may be needed for tests
}

# Check for any potential log files
$logFiles = Get-ChildItem -Path "." -Filter "*.log" -Recurse -File
if ($logFiles.Count -gt 0) {
    Write-Host "Found $($logFiles.Count) log files. Removing..."
    foreach ($file in $logFiles) {
        Write-Host "  Removing $($file.FullName)..."
        Remove-Item -Path $file.FullName -Force
    }
}

# Check for any temp files
$tempFiles = Get-ChildItem -Path "." -Include "*.tmp", "*.temp", "*~" -Recurse -File
if ($tempFiles.Count -gt 0) {
    Write-Host "Found $($tempFiles.Count) temporary files. Removing..."
    foreach ($file in $tempFiles) {
        Write-Host "  Removing $($file.FullName)..."
        Remove-Item -Path $file.FullName -Force
    }
}

# Clean up this cleanup script itself when done
Write-Host "Cleanup completed. To remove this cleanup script itself, run:"
Write-Host "Remove-Item -Path 'cleanup.ps1' -Force" 