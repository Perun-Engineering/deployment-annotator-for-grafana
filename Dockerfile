# Build stage
FROM --platform=$BUILDPLATFORM golang:1.24-alpine AS builder

# Build arguments for multi-arch support
ARG TARGETOS
ARG TARGETARCH
ARG TARGETVARIANT
ARG BUILDTIME
ARG VERSION
ARG REVISION

WORKDIR /app

# Install build dependencies for static compilation
RUN apk add --no-cache git ca-certificates tzdata

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download && go mod verify

# Copy source code
COPY main.go ./

# Ensure go.mod is up to date with dependencies
RUN go mod tidy

# Build the static binary
RUN CGO_ENABLED=0 \
    GOOS=${TARGETOS} \
    GOARCH=${TARGETARCH} \
    GOARM=${TARGETVARIANT#v} \
    go build \
    -a \
    -installsuffix cgo \
    -ldflags="-w -s -extldflags '-static' -X main.version=${VERSION} -X main.commit=${REVISION} -X main.buildTime=${BUILDTIME}" \
    -tags netgo \
    -o controller .

# Final stage - minimal scratch image
FROM scratch

# Copy CA certificates for HTTPS requests
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Copy timezone data
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo

# Copy the static binary
COPY --from=builder /app/controller /controller

# Create passwd file for non-root user
COPY --from=builder /etc/passwd /etc/passwd

# Use non-root user (nobody)
USER 65534:65534

EXPOSE 8080 8081

# Add labels for better metadata
LABEL org.opencontainers.image.title="Grafana Annotation Controller" \
    org.opencontainers.image.description="Kubernetes Controller for creating Grafana deployment annotations" \
    org.opencontainers.image.vendor="Platform Team" \
    org.opencontainers.image.licenses="MIT" \
    org.opencontainers.image.source="https://github.com/perun-engineering/deployment-annotator-for-grafana" \
    org.opencontainers.image.documentation="https://github.com/perun-engineering/deployment-annotator-for-grafana/blob/main/README.md"

ENTRYPOINT ["/controller"]
