# Build stage
FROM golang:1.23-bookworm AS builder

WORKDIR /build

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o matter2mqtt ./cmd/main.go

# Runtime stage
FROM debian:bookworm-slim

# Install dependencies for chip-tool
RUN apt-get update && apt-get install -y \
    ca-certificates \
    libssl3 \
    libglib2.0-0 \
    libavahi-client3 \
    libavahi-common3 \
    wget \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app

# Copy the built binary
COPY --from=builder /build/matter2mqtt /app/matter2mqtt

# Create directories for Matter storage and config
RUN mkdir -p /var/lib/matter2mqtt /app/config

# Download chip-tool (ARM64 or AMD64)
# Note: You may need to build chip-tool yourself for your architecture
# See chip-tool.md for instructions
# For now, we'll document where to place it
RUN mkdir -p /usr/local/bin

# Add a volume for chip-tool if you want to mount it from host
VOLUME ["/usr/local/bin/chip-tool"]

# Expose port for bridge API (if implemented)
EXPOSE 8080

# Set environment variables
ENV MATTER_STORAGE_PATH=/var/lib/matter2mqtt

# Run the application
CMD ["/app/matter2mqtt", "-config", "/app/config/config.yaml", "-devices", "/app/config/devices.yaml"]
