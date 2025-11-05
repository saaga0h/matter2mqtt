# matter2mqtt

Bridge Matter/Thread devices to MQTT, inspired by the Zigbee2MQTT architecture.

> **Status: Work in Progress**
>
> This project is under active development and is not yet production-ready. Breaking changes and incomplete features are expected. Contributions and feedback are welcome!

## Features
- MQTT-first architecture (not controller-centric)
- Clean separation: devices → MQTT → your services
- Works with any MQTT subscriber
- No vendor lock-in

## Requirements
- Thread Border Router (OTBR + compatible radio)
- MQTT broker
- Matter/Thread devices

## Quick Start

### Option 1: Docker Compose (Recommended)

```bash
# Copy example configuration
cp config.yaml.example config.yaml
cp devices.yaml.example devices.yaml

# Edit configuration with your settings
nano config.yaml

# Start services
docker compose up -d
```

**See:** [DOCKER.md](DOCKER.md) for complete Docker setup instructions including USB device configuration.

**macOS users:** See [DOCKER_MACOS.md](DOCKER_MACOS.md) for macOS-specific instructions.

### Option 2: Binary Release

Download the [latest release](releases) and extract it, or build from source (see below).

## Documentation

- **[DOCKER.md](DOCKER.md)** - Docker Compose setup with OTBR
- **[DOCKER_MACOS.md](DOCKER_MACOS.md)** - macOS-specific Docker instructions
- **[docs/NETWORK_ISOLATION.md](docs/NETWORK_ISOLATION.md)** - Understanding IPv6 and network isolation
- **[chip-tool.md](chip-tool.md)** - How to get chip-tool binary

## Configuration

Configuration is managed via `config.yaml` and `devices.yaml`.

**Example config.yaml:**
```yaml
mqtt:
  server: localhost
  port: 1883
  base_topic: matter

matter:
  storage_path: /var/lib/matter2mqtt
```

**Example devices.yaml:**
```yaml
devices:
  12345:  # Matter node ID
    topic: living-room/light
    sensitivity: high
```

See `config.yaml.example` and `devices.yaml.example` for complete templates.

## Building from Source

```sh
# Clone the repository
git clone https://github.com/yourusername/matter2mqtt.git
cd matter2mqtt

# Initialize Go modules
go mod tidy

# Build the bridge
make build
```

## Contributing

Contributions, bug reports, and feature requests are welcome! Please open an issue or pull request on GitHub.

## License

This project is licensed under the terms of the MIT License. See the [LICENSE](LICENSE) file for details.
