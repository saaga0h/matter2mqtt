# macOS Docker Networking for matter2mqtt

## Critical Differences: macOS vs Linux

### Docker on macOS Uses a VM

**Unlike Linux,** Docker Desktop on macOS/Windows runs containers inside a Linux VM:

```
┌─────────────────────────────────────────────────┐
│                macOS Host                        │
│  ┌───────────────────────────────────────────┐  │
│  │   Docker Desktop (Hypervisor)             │  │
│  │  ┌─────────────────────────────────────┐  │  │
│  │  │      Linux VM                       │  │  │
│  │  │  ┌───────────────┐  ┌─────────────┐ │  │  │
│  │  │  │  matter2mqtt  │  │    OTBR     │ │  │  │
│  │  │  └───────────────┘  └─────────────┘ │  │  │
│  │  └─────────────────────────────────────┘  │  │
│  └───────────────────────────────────────────┘  │
└─────────────────────────────────────────────────┘
```

### `network_mode: host` Limitation

**On Linux:**
- Container shares host's network namespace directly
- Container sees exact same `ip addr` output as host

**On macOS/Windows:**
- Container shares the **VM's** network namespace, not macOS
- `network_mode: host` means "host to the VM", not "host to Mac"
- This is a Docker Desktop limitation, not a bug

**Implication:** The containers can't directly access macOS network interfaces, but this is usually fine for Matter/Thread.

### USB Device Passthrough on macOS

**Docker Desktop for Mac has limitations:**

1. **USB passthrough is not fully supported** natively
2. **Options:**
   - Use USB/IP over network (complex)
   - Use Docker Desktop with VirtualBox backend (deprecated)
   - Run on Linux (VM or bare metal)
   - Use macOS native chip-tool (mount binary into container)

**Recommended approach for macOS:**

```yaml
# Option 1: Mount macOS chip-tool binary
matter2mqtt:
  volumes:
    - /usr/local/bin/chip-tool:/usr/local/bin/chip-tool:ro
  # Don't use devices: section on macOS (won't work)

# Option 2: Use mock mode for development
matter2mqtt:
  environment:
    - MOCK_CHIPTOOL=true
```

**For production Matter/Thread:** Consider running on:
- Raspberry Pi with Linux
- Linux VM on Mac (UTM, Parallels, VMware Fusion)
- Native Linux server

### IPv6 on macOS

**Good news:** macOS handles IPv6 well:

```bash
# Check IPv6 on Mac
ifconfig | grep inet6

# System Preferences → Network → Advanced → TCP/IP
# IPv6: Configure Automatically (default)
```

**What you get:**
- ✅ Link-local IPv6 (fe80::) - always enabled
- ✅ ULA if router advertises (fc00::, fd00::)
- ⚠️ Global IPv6 only if ISP/router provides it

**For Matter/Thread:** Link-local is sufficient!

### LAN IPv6 Requirements

**Question:** Does my LAN need IPv6 enabled?

**Answer:** **NO!** Here's why:

#### Scenario 1: Matter over Thread (Your Case)

```
Matter Device (Thread network: fd11:22::1234)
    ↓ (802.15.4 radio - NOT Ethernet/WiFi)
USB Dongle (Serial communication)
    ↓ (USB/Serial - NOT network)
matter2mqtt container
    ↓ (MQTT over IPv4: localhost:1883)
MQTT Broker
```

**IPv6 is ONLY used:**
- Within the Thread radio network (isolated from LAN)
- For Matter commissioning (link-local scope)
- Inside Docker VM (has IPv6 regardless of LAN)

**Your LAN traffic:** IPv4 only (MQTT, DNS, HTTP, etc.)

#### Scenario 2: Matter over Wi-Fi (Not Thread)

If you had Matter-over-Wi-Fi devices (not Thread):
- Then devices would be on your LAN
- Would need LAN to support IPv6
- But you're using Thread, so this doesn't apply!

### Network Isolation Verification

**Thread network is completely isolated from your LAN:**

```bash
# On Mac (after OTBR is running)
# You won't see Thread devices in:
ping6 -c 1 fd11:22::1234  # Won't work from Mac
# Because Thread network is inside Docker VM/USB dongle

# But matter2mqtt can reach them via chip-tool:
docker exec -it matter2mqtt chip-tool onoff read on-off 12345 1
# This works! Uses USB serial, not network
```

### What Gets Routed Where

