# Build stage
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o main .

# Final stage
FROM alpine:latest

WORKDIR /app

# Copy the binary and config
COPY --from=builder /app/main .
COPY --from=builder /app/config ./config
COPY --from=builder /app/static ./static

# Install ca-certificates for secure Redis connections
RUN apk --no-cache add ca-certificates

# Expose the port
EXPOSE 8080

# Run the application
CMD ["./main"] 