# Docker Setup with Network-Attached Thread Radio (TCP)

> **Status:** ⚠️ **UNTESTED** - This guide is based on community feedback and research but has not been verified with actual hardware. If you test this setup, please share your results!

This guide explains how to use **matter2mqtt** with a **network-attached Thread radio** (like SLZB-MR1) that connects via TCP instead of USB.

## Table of Contents

1. [Overview](#overview)
2. [Prerequisites](#prerequisites)
3. [Architecture](#architecture)
4. [Setup Instructions](#setup-instructions)
5. [Configuration](#configuration)
6. [Troubleshooting](#troubleshooting)

---

## Overview

### Why Network-Attached Thread Radios?

Network-attached Thread radios (like SLZB-MR1, SLZB-06, etc.) connect to your network via Ethernet/WiFi and expose the Thread radio over TCP. This offers several advantages:

- **No USB passthrough needed** - Great for Docker/VM environments
- **Physical separation** - Place the radio in an optimal location
- **Works on macOS** - Avoids Docker USB passthrough limitations
- **Reliability** - Ethernet is more stable than USB over long cables

### How This Works

The key insight: **matter2mqtt doesn't need direct access to the Thread radio!**

```
┌──────────────────────────────────────────────────────────────┐
│ Your Network (TCP/IP)                                         │
├──────────────────────────────────────────────────────────────┤
│                                                                │
│  SLZB-MR1 (192.168.1.50:6638)                                │
│      ↓ TCP connection                                         │
│  OTBR Container (bnutzer/otbr-tcp)                           │
│      ↓ Creates Thread Network (isolated IPv6 network)        │
│                                                                │
│  Thread Network (802.15.4 radio, IPv6)                       │
│      ↓                                                         │
│  Matter Devices (IPv6 link-local addresses)                  │
│      ↓                                                         │
│  matter2mqtt (chip-tool via IPv6 network)                    │
│      ↓ MQTT over TCP                                          │
│  Mosquitto MQTT Broker                                        │
│                                                                │
└──────────────────────────────────────────────────────────────┘
```

**Critical Understanding:**
- OTBR handles the radio connection (via TCP to SLZB-MR1)
- OTBR creates a Thread network with IPv6 addresses
- matter2mqtt communicates with Matter devices over IPv6
- **matter2mqtt never touches the radio hardware directly**

---

## Prerequisites

### Required Hardware

- **Network-attached Thread radio device**, such as:
  - [SMLIGHT SLZB-MR1](https://smlight.tech/product/slzb-mr1/) (Zigbee + Thread + Matter)
  - [SMLIGHT SLZB-06](https://smlight.tech/product/slzb-06/) series
  - Any Thread RCP (Radio Co-Processor) with TCP/IP connectivity

### Required Software

- Docker Engine 20.10+
- Docker Compose 2.0+
- IPv6 enabled on host (see [DOCKER.md IPv6 Requirements](DOCKER.md#ipv6-requirements))

### Network Requirements

- Thread radio device accessible on your network
- Know the IP address and port (typically port 6638 for SMLIGHT devices)

---

## Architecture

### Component Separation

Unlike USB-based setups, the OTBR and matter2mqtt services are more clearly separated:

**OTBR (Border Router):**
- Connects to Thread radio via TCP (using socat internally)
- Manages Thread network formation
- Bridges Thread (IPv6) to your infrastructure IPv6
- Provides network interface for Thread devices

**matter2mqtt (Matter Controller):**
- Uses chip-tool to commission and control Matter devices
- Communicates over IPv6 network (created by OTBR)
- Converts Matter messages to MQTT
- **No direct radio access needed**

---

## Setup Instructions

### Step 1: Find Your Thread Radio IP and Port

**For SLZB-MR1/SLZB-06:**
```bash
# Access the device web interface
# Default: http://slzb-mr1.local or check your router's DHCP table

# Common default port: 6638 (Thread RCP)
```

**Verify TCP connectivity:**
```bash
# Test if the port is open
nc -zv 192.168.1.50 6638

# Or use telnet
telnet 192.168.1.50 6638
```

### Step 2: Set Up OTBR with TCP Support

You have two options:

#### Option A: Use Existing OTBR Container (Recommended)

If you already have OTBR running (e.g., from Home Assistant or standalone), you can **reuse it**. Just ensure:
- It's configured to connect to your Thread radio via TCP
- It's creating a Thread network
- Skip the OTBR service in docker-compose.yml below

#### Option B: Add OTBR to docker-compose.yml

Use the [bnutzer/docker-otbr-tcp](https://github.com/bnutzer/docker-otbr-tcp) image:

```yaml
services:
  otbr:
    image: bnutzer/otbr-tcp:latest
    container_name: otbr-tcp
    network_mode: host
    privileged: true
    environment:
      # Thread Radio Connection (REQUIRED)
      - RCP_HOST=192.168.1.50          # Your Thread radio IP
      - RCP_PORT=6638                   # Thread radio port (default 6638)
      - RCP_USE_TCP=1                   # Enable TCP mode

      # Thread Network Configuration
      - NETWORK_NAME=matter2mqtt-thread
      - CHANNEL=15
      - PANID=0x1234
      - EXTPANID=1111111122222222
      - NETWORKKEY=00112233445566778899aabbccddeeff

      # OTBR Settings
      - INFRA_IF_NAME=eth0              # Your network interface
      - BACKBONE_ROUTER=1
      - NAT64=0
      - DNS64=0
    volumes:
      - otbr_data:/data
    restart: unless-stopped
```

### Step 3: Configure matter2mqtt

Create `docker-compose.yml`:

```yaml
version: '3.8'

services:
  # MQTT Broker
  mosquitto:
    image: eclipse-mosquitto:2
    container_name: mosquitto
    ports:
      - "1883:1883"
      - "9001:9001"
    volumes:
      - ./mosquitto.conf:/mosquitto/config/mosquitto.conf:ro
      - mosquitto_data:/mosquitto/data
      - mosquitto_logs:/mosquitto/log
    restart: unless-stopped

  # OTBR with TCP support (if not using existing OTBR)
  otbr:
    image: bnutzer/otbr-tcp:latest
    container_name: otbr-tcp
    network_mode: host
    privileged: true
    environment:
      # IMPORTANT: Change these to match your Thread radio
      - RCP_HOST=192.168.1.50          # Your Thread radio IP
      - RCP_PORT=6638
      - RCP_USE_TCP=1
      - NETWORK_NAME=matter2mqtt-thread
      - CHANNEL=15
      - PANID=0x1234
      - EXTPANID=1111111122222222
      - NETWORKKEY=00112233445566778899aabbccddeeff
      - INFRA_IF_NAME=eth0
      - BACKBONE_ROUTER=1
      - NAT64=0
      - DNS64=0
    volumes:
      - otbr_data:/data
    restart: unless-stopped
    depends_on:
      - mosquitto

  # matter2mqtt Bridge
  matter2mqtt:
    build:
      context: .
      dockerfile: Dockerfile
    container_name: matter2mqtt
    # IMPORTANT: host networking required for IPv6 and mDNS
    network_mode: host
    volumes:
      - ./config.yaml:/app/config/config.yaml:ro
      - ./devices.yaml:/app/config/devices.yaml:ro
      - matter_data:/var/lib/matter2mqtt
    # NO USB DEVICES NEEDED!
    # matter2mqtt talks to devices over IPv6 network, not via radio hardware
    environment:
      - MATTER_STORAGE_PATH=/var/lib/matter2mqtt
      # Uncomment for mock mode (testing)
      # - MOCK_CHIPTOOL=true
    restart: unless-stopped
    depends_on:
      - mosquitto
      - otbr

volumes:
  mosquitto_data:
  mosquitto_logs:
  otbr_data:
  matter_data:
```

### Step 4: Create Configuration Files

**mosquitto.conf:**
```conf
listener 1883
allow_anonymous true
persistence true
persistence_location /mosquitto/data/
log_dest file /mosquitto/log/mosquitto.log
log_dest stdout
```

**config.yaml:**
```yaml
mqtt:
  server: localhost
  port: 1883
  user: ""
  password: ""
  base_topic: matter

matter:
  storage_path: /var/lib/matter2mqtt

bridge:
  log_level: info
  port: 8080
```

**devices.yaml:**
```yaml
# Configure after commissioning devices
devices: []
```

### Step 5: Start Services

```bash
# Verify IPv6 is enabled
./scripts/verify-ipv6.sh

# Build and start
docker compose build
docker compose up -d

# Check logs
docker compose logs -f

# Specifically check OTBR connection
docker compose logs otbr | grep -i "RCP\|connection\|radio"
```

---

## Configuration

### Environment Variables for OTBR (bnutzer/otbr-tcp)

**Required:**
- `RCP_HOST` - IP address of your Thread radio device
- `RCP_USE_TCP=1` - Enable TCP mode

**Optional:**
- `RCP_PORT=6638` - Thread radio port (default: 6638)
- `RCP_TTY=/tmp/ttyOTBR` - Internal socket path
- `SOCAT_SOURCE_PARAMETERS` - socat local socket parameters
- `SOCAT_DESTINATION_PARAMETERS` - socat remote connection parameters

**Thread Network:**
- `NETWORK_NAME` - Thread network name
- `CHANNEL` - Thread channel (11-26)
- `PANID` - PAN ID (0x0000-0xFFFF)
- `EXTPANID` - Extended PAN ID (16 hex digits)
- `NETWORKKEY` - Network key (32 hex digits) - **CHANGE THIS!**

### Network Interface Selection

The `INFRA_IF_NAME` variable tells OTBR which network interface to use:

```bash
# Find your network interface
ip link show

# Common values:
# - eth0 (wired Ethernet)
# - wlan0 (WiFi)
# - ens33 (VM interface)
# - enp0s3 (some systems)
```

---

## Troubleshooting

### Issue: OTBR can't connect to Thread radio

**Check network connectivity:**
```bash
# Ping the device
ping 192.168.1.50

# Test TCP port
nc -zv 192.168.1.50 6638
telnet 192.168.1.50 6638
```

**Check OTBR logs:**
```bash
docker compose logs otbr | grep -i error
docker compose logs otbr | grep -i "RCP"
```

**Common issues:**
- Wrong IP address in `RCP_HOST`
- Firewall blocking port 6638
- Thread radio not configured for TCP mode
- Network connectivity issues

### Issue: Thread radio device configuration

**For SLZB-MR1:**
1. Access web interface: `http://slzb-mr1.local`
2. Navigate to Thread settings
3. Ensure Thread mode is enabled
4. Note the TCP port (default 6638)

### Issue: matter2mqtt can't discover devices

**Verify Thread network is up:**
```bash
# Check OTBR status
docker exec -it otbr-tcp ot-ctl state
# Should show: leader or router

# Check Thread network
docker exec -it otbr-tcp ot-ctl dataset active
```

**Verify IPv6 connectivity:**
```bash
# From host
ip -6 addr show

# Should see Thread network prefix (fd prefixes)
ip -6 route show
```

**Check matter2mqtt can reach devices:**
```bash
# Test IPv6 ping to a known Matter device
ping6 -c 3 fe80::xxxx:xxxx:xxxx:xxxx%wpan0
```

### Issue: Services can't communicate

**All services should use host networking:**
- OTBR needs host networking for border router functionality
- matter2mqtt needs host networking for IPv6/mDNS
- Mosquitto can use bridge networking (exposes port 1883)

**Check IPv6 is enabled:**
```bash
cat /proc/sys/net/ipv6/conf/all/disable_ipv6
# Should output: 0 (enabled)

cat /proc/sys/net/ipv6/conf/all/forwarding
# Should output: 1 (enabled for OTBR)
```

### Issue: No MQTT messages

**Debug MQTT:**
```bash
# Subscribe to all matter topics
mosquitto_sub -h localhost -p 1883 -t "matter/#" -v

# Test MQTT broker
mosquitto_pub -h localhost -p 1883 -t "matter/test" -m "hello"
```

**Check matter2mqtt logs:**
```bash
docker compose logs -f matter2mqtt
```

### Verify socat Connection (Advanced)

If using the bnutzer/otbr-tcp image:

```bash
# Check socat process is running
docker exec -it otbr-tcp ps aux | grep socat

# Should see something like:
# socat pty,link=/tmp/ttyOTBR,raw,echo=0 tcp:192.168.1.50:6638
```

---

## Advanced Configuration

### Custom socat Parameters

The bnutzer/otbr-tcp image allows fine-tuning socat:

```yaml
environment:
  # Local socket parameters
  - SOCAT_SOURCE_PARAMETERS=raw,echo=0,wait-slave,ignoreeof

  # Remote TCP parameters
  # These enable persistent reconnection
  - SOCAT_DESTINATION_PARAMETERS=nodelay,keepalive,forever,interval=5
```

### Using Multiple Thread Networks

You can run multiple OTBR instances with different Thread radios:

```yaml
services:
  otbr-network1:
    image: bnutzer/otbr-tcp:latest
    environment:
      - RCP_HOST=192.168.1.50
      - RCP_PORT=6638
      - NETWORK_NAME=home-thread
      # ...

  otbr-network2:
    image: bnutzer/otbr-tcp:latest
    environment:
      - RCP_HOST=192.168.1.51
      - RCP_PORT=6638
      - NETWORK_NAME=garage-thread
      # ...
```

However, matter2mqtt would need to be configured to handle multiple Thread networks (currently not supported).

---

## Testing This Setup

If you test this guide with actual hardware, please consider:

1. **Document your hardware:**
   - Thread radio model and firmware version
   - Host OS and Docker version
   - Network configuration

2. **Share results:**
   - Open an issue on GitHub with your findings
   - Include logs (sanitize sensitive info)
   - Note any deviations from this guide

3. **Contribute improvements:**
   - Submit PRs with corrections
   - Add device-specific notes
   - Share configuration examples

---

## Comparison: TCP vs USB

| Aspect | USB Setup | TCP Setup |
|--------|-----------|-----------|
| **Hardware** | USB dongle directly attached | Network-attached radio |
| **Docker** | Needs USB passthrough | No USB passthrough needed |
| **macOS** | Limited/broken USB support | Works well |
| **Placement** | Must be near host | Can be anywhere on network |
| **Reliability** | USB connection issues | Network dependency |
| **Complexity** | Simpler (direct connection) | More components (OTBR + socat) |
| **Cost** | USB dongle (~$10-30) | Network radio (~$30-80) |

---

## Related Documentation

- [DOCKER.md](DOCKER.md) - USB-based Docker setup
- [DOCKER_MACOS.md](DOCKER_MACOS.md) - macOS-specific instructions
- [README.md](README.md) - Project overview

## External Resources

- [bnutzer/docker-otbr-tcp](https://github.com/bnutzer/docker-otbr-tcp) - OTBR Docker image with TCP support
- [SMLIGHT SLZB-MR1](https://smlight.tech/product/slzb-mr1/) - Multi-protocol coordinator
- [OpenThread Border Router Guide](https://openthread.io/guides/border-router)
- [chip-tool Guide](https://project-chip.github.io/connectedhomeip-doc/development_controllers/chip-tool/chip_tool_guide.html)

---

## Community Contributions

This guide was created based on community feedback. Special thanks to users who reported their TCP-based setups and helped identify this use case.

**Status tracking:**
- [ ] Tested with SLZB-MR1
- [ ] Tested with SLZB-06
- [ ] Tested with other TCP-based Thread radios
- [ ] Verified Matter device commissioning works
- [ ] Verified MQTT message flow
- [ ] Tested on Linux
- [ ] Tested on macOS
- [ ] Tested on Windows (WSL2)

If you successfully test any of these scenarios, please update this list via PR or issue!
