# Use the official Golang image (1.22.2) for building the application
FROM golang:1.23.0-alpine AS builder

# Install build dependencies in the builder image
RUN apk add --no-cache git

# Set the working directory inside the container
WORKDIR /app

# Copy the Go module files and download dependencies
COPY go.mod ./
RUN go mod download

# Copy the rest of the application code
COPY . .

# Build the Go application
RUN go build -o ea-credentials-manager .

# Use a minimal Alpine image to run the built application
FROM alpine:latest

# Set the working directory in the runtime container
WORKDIR /app

# Copy the built application binary from the builder stage
COPY --from=builder /app/ea-credentials-manager .

# Expose port 8080
EXPOSE 8080

CMD ["./ea-credentials-manager"]
