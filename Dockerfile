# Build stage
FROM golang:1.24 AS builder
WORKDIR /app

# Add build metadata
LABEL maintainer="Gayan Sanjeewa <iamgayan@gmail.com>"
LABEL description="Go get url!"
LABEL version="1.0"

# Download dependencies first for better layer caching
COPY go.mod go.sum ./
RUN go mod download

# Build the application
COPY . ./
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o gogeturl ./cmd/gogeturl

# Final stage
FROM alpine:3.19
ARG APP_USER=appuser
ARG APP_DIR=/app
ARG PORT=8080

# Create non-root user
RUN adduser -D -h ${APP_DIR} ${APP_USER}

WORKDIR ${APP_DIR}

# Copy only necessary files
COPY --from=builder /app/gogeturl .
COPY --chown=${APP_USER}:${APP_USER} cmd/templates ./cmd/templates
COPY --chown=${APP_USER}:${APP_USER} static ./static

# Set environment variables
ENV PORT=${PORT}

# Switch to non-root user
USER ${APP_USER}

# Add healthcheck
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --quiet --tries=1 --spider http://localhost:${PORT}/health || exit 1

# Expose port
EXPOSE ${PORT}

# Run the application
CMD ["./gogeturl"]
