# Use the official Golang image to create a build artifact.
FROM golang:1.23-alpine AS builder


# Set the Current Working Directory inside the container
WORKDIR /app

# Copy go mod and sum files
COPY go.mod .

# Download all dependencies. Dependencies will be cached if the go.mod and go.sum files are not changed
RUN go mod download

# Copy the source from the current directory to the Working Directory inside the container
COPY . .

ENV CGO_ENABLED=0 GOOS=linux GOARCH=amd64
RUN go build -ldflags="-s -w" -o apiserver .


# Copy the .env file
COPY .env /app/.env

# Expose port 8080 to the outside world
EXPOSE 8080

# Command to run the executable
CMD ["./apiserver"]