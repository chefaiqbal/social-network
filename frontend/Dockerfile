FROM node:20

WORKDIR /app

# Copy package files
COPY package*.json ./

# Install dependencies
ENV NEXT_TELEMETRY_DISABLED 1
ENV NODE_ENV development

# Don't build during image creation - we'll run npm install and dev when container starts
CMD ["sh", "-c", "npm install && npm run dev -- -H 0.0.0.0"] 