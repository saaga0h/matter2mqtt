# matter2mqtt

**Bridge Matter/Thread devices to MQTT**, inspired by the Zigbee2MQTT architecture.

Break free from vendor lock-in. Control your Matter devices through open, transparent MQTT messaging.

> **Status: Work in Progress**
>
> This project is under active development and is not yet production-ready. Breaking changes and incomplete features are expected. Contributions and feedback are welcome!

## Why matter2mqtt?

**The Problem:** Most Matter border routers lock you into vendor ecosystems (Apple Home, Google Home, SmartThings, etc.). You can't see what's happening, can't easily integrate with other systems, and are dependent on vendor apps and clouds.

**The Solution:** matter2mqtt gives you **open architecture** with **transparent messaging**.

### Key Benefits

🔓 **No Vendor Lock-In**
- Not tied to Apple, Google, Amazon, or any vendor ecosystem
- Works with any MQTT-compatible platform (Home Assistant, Node-RED, custom scripts)
- Your devices, your data, your control

🔍 **Full Transparency**
- See exactly what's happening: `mosquitto_sub -v -t "matter/#"`
- Plain JSON messages you can read and understand
- Debug and inspect all device communication
- Unlike closed ecosystems where you can't see anything

🔒 **Network Isolation**
- Thread network stays isolated from your LAN (security)
- No IPv6 routing complexity or firewall headaches
- Matter devices can't phone home to vendors
- All communication flows through your controlled MQTT broker

📡 **MQTT is Better for Messaging**
- Open standard protocol (not proprietary)
- Universal integration (works with everything)
- Real-time, lightweight, reliable
- Easy to log, monitor, and automate
- Superior to closed vendor messaging systems

**See:** [Architecture Philosophy](docs/ARCHITECTURE_PHILOSOPHY.md) for complete design rationale.

## Features

- **MQTT-first architecture** - Not controller-centric, message-oriented
- **Clean separation** - Devices → MQTT → Your services (any consumer)
- **Universal compatibility** - Works with any MQTT subscriber
- **Network isolation** - Thread stays separate from your LAN (security + simplicity)
- **No cloud dependencies** - Fully local, no vendor accounts required
- **Transparent operation** - See all messages, understand all behavior
- **Open source** - Auditable, modifiable, community-driven

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

### Architecture & Philosophy
- **[docs/ARCHITECTURE_PHILOSOPHY.md](docs/ARCHITECTURE_PHILOSOPHY.md)** - Why matter2mqtt exists and design principles
- **[docs/NETWORK_ISOLATION.md](docs/NETWORK_ISOLATION.md)** - Understanding IPv6 and Thread network isolation

### Setup Guides
- **[DOCKER.md](DOCKER.md)** - Docker Compose setup with OTBR (Linux)
- **[DOCKER_MACOS.md](DOCKER_MACOS.md)** - macOS-specific Docker instructions and limitations
- **[chip-tool.md](chip-tool.md)** - How to get or build chip-tool binary

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
