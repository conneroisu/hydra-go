FROM golang:1.21-alpine AS builder

WORKDIR /app
COPY main.go .

# Build the mock server
RUN go build -o mock-hydra main.go

FROM alpine:latest

# Install ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy the binary from builder stage
COPY --from=builder /app/mock-hydra .

EXPOSE 3000 8080

CMD ["./mock-hydra"]