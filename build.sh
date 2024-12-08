#!/bin/bash

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
NC='\033[0m' # No Color

echo -e "${GREEN}Starting build process...${NC}"

# Stop any running containers
echo -e "${GREEN}Stopping any running containers...${NC}"
docker-compose down

# Clean up old images
echo -e "${GREEN}Cleaning up old images...${NC}"
docker system prune -f

# Build and start the containers
echo -e "${GREEN}Building and starting containers...${NC}"
docker-compose up --build -d

# Check if containers are running
if [ "$(docker ps -q -f name=social-network-backend)" ] && [ "$(docker ps -q -f name=social-network-frontend)" ]; then
    echo -e "${GREEN}Build successful!${NC}"
    echo -e "${GREEN}Frontend is running at: http://localhost:3000${NC}"
    echo -e "${GREEN}Backend is running at: http://localhost:8080${NC}"
    
    # Show logs
    echo -e "${GREEN}Showing container logs (Ctrl+C to exit)...${NC}"
    docker-compose logs -f
else
    echo -e "${RED}Build failed! Check the logs above for errors.${NC}"
    exit 1
fi 