
FROM golang:1.24-alpine AS builder
LABEL authors="Angelo Reyes"

WORKDIR /app

#COPY go.mod go.sum ./ # uncomment when we have non-standard deps
COPY go.mod ./

RUN #go mod download # uncomment when we have non-standard deps

COPY cmd/server/main.go ./cmd/server/main.go

# Build the Go app - Creates a static binary
# CGO_ENABLED=0 prevents linking against C libraries
# -ldflags="-w -s" strips debugging information, reducing binary size
# -o /app/server creates the binary at /app/server
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /app/server ./cmd/server/main.go

# Stage 2: Create the final, minimal runtime image
FROM alpine:3.21.3

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy only the executable binary from the builder stage
COPY --from=builder /app/server /app/server

# Expose port 8080 to the outside world
EXPOSE 8080

# Command to run the executable
# Use the full path to the executable
ENTRYPOINT ["/app/server"]