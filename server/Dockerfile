FROM golang:1.22

WORKDIR /app

# Install air for hot reloading (specific version compatible with Go 1.22)
RUN go install github.com/cosmtrek/air@v1.49.0

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download all dependencies
RUN go mod download

# Copy the air config
COPY .air.toml ./

# Copy the source code
COPY . .

# Expose port 8080
EXPOSE 8080

# Run with air for development
CMD ["air", "-c", ".air.toml"]

