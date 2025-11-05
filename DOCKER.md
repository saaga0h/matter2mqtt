# Docker Compose Setup for matter2mqtt

This guide explains how to run **matter2mqtt** and **Open Thread Border Router (OTBR)** using Docker Compose, including how to expose USB Matter/Thread dongles to the containers.

> **Important Note:** Your LAN does **NOT** need IPv6 routing! Thread creates an isolated IPv6 network via radio. Only your host needs link-local IPv6 (enabled by default). See [IPv6 Requirements](#ipv6-requirements) for details.

> **macOS Users:** See [DOCKER_MACOS.md](DOCKER_MACOS.md) for macOS-specific setup (USB passthrough limitations apply).

## Table of Contents

1. [Prerequisites](#prerequisites)
2. [IPv6 Requirements](#ipv6-requirements)
3. [Quick Start](#quick-start)
4. [USB Device Configuration](#usb-device-configuration)
5. [Service Architecture](#service-architecture)
6. [Configuration](#configuration)
7. [Troubleshooting](#troubleshooting)

---

## Prerequisites

### Required Hardware
- **Thread/Matter USB Dongle**: Examples include:
  - Nordic nRF52840 DK or Dongle
  - Silicon Labs EFR32 series
  - Texas Instruments CC2652/CC1352
  - Any Thread-compatible radio module

### Required Software
- Docker Engine 20.10+
- Docker Compose 2.0+
- Linux host (for USB device access)

### Check Docker Installation
```bash
docker --version
docker compose version
```

---

## IPv6 Requirements

### Why Matter Needs IPv6

**Matter protocol is built entirely on IPv6:**
- Thread networks use IPv6 exclusively (6LoWPAN)
- Matter devices communicate using IPv6 link-local addresses
- Device commissioning requires IPv6 multicast/unicast
- The Border Router bridges Thread's IPv6 network to your infrastructure

### Current Docker Setup: Host Networking

**Good news: The docker-compose.yml already handles IPv6 correctly!**

Both `otbr` and `matter2mqtt` use **`network_mode: host`**, which means:
- ✅ Containers share the host's IPv6 stack directly
- ✅ No Docker network virtualization (full IPv6 access)
- ✅ Can use link-local addresses for Matter commissioning
- ✅ OTBR can create Thread network interfaces on the host

### Communication Paths

```
Host IPv6 Stack (Linux kernel)
     ↓
┌────┴─────────────┐
│   matter2mqtt    │ ←→ IPv6 ←→ Matter Devices (via Thread)
│  (host network)  │
└─────┬────────────┘
      │ USB Serial
      ↓
   Dongle (Radio) ←→ Thread Network (IPv6/6LoWPAN) ←→ Matter Devices
```

**Key Point:** USB dongle communication is **serial/UART**, not TCP/IP. IPv6 is only needed for:
1. Matter device discovery (mDNS over IPv6)
2. Matter commissioning (IPv6 unicast)
3. Thread routing (OTBR advertises IPv6 routes)

### Verify Your System Has IPv6

**Run the verification script:**

```bash
./scripts/verify-ipv6.sh
```

This checks:
- IPv6 kernel support
- IPv6 sysctl configuration
- IPv6 addresses (global and link-local)
- IPv6 forwarding (for OTBR)
- IPv6 routing table
- Docker IPv6 support
- mDNS/Avahi service
- USB devices

**Manual checks:**

```bash
# Check if IPv6 is enabled
cat /proc/sys/net/ipv6/conf/all/disable_ipv6
# Should output: 0 (enabled)

# List IPv6 addresses
ip -6 addr show

# Check for link-local addresses (fe80::)
ip -6 addr show scope link

# Test IPv6 connectivity (optional)
ping6 -c 3 2001:4860:4860::8888
```

### Enable IPv6 on Linux Host

If IPv6 is disabled on your host:

```bash
# Temporarily enable IPv6
sudo sysctl -w net.ipv6.conf.all.disable_ipv6=0
sudo sysctl -w net.ipv6.conf.default.disable_ipv6=0

# Permanently enable IPv6
echo "net.ipv6.conf.all.disable_ipv6=0" | sudo tee -a /etc/sysctl.conf
echo "net.ipv6.conf.default.disable_ipv6=0" | sudo tee -a /etc/sysctl.conf
sudo sysctl -p

# Enable IPv6 forwarding (required for OTBR)
sudo sysctl -w net.ipv6.conf.all.forwarding=1
echo "net.ipv6.conf.all.forwarding=1" | sudo tee -a /etc/sysctl.conf
```

### Do You Need Global IPv6?

**Short answer: No!**

- **Link-local IPv6 (fe80::) is sufficient** for Matter commissioning
- Matter devices use link-local addresses for initial pairing
- OTBR creates a Thread network with its own IPv6 prefix
- Global IPv6 connectivity is optional (only for remote access)

### Does My LAN/Router Need IPv6 Routing?

**Critical Question:** Does my network router need IPv6 enabled? Does my LAN need IPv6 routing?

**Answer: NO!** Thread network is **isolated from your LAN**.

**Why this works:**

```
Your LAN (IPv4 only) ←→ Computer ←→ Docker ←→ USB ←→ Thread Radio Network (IPv6)
                                                          ↓
                                                    Matter Devices

• LAN traffic: MQTT over IPv4 (matter/topic ← "data")
• Thread traffic: IPv6 over 802.15.4 radio (isolated, not on Ethernet/Wi-Fi)
• The USB dongle bridges these via serial (not network routing)
```

**Communication flow:**
1. Matter device (Thread IPv6) → Radio → USB dongle
2. USB dongle → Serial/UART → matter2mqtt container
3. matter2mqtt → IPv4 MQTT localhost:1883 → Mosquitto
4. Mosquitto → IPv4 over your LAN → Other devices

**Your LAN only carries IPv4 MQTT traffic!** The Thread network (IPv6) is completely separate, using 802.15.4 radio, not Ethernet/Wi-Fi.

**See:** [docs/NETWORK_ISOLATION.md](docs/NETWORK_ISOLATION.md) for detailed diagrams and explanation.

**Summary:**
- ✅ LAN can be IPv4-only
- ✅ Router doesn't need IPv6 routing
- ✅ ISP doesn't need to provide IPv6
- ✅ Only host needs link-local IPv6 (fe80::) - enabled by default
- ✅ Thread network is isolated via radio, not routed through your LAN

### What About Docker Bridge Networks?

If you wanted to use Docker bridge networking instead of host networking (not recommended for Matter), you would need to enable IPv6:

```yaml
# Example: NOT used in current setup
networks:
  matter-net:
    driver: bridge
    enable_ipv6: true
    ipam:
      config:
        - subnet: fd00:dead:beef::/48
          gateway: fd00:dead:beef::1
```

**However**, this is **NOT needed** because:
- OTBR **requires** host networking for border router functionality
- matter2mqtt **requires** host networking for mDNS discovery
- The current setup is correct as-is

### Common IPv6 Issues

**Issue: IPv6 disabled in kernel**
```bash
# Check boot parameters
cat /proc/cmdline | grep ipv6.disable

# If shows ipv6.disable=1, edit GRUB config:
sudo nano /etc/default/grub
# Remove or change: ipv6.disable=1 to ipv6.disable=0
sudo update-grub
sudo reboot
```

**Issue: No IPv6 addresses**
```bash
# Check if NetworkManager is managing interfaces
nmcli device status

# Enable IPv6 on interface (example: eth0)
nmcli connection modify eth0 ipv6.method auto
nmcli connection up eth0
```

**Issue: IPv6 works on host but not in container**
- This should NOT happen with `network_mode: host`
- Verify with: `docker exec -it matter2mqtt ip -6 addr show`
- If different from host output, check Docker daemon config

---

## Quick Start

### 1. Create Configuration Files

Create your `config.yaml`:
```bash
cp config.yaml.example config.yaml
```

Edit `config.yaml`:
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

Create your `devices.yaml`:
```bash
cp devices.yaml.example devices.yaml
```

### 2. Verify IPv6 is Enabled

**IMPORTANT**: Matter requires IPv6. Run the verification script:

```bash
./scripts/verify-ipv6.sh
```

If the script reports issues, follow the instructions in the [IPv6 Requirements](#ipv6-requirements) section above.

**Quick fix if IPv6 is disabled:**
```bash
sudo sysctl -w net.ipv6.conf.all.disable_ipv6=0
sudo sysctl -w net.ipv6.conf.all.forwarding=1
```

### 3. Identify Your USB Device

**CRITICAL STEP**: You need to find the USB device path for your Thread/Matter dongle.

```bash
# Method 1: By ID (RECOMMENDED - persistent across reboots)
ls -l /dev/serial/by-id/

# Output example:
# lrwxrwxrwx 1 root root 13 Nov  5 10:00 usb-SEGGER_J-Link_000683012345-if00 -> ../../ttyACM0

# Method 2: By path
ls -l /dev/serial/by-path/

# Method 3: Check dmesg after plugging in device
dmesg | grep -i "tty\|usb"

# Method 4: List all USB serial devices
ls -l /dev/tty{ACM,USB}*
```

**Important Notes:**
- `/dev/ttyACM0`, `/dev/ttyUSB0`, etc. may change after reboot
- Using `/dev/serial/by-id/...` is MUCH more reliable
- Note the device name for the next step

### 4. Configure USB Device in Docker Compose

Create a `.env` file to specify your USB device:

```bash
# .env file
THREAD_DEVICE=/dev/serial/by-id/usb-Nordic_Semiconductor_nRF52840_DK_000680012345
MATTER_DEVICE=/dev/serial/by-id/usb-Nordic_Semiconductor_nRF52840_DK_000680012345
```

**Or** edit `docker-compose.yml` directly and replace:
```yaml
devices:
  - ${THREAD_DEVICE:-/dev/ttyACM0}:/dev/ttyACM0
```

With your actual device:
```yaml
devices:
  - /dev/serial/by-id/usb-Nordic_Semiconductor_nRF52840_DK_000680012345:/dev/ttyACM0
```

### 5. Build and Start Services

```bash
# Build the matter2mqtt image
docker compose build

# Start all services (detached mode)
docker compose up -d

# View logs
docker compose logs -f

# View logs for specific service
docker compose logs -f matter2mqtt
docker compose logs -f otbr
```

### 6. Verify Services are Running

```bash
# Check service status
docker compose ps

# Should show:
# - mosquitto (running)
# - otbr (running)
# - matter2mqtt (running)

# Check MQTT broker
mosquitto_sub -h localhost -p 1883 -t "matter/#" -v

# Check OTBR web interface (if enabled)
# Open browser: http://localhost:8081
```

---

## USB Device Configuration

### Understanding Device Passthrough

Docker containers need explicit permission to access USB devices. The `devices:` section in `docker-compose.yml` maps host devices into containers.

### Single Dongle (Thread + Matter combined)

If you have ONE dongle that handles both Thread and Matter:

```yaml
# In docker-compose.yml
services:
  otbr:
    devices:
      - /dev/serial/by-id/usb-YOUR_DEVICE:/dev/ttyACM0

  matter2mqtt:
    devices:
      - /dev/serial/by-id/usb-YOUR_DEVICE:/dev/ttyACM0
```

Both services share the same physical device.

### Multiple Dongles (Separate Thread and Matter)

If you have SEPARATE dongles:

```yaml
# In docker-compose.yml
services:
  otbr:
    devices:
      - /dev/serial/by-id/usb-THREAD_DEVICE:/dev/ttyACM0

  matter2mqtt:
    devices:
      - /dev/serial/by-id/usb-MATTER_DEVICE:/dev/ttyACM1
```

Make sure to update the device paths inside the containers accordingly.

### Device Permissions

If you get "Permission Denied" errors:

```bash
# Check device ownership
ls -l /dev/ttyACM0

# Add your user to dialout group
sudo usermod -aG dialout $USER

# Or change device permissions (temporary)
sudo chmod 666 /dev/ttyACM0

# Or use udev rules (permanent)
# Create /etc/udev/rules.d/99-usb-serial.rules:
SUBSYSTEM=="tty", ATTRS{idVendor}=="1366", ATTRS{idProduct}=="1015", MODE="0666"

# Reload udev rules
sudo udevadm control --reload-rules
sudo udevadm trigger
```

### Finding Vendor/Product IDs

```bash
# Get USB device info
lsusb

# Output example:
# Bus 001 Device 005: ID 1366:1015 SEGGER J-Link

# Detailed info
lsusb -v -d 1366:1015
```

---

## Service Architecture

### Components

```
┌─────────────────────────────────────────────────────────┐
│                    Docker Compose Stack                  │
├─────────────────────────────────────────────────────────┤
│                                                           │
│  ┌──────────────┐    ┌──────────────┐    ┌───────────┐  │
│  │  Mosquitto   │◄───│ matter2mqtt  │◄───│ USB       │  │
│  │  MQTT Broker │    │  Bridge      │    │ Matter    │  │
│  │  :1883       │    │              │    │ Dongle    │  │
│  └──────┬───────┘    └──────────────┘    └───────────┘  │
│         │                                                 │
│         │            ┌──────────────┐    ┌───────────┐  │
│         └────────────►│   OTBR       │◄───│ USB       │  │
│                      │   Thread BR  │    │ Thread    │  │
│                      │   :8080      │    │ Dongle    │  │
│                      └──────────────┘    └───────────┘  │
│                                                           │
└─────────────────────────────────────────────────────────┘
```

### Network Mode: `host`

Both OTBR and matter2mqtt use `network_mode: host` because:
- OTBR needs direct access to the host network for border router functionality
- mDNS/Avahi service discovery requires host networking
- Matter commissioning uses link-local IPv6 addresses

**Implications:**
- Containers share host's network namespace
- Ports are directly exposed on the host
- Some firewall rules may be needed

### Privileged Mode

Both services run with `privileged: true` to:
- Access USB devices without complex permission mapping
- Manage network interfaces (for OTBR)
- Handle Thread/Matter radio control

**Security Note**: For production, consider using specific capabilities instead:
```yaml
cap_add:
  - NET_ADMIN
  - NET_RAW
devices:
  - /dev/ttyACM0:/dev/ttyACM0
```

---

## Configuration

### Environment Variables

Create `.env` file in the same directory as `docker-compose.yml`:

```bash
# USB Devices
THREAD_DEVICE=/dev/serial/by-id/usb-YOUR_THREAD_DEVICE
MATTER_DEVICE=/dev/serial/by-id/usb-YOUR_MATTER_DEVICE

# OTBR Configuration
NETWORK_NAME=MyThreadNetwork
CHANNEL=15
PANID=0x1234
EXTPANID=1111111122222222
NETWORKKEY=00112233445566778899aabbccddeeff

# matter2mqtt Configuration
# MOCK_CHIPTOOL=true  # Uncomment for testing without hardware
```

### Volume Mounts

Persistent data is stored in Docker volumes:
- `mosquitto_data`: MQTT message persistence
- `mosquitto_logs`: MQTT broker logs
- `otbr_data`: Thread network state
- `matter_data`: Matter device credentials and state

**To inspect volumes:**
```bash
docker volume ls
docker volume inspect matter2mqtt_matter_data
```

**To backup volumes:**
```bash
docker run --rm -v matter2mqtt_matter_data:/data -v $(pwd):/backup ubuntu tar czf /backup/matter_data_backup.tar.gz /data
```

---

## Troubleshooting

### Issue: Container can't access USB device

**Symptoms:**
```
Error: Could not open device /dev/ttyACM0
Permission denied
```

**Solutions:**
1. Verify device path:
   ```bash
   ls -l /dev/serial/by-id/
   ```

2. Check device is not in use:
   ```bash
   lsof /dev/ttyACM0
   ```

3. Restart with device unplugged, then plug back in:
   ```bash
   docker compose down
   # Unplug USB device
   # Plug USB device back in
   docker compose up -d
   ```

4. Check container can see device:
   ```bash
   docker exec -it matter2mqtt ls -l /dev/ttyACM0
   ```

### Issue: OTBR fails to start

**Check logs:**
```bash
docker compose logs otbr
```

**Common issues:**
- Wrong device path in `devices:` section
- Device already in use by host services (ModemManager, etc.)
- Incompatible firmware on Thread dongle

**Disable ModemManager (often interferes):**
```bash
sudo systemctl stop ModemManager
sudo systemctl disable ModemManager
```

### Issue: matter2mqtt can't find chip-tool

**Symptoms:**
```
Error: chip-tool binary not found
```

**Solution:**
You need to provide chip-tool binary. Options:

**Option 1: Build chip-tool and mount it**
```bash
# Build chip-tool (see chip-tool.md)
# Then mount it in docker-compose.yml:
volumes:
  - /usr/local/bin/chip-tool:/usr/local/bin/chip-tool:ro
```

**Option 2: Use mock mode for testing**
```yaml
environment:
  - MOCK_CHIPTOOL=true
```

**Option 3: Build chip-tool into Docker image**
Update Dockerfile to include chip-tool build steps.

### Issue: No MQTT messages appearing

**Debug steps:**
```bash
# Check MQTT broker is accessible
docker exec -it mosquitto mosquitto_sub -t "matter/#" -v

# Check matter2mqtt logs
docker compose logs -f matter2mqtt

# Check device subscriptions were created
docker exec -it matter2mqtt ps aux | grep chip-tool

# Test MQTT manually
docker exec -it mosquitto mosquitto_pub -t "matter/test" -m "hello"
```

### Issue: Services keep restarting

```bash
# Check which service is failing
docker compose ps

# View recent logs
docker compose logs --tail=50 matter2mqtt

# Check resource usage
docker stats
```

### View Container Shell

```bash
# Access matter2mqtt container
docker exec -it matter2mqtt /bin/bash

# Access OTBR container
docker exec -it otbr /bin/bash

# List processes
docker exec -it matter2mqtt ps aux
```

---

## Advanced Configuration

### Custom Network Setup

If you don't want `host` networking:

```yaml
services:
  mosquitto:
    networks:
      - matter-net
    ports:
      - "1883:1883"

  otbr:
    network_mode: host  # OTBR still needs host mode

  matter2mqtt:
    networks:
      - matter-net
    depends_on:
      - mosquitto

networks:
  matter-net:
    driver: bridge
```

### Adding Additional Services

Example: Add Home Assistant:

```yaml
services:
  homeassistant:
    image: ghcr.io/home-assistant/home-assistant:stable
    container_name: homeassistant
    network_mode: host
    volumes:
      - ./homeassistant:/config
    restart: unless-stopped
    depends_on:
      - mosquitto
```

### Production Security

```yaml
services:
  mosquitto:
    volumes:
      - ./mosquitto.conf:/mosquitto/config/mosquitto.conf:ro
      - ./mosquitto_passwd:/mosquitto/config/passwd:ro

  matter2mqtt:
    environment:
      - MQTT_USERNAME=matter
      - MQTT_PASSWORD_FILE=/run/secrets/mqtt_password
    secrets:
      - mqtt_password

secrets:
  mqtt_password:
    file: ./secrets/mqtt_password.txt
```

---

## Commands Reference

```bash
# Start services
docker compose up -d

# Stop services
docker compose down

# Restart a single service
docker compose restart matter2mqtt

# View logs
docker compose logs -f

# Rebuild after code changes
docker compose build --no-cache
docker compose up -d

# Clean up everything
docker compose down -v  # WARNING: Deletes volumes!

# Update images
docker compose pull
docker compose up -d
```

---

## Next Steps

1. Commission Matter devices using `chip-tool`
2. Configure devices in `devices.yaml`
3. Monitor MQTT messages: `mosquitto_sub -h localhost -t "matter/#" -v`
4. Integrate with Home Assistant or other automation platforms

For more information:
- [README.md](README.md) - Project overview
- [chip-tool.md](chip-tool.md) - How to get/build chip-tool
- [OTBR Documentation](https://openthread.io/guides/border-router)
