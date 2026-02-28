# Multi-stage Dockerfile for the OpenShift cluster health-checker.
#
# Stage 1: Builder
#   - Uses the official Go image to compile the binary.
#   - CGO_ENABLED=0 produces a fully static binary with no C dependencies.
#   - GOOS=linux ensures the binary targets Linux regardless of build host OS.
#   - Sets group-0 ownership and group-executable permissions on the binary so
#     it is executable by any UID assigned at runtime (arbitrary UID support).
#
# Stage 2: Final image
#   - Uses Google's distroless/static image (no shell, no package manager).
#   - No USER instruction with a fixed numeric UID — OpenShift's 'restricted' SCC
#     assigns an arbitrary UID from the namespace's allocated range at pod start.
#   - Binary is owned by root group (GID 0) with group-executable permissions,
#     following the Red Hat arbitrary UID pattern. Any non-root UID in group 0
#     can execute the binary.
#   - Minimal attack surface: only the compiled binary is present.

# ---- Stage 1: Build ----
FROM golang:1.21-alpine AS builder

WORKDIR /build

# Copy go module files first for better layer caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the static binary, then set group-0 ownership and group-executable
# permissions so the binary is executable by any UID in group 0 (arbitrary UID support).
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-s -w" \
    -o /health-checker \
    ./cmd/health-checker && \
    chown 0:0 /health-checker && \
    chmod g+x /health-checker

# ---- Stage 2: Final image ----
# distroless/static-debian12 (without :nonroot tag) does not set a fixed USER,
# allowing OpenShift to inject an arbitrary UID at runtime via the 'restricted' SCC.
# The binary is group-0 executable, so it runs correctly under any assigned UID.
FROM gcr.io/distroless/static-debian12

# Copy the compiled binary from the builder stage (retains group-0 ownership and permissions)
COPY --from=builder /health-checker /health-checker

# No USER instruction — OpenShift's 'restricted' SCC assigns an arbitrary non-root
# UID from the namespace's allocated UID range. runAsNonRoot: true in the Deployment
# securityContext ensures the assigned UID is always non-zero.

# Expose the metrics port (default 8080; configurable via METRICS_PORT env var)
EXPOSE 8080

ENTRYPOINT ["/health-checker"]
