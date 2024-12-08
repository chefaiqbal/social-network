@echo off
echo Starting build process...

REM Stop any running containers
echo Stopping any running containers...
docker-compose down

REM Clean up old images
echo Cleaning up old images...
docker system prune -f

REM Build and start the containers
echo Building and starting containers...
docker-compose up --build -d

REM Check if containers are running
docker ps | findstr "social-network-backend" > nul
if %errorlevel% equ 0 (
    docker ps | findstr "social-network-frontend" > nul
    if %errorlevel% equ 0 (
        echo Build successful!
        echo Frontend is running at: http://localhost:3000
        echo Backend is running at: http://localhost:8080
        
        REM Show logs
        echo Showing container logs (Ctrl+C to exit)...
        docker-compose logs -f
    ) else (
        echo Build failed! Check the logs above for errors.
        exit /b 1
    )
) else (
    echo Build failed! Check the logs above for errors.
    exit /b 1
) 