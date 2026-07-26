# Build stage
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Copy source code and build
COPY . .
RUN go mod tidy
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o server .

# Final stage
FROM alpine:3.19

WORKDIR /app

# Copy the compiled binary from builder
COPY --from=builder /app/server .

# Expose the application port
EXPOSE 8080

# Run the server binary
CMD ["./server"]
