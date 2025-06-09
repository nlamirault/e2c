FROM golang:1.24-alpine@sha256:68932fa6d4d4059845c8f40ad7e654e626f3ebd3706eef7846f319293ab5cb7a AS builder

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
FROM alpine:3.19@sha256:e5d0aea7f7d2954678a9a6269ca2d06e06591881161961ea59e974dff3f12377

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /root/

# Copy binary from builder
COPY --from=builder /bin/e2c /bin/e2c

# Copy example config
COPY examples/config.yaml /etc/e2c/config.yaml

# Set up environment
ENV E2C_CONFIG_FILE=/etc/e2c/config.yaml

ENTRYPOINT ["/bin/e2c"]
