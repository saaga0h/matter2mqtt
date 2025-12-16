# matter2mqtt

## Zigbee2MQTT, but for Matter

**If you use [Zigbee2MQTT](https://www.zigbee2mqtt.io/), you already know why this exists.**

Just like Zigbee2MQTT freed Zigbee devices from proprietary hubs (Philips Hue Bridge, IKEA Gateway, etc.), **matter2mqtt frees Matter devices from vendor ecosystems** (Apple Home, Google Home, SmartThings).

```bash
# See everything happening in real-time (just like Zigbee2MQTT)
mosquitto_sub -v -t "matter/#"

matter/bedroom/motion {"presence": true, "timestamp": "..."}
matter/living-room/temp {"temperature": 21.5, "humidity": 45}
matter/kitchen/light {"state": "on", "brightness": 80}
```

**Same philosophy:**
- 🔓 Break vendor lock-in → Use any MQTT client
- 🔍 Full transparency → See all messages
- 📡 Open standard → MQTT, not proprietary protocols
- 🏠 Local control → No cloud dependencies
- 🛠️ Universal integration → Works with everything

> **Status: Work in Progress**
>
> This project is under active development and is not yet production-ready. Breaking changes and incomplete features are expected. Contributions and feedback are welcome!

## Why matter2mqtt?

**The Problem:** Most Matter border routers lock you into vendor ecosystems (Apple Home, Google Home, SmartThings, etc.). You can't see what's happening, can't easily integrate with other systems, and are dependent on vendor apps and clouds.

**The Solution:** Apply the proven Zigbee2MQTT model to Matter/Thread.

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
  - USB Thread dongle (nRF52840, EFR32, CC2652, etc.) **OR**
  - Network-attached Thread radio (SLZB-MR1, SLZB-06, etc. via TCP)
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

**Setup guides:**
- **USB Thread dongle:** [DOCKER.md](DOCKER.md) - Complete Docker setup with USB device configuration
- **Network-attached radio (TCP):** [DOCKER_TCP.md](DOCKER_TCP.md) - Setup for SLZB-MR1 and similar devices
- **macOS users:** [DOCKER_MACOS.md](DOCKER_MACOS.md) - macOS-specific instructions and limitations

### Option 2: Binary Release

Download the [latest release](releases) and extract it, or build from source (see below).

## Documentation

### Architecture & Philosophy
- **[docs/ARCHITECTURE_PHILOSOPHY.md](docs/ARCHITECTURE_PHILOSOPHY.md)** - Why matter2mqtt exists and design principles
- **[docs/NETWORK_ISOLATION.md](docs/NETWORK_ISOLATION.md)** - Understanding IPv6 and Thread network isolation

### Setup Guides
- **[DOCKER.md](DOCKER.md)** - Docker Compose setup with USB Thread dongle
- **[DOCKER_TCP.md](DOCKER_TCP.md)** - Docker setup with network-attached Thread radio (SLZB-MR1, etc.) ⚠️ Untested
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
