@echo off
echo Running LocalCA-Go Tests
echo ===============================

echo Running tests with coverage...
go test -v -cover ./pkg/...

if %ERRORLEVEL% EQU 0 (
    echo All tests passed!
    exit /b 0
) else (
    echo Tests failed!
    exit /b 1
) 