# Start from the official golang image
FROM golang:alpine AS build

# Set the current working directory inside the container
WORKDIR /app

# Copy go.mod and go.sum files to the working directory
COPY go.mod go.sum ./

# Download all dependencies. Dependencies will be cached if the go.mod and go.sum files are not changed
RUN go mod download

# Copy the source code into the container
COPY . .

# Build the Go app
RUN go build -o reminder-api .

# Start a new stage from scratch
FROM alpine:latest

# Copy the Pre-built binary file from the previous stage
COPY --from=build /app/reminder-api /reminder-api

# Expose port 8080 to the outside world
EXPOSE 8080

# Command to run the executable
CMD ["./reminder-api"]