| Traffic Type | Protocol | Path | Needs LAN IPv6? |
|--------------|----------|------|-----------------|
| MQTT messages | IPv4 | matter2mqtt → Mosquitto (localhost) | ❌ No |
| Matter commissioning | IPv6 link-local | chip-tool → USB dongle → Thread radio | ❌ No |
| Thread network | IPv6 (fd11:22::/64) | USB dongle → Matter devices (radio) | ❌ No |
| Device commands | Serial + IPv6 | chip-tool → USB → dongle → Thread | ❌ No |
| Web browsing from Mac | IPv4 | Mac → Router → Internet | ❌ No |
| SSH to Mac | IPv4 | LAN → Mac | ❌ No |

**Nothing crosses from Thread network to your LAN!**

### Configuration for macOS

**Minimal setup:**

```yaml
# docker-compose.yml
version: '3.8'

services:
  mosquitto:
    image: eclipse-mosquitto:2
    ports:
      - "1883:1883"
    volumes:
      - ./mosquitto.conf:/mosquitto/config/mosquitto.conf:ro
    restart: unless-stopped

  # OTBR - Limited functionality on macOS due to USB
  otbr:
    image: openthread/otbr:latest
    privileged: true
    # Note: USB passthrough may not work on macOS
    # Consider running OTBR on separate Linux device
    restart: unless-stopped

  matter2mqtt:
    build: .
    volumes:
      - ./config.yaml:/app/config/config.yaml:ro
      - ./devices.yaml:/app/config/devices.yaml:ro
      # Mount macOS chip-tool if available
      - /usr/local/bin/chip-tool:/usr/local/bin/chip-tool:ro
    environment:
      # Use mock mode if USB doesn't work
      - MOCK_CHIPTOOL=false
    restart: unless-stopped
```

### Production Recommendation

**For reliable Matter/Thread operation:**

**Option 1: Raspberry Pi as Border Router**
```
Mac (development) ←→ Network ←→ Raspberry Pi (OTBR + matter2mqtt)
                                      ↓ USB
                                Thread Dongle
                                      ↓ Radio
                                Matter Devices
```

**Option 2: Linux VM on Mac**
```
macOS Host
  └─ Linux VM (UTM/Parallels)
      └─ Docker with USB passthrough
          └─ OTBR + matter2mqtt
```

**Option 3: macOS Native**
```
macOS
  ├─ Docker (Mosquitto only)
  ├─ OTBR (run natively with Homebrew)
  └─ matter2mqtt (run natively with go run)
```

### Summary for macOS Users

✅ **What works:**
- MQTT broker (Mosquitto) in Docker
- IPv6 is available (link-local minimum)
- matter2mqtt can run in Docker (with caveats)

⚠️ **What's limited:**
- USB passthrough to Docker containers
- `network_mode: host` doesn't mean Mac's network
- OTBR may not work in Docker

❌ **What's NOT needed:**
- Global IPv6 on LAN
- IPv6 routing on your router
- IPv6 from ISP

**Best approach:** Develop on Mac, deploy on Linux for production.

### Testing IPv6 on macOS

```bash
# Verify IPv6 is active
ifconfig en0 | grep inet6
# Should show: inet6 fe80::xxxx

# Check Docker VM has IPv6
docker run --rm alpine ip -6 addr show
# Should show: inet6 ::1/128 scope host (loopback)
#              inet6 fe80::xxxx scope link

# Verify link-local is sufficient
ping6 -c 1 fe80::1%en0
# Should work (pinging your own Mac)

# LAN doesn't need routing
ping6 -c 1 2001:4860:4860::8888
# May fail - that's OK! Link-local is enough.
```

### Troubleshooting macOS

**Issue: USB device not visible in container**
```bash
# Check on Mac
ls -l /dev/tty.*

# Docker Desktop doesn't support USB passthrough directly
# Solutions:
# 1. Use USB/IP network protocol
# 2. Run matter2mqtt natively on macOS
# 3. Use Linux VM with USB passthrough
```

**Issue: OTBR won't start**
```bash
# OTBR expects Linux network interfaces
# On macOS, run OTBR outside Docker:
brew install openthread
# Follow OpenThread macOS instructions
```

**Issue: IPv6 not working**
```bash
# Check macOS IPv6 settings
networksetup -getinfo "Wi-Fi"
# Should show: IPv6: Automatic

# Enable IPv6 if disabled
networksetup -setv6automatic "Wi-Fi"

# Verify
ping6 -c 1 ::1
# Should work
```

### Next Steps for macOS Users

1. **Development:**
   - Run MQTT in Docker ✅
   - Run matter2mqtt with mock mode ✅
   - Test logic without hardware

2. **Hardware testing:**
   - Use Linux (VM or separate device)
   - Pass USB through to Linux VM
   - Run full stack with real Thread devices

3. **Production:**
   - Deploy to Raspberry Pi or Linux server
   - Keep Mac as development environment
   - Use mock mode for testing logic
