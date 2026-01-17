# ------------ 1) Build stage ------------
FROM golang:1.25-alpine AS builder
WORKDIR /src

# Deterministic static build
ENV CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64 \
    GOTOOLCHAIN=auto

# Needed for go modules + TLS
RUN apk add --no-cache ca-certificates git

# Cache dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy source
COPY . .

# Build binary
RUN go build \
    -trimpath \
    -ldflags="-s -w" \
    -o /out/nvimanywhere \
    ./cmd/gateway

# ------------ 2) Runtime stage ------------
FROM gcr.io/distroless/static:nonroot
WORKDIR /app

# Copy binary only
COPY --from=builder /out/nvimanywhere /app/nvimanywhere

# Default server config path (must be mounted)
ENV NVA_CONFIG=/etc/nva/config.yaml

EXPOSE 8080
USER nonroot:nonroot

ENTRYPOINT ["/app/nvimanywhere"]

