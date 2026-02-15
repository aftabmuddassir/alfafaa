# ============================================================
# STAGE 1: Build the Go binary
# ============================================================
# We use a full Go image to compile. "alpine" variant is smaller
# than the default Debian-based image (~300MB vs ~800MB).
FROM golang:1.24-alpine AS builder

# Install git and ca-certificates:
#   - git: needed for `go mod download` to fetch private/VCS deps
#   - ca-certificates: needed so the final binary can make HTTPS
#     calls (e.g., to Neon's PostgreSQL over SSL)
RUN apk add --no-cache git ca-certificates

# Set the working directory inside the container.
# All subsequent commands (COPY, RUN) happen relative to this.
WORKDIR /app

# Copy go.mod and go.sum FIRST, then download dependencies.
# Docker caches each layer — if these files haven't changed,
# Docker reuses the cached layer and skips `go mod download`.
# This makes rebuilds MUCH faster when only your .go files change.
COPY go.mod go.sum ./
RUN go mod download

# Now copy the entire source code.
# This layer is only rebuilt when your code changes.
COPY . .

# Build the Go binary:
#   CGO_ENABLED=0  — produce a statically linked binary (no C libs needed)
#   GOOS=linux     — target Linux (the container OS)
#   -o /app/server — output binary path
#   -ldflags="-s -w" — strip debug info to reduce binary size (~30% smaller)
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /app/server ./cmd/server/main.go

# ============================================================
# STAGE 2: Create the final minimal image
# ============================================================
# We use `alpine` (~5MB) instead of `scratch` because:
#   - We get a shell for debugging (`docker exec -it ... sh`)
#   - We get ca-certificates for TLS/SSL connections to Neon
#   - We get timezone data
FROM alpine:3.19

# Install CA certificates so the app can connect to Neon over SSL.
# Also install tzdata for timezone support.
RUN apk add --no-cache ca-certificates tzdata

# Create a non-root user for security.
# Running as root inside a container is a security risk — if
# an attacker breaks out, they'd have root on the host.
RUN addgroup -S appgroup && adduser -S appuser -G appgroup

# Set working directory
WORKDIR /app

# Copy ONLY the compiled binary from the builder stage.
# The final image won't contain Go, source code, or build tools.
COPY --from=builder /app/server .

# Create the uploads directory and give our non-root user ownership.
# This is where uploaded images will be stored.
RUN mkdir -p /app/uploads && chown -R appuser:appgroup /app

# Switch to the non-root user
USER appuser

# Document which port the container listens on.
# This doesn't actually publish the port — it's metadata for
# docker-compose and anyone reading the Dockerfile.
EXPOSE 8081

# The command to run when the container starts.
# This runs your compiled Go binary.
CMD ["./server"]
