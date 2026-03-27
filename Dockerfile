FROM golang:1.26-alpine@sha256:2389ebfa5b7f43eeafbd6be0c3700cc46690ef842ad962f6c5bd6be49ed82039 AS builder

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
FROM alpine:3.23@sha256:51183f2cfa6320055da30872f211093f9ff1d3cf06f39a0bdb212314c5dc7375

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /root/

# Copy binary from builder
COPY --from=builder /bin/e2c /bin/e2c

# Copy example config
COPY examples/config.yaml /etc/e2c/config.yaml

# Set up environment
ENV E2C_CONFIG_FILE=/etc/e2c/config.yaml

ENTRYPOINT ["/bin/e2c"]
