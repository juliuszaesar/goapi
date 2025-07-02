# Start from the official golang image
FROM golang:1.23-alpine AS build

# Install git (needed for some Go modules)
RUN apk add --no-cache git

# Set the current working directory inside the container
WORKDIR /app

# Copy go.mod and go.sum files to the working directory
COPY go.mod go.sum ./

# Download all dependencies. Dependencies will be cached if the go.mod and go.sum files are not changed
RUN go mod download

# Copy the source code into the container
COPY . .

# Build the Go app
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o reminder-api ./cmd/server

# Start a new stage from scratch
FROM alpine:latest

# Install ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates

# Create a non-root user
RUN adduser -D -s /bin/sh appuser

WORKDIR /home/appuser/

# Copy the Pre-built binary file from the previous stage
COPY --from=build /app/reminder-api .
COPY --from=build /app/web ./web

# Change ownership to appuser
RUN chown -R appuser:appuser /home/appuser/

# Switch to non-root user
USER appuser

# Expose port 3000 to the outside world
EXPOSE 3000

# Command to run the executable
CMD ["./reminder-api"]

