# ------------ 1) Build stage ------------
FROM golang:1.25-alpine AS builder
WORKDIR /src

# Make toolchain auto-fetch patch versions; build static
ENV CGO_ENABLED=0 GOTOOLCHAIN=auto

# Needed for some modules + TLS during "go mod download"
RUN apk add --no-cache ca-certificates git

# Cache deps first
COPY go.mod go.sum ./
RUN go mod download

# Copy source (includes embedded web assets)
COPY . .

# Build the gateway binary (cmd/gateway)
RUN go build -trimpath -ldflags='-s -w' -o /out/nvimanywhere ./cmd/gateway

# ------------ 2) Runtime stage ------------
# Minimal, non-root, includes CA certs
FROM gcr.io/distroless/static:nonroot
WORKDIR /app

# Copy binary only (no config baked in)
COPY --from=builder /out/nvimanywhere /app/nvimanywhere

# Default config path inside the container (must be mounted at run)
ENV NVA_CONFIG=/etc/nva/config.yaml

EXPOSE 8080
USER nonroot:nonroot
ENTRYPOINT ["/app/nvimanywhere"]

