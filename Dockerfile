# Build Stage
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Copy dependency files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code and prompts
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o cinesearch-pro .

# Final Stage
FROM alpine:latest

# Add CA certificates for HTTPS requests
RUN apk --no-cache add ca-certificates

WORKDIR /app

# Copy the binary and prompts from the builder
COPY --from=builder /app/cinesearch-pro .
COPY --from=builder /app/prompts ./prompts

# Expose the API port
EXPOSE 8080

# Run the application
CMD ["./cinesearch-pro"]
