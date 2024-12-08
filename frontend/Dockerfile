# Frontend Dockerfile
FROM node:20

WORKDIR /app

# Copy package files
COPY package*.json ./

# Install dependencies
RUN npm install

# Copy the source code
COPY . .

# Expose port 3000
EXPOSE 3000

# Command to run the application in development mode
CMD ["npm", "run", "dev"]