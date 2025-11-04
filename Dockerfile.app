# =========================
# Stage 1: Build Stage
# =========================
FROM golang:1.25-alpine AS builder

# Set Go environment
ENV GO111MODULE=on \
    GOPROXY=https://proxy.golang.org

# Install build dependencies
RUN apk add --no-cache git bash make

# Set working directory
WORKDIR /app

# Copy Go modules to leverage Docker cache
COPY go.mod go.sum ./

# Download dependencies
RUN go mod tidy

# Copy application code
COPY . .

# Build statically-linked, optimized Go binary
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o statio-backend ./cmd/app

# =========================
# Stage 2: Final Stage (Minimal)
# =========================
FROM alpine:3.22

# Install SSL certificates
RUN apk add --no-cache ca-certificates

# Add non-root user
RUN adduser -D -g '' statio

# Set working directory
WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/statio-backend .

# Switch to non-root user
USER statio

# Run the app
CMD ["./statio-backend"]
