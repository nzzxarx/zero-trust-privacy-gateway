# Stage 1: Build the Go binary
FROM golang:1.26-alpine AS builder

WORKDIR /app
COPY go.mod ./
# COPY go.sum ./ (Uncomment if you have external dependencies)

COPY *.go ./
# Compile with optimizations (-s -w removes debug symbols)
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o scrubber .

# Stage 2: Create a minimal, secure production image
FROM alpine:3.20

# Create a non-root user for security (Zero-Trust principle)
RUN addgroup -S appgroup && adduser -S appuser -G appgroup

WORKDIR /app

# Copy the compiled binary from the builder stage
COPY --from=builder /app/scrubber .

# Enforce non-root execution
USER appuser

# Expose port
EXPOSE 8080

# K8s/OpenShift Healthcheck
HEALTHCHECK --interval=30s --timeout=3s \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8080/healthz || exit 1

ENTRYPOINT ["./scrubber"]