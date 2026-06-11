FROM golang:1.26-alpine@sha256:d4c4845f5d60c6a974c6000ce58ae079328d03ab7f721a0734277e69905473e5 AS builder

WORKDIR /app

# Copy go.mod and go.sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 go build -ldflags="-s -w -X github.com/nlamirault/e2c/internal/version.Version=$(git describe --tags --always --dirty 2>/dev/null || echo dev)" -o /bin/e2c ./cmd/e2c

# Create final image
FROM alpine:3.24@sha256:a2d49ea686c2adfe3c992e47dc3b5e7fa6e6b5055609400dc2acaeb241c829f4

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /root/

# Copy binary from builder
COPY --from=builder /bin/e2c /bin/e2c

# Copy example config
COPY examples/config.yaml /etc/e2c/config.yaml

# Set up environment
ENV E2C_CONFIG_FILE=/etc/e2c/config.yaml

ENTRYPOINT ["/bin/e2c"]
