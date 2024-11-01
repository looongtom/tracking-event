# Use the official Golang image to create a build artifact.
FROM golang:1.20-alpine AS builder

RUN apk add --no-cache gcc musl-dev libc6-compat librdkafka-dev

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

ENV CGO_ENABLED=1 \
    GOOS=linux \
    GOARCH=amd64

RUN go build -o app -tags musl .

COPY .env /app/.env

# Stage 2: Final stage
FROM alpine:latest

# Install any required runtime dependencies
RUN apk --no-cache add ca-certificates libc6-compat librdkafka gcc

# Copy the built binary from the builder stage
COPY --from=builder /app/app /app/app
COPY --from=builder /app/.env /app/.env

# Command to run the executable
CMD ["/app/app"]
