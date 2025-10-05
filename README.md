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

Download the [latest release](releases) and extract it, or build from source (see below).

## Configuration

Configuration is managed via `config.yaml` and `devices.yaml`.

> **TODO:** Add detailed configuration instructions.

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
