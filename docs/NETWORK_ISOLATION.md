# Matter/Thread Network Isolation

## Question: Does my LAN need IPv6 enabled?

**Answer: NO!** Thread network is isolated from your LAN.

## Network Architecture Diagram

```
╔═══════════════════════════════════════════════════════════════════╗
║                  YOUR LAN (IPv4 only is fine!)                    ║
║                                                                   ║
║  Internet ←→ Router (IPv4) ←→ Switch ←→ Devices (IPv4)          ║
║                                   ↓                               ║
║                              Your Computer                        ║
║                         (IPv4 + link-local IPv6)                 ║
╚═════════════════════════════════╦═════════════════════════════════╝
                                  │
                                  │ Only IPv4 traffic on LAN
                                  │ (MQTT, HTTP, SSH, etc.)
                                  │
╔═════════════════════════════════╩═════════════════════════════════╗
║                         Docker Environment                        ║
║                        (Has own IPv6 stack)                       ║
║                                                                   ║
║  ┌─────────────────┐         ┌──────────────────────┐           ║
║  │   Mosquitto     │         │   matter2mqtt        │           ║
║  │   MQTT Broker   │◄────────│   Bridge             │           ║
║  │                 │  IPv4   │                      │           ║
║  │ localhost:1883  │         │  chip-tool process   │           ║
║  └─────────────────┘         └──────────┬───────────┘           ║
║                                          │                       ║
║                                          │ USB Serial            ║
║                                          │ (NOT network!)        ║
║                              ┌───────────▼────────────┐          ║
║                              │   USB Thread Dongle    │          ║
║                              │   (Radio Hardware)     │          ║
║                              └───────────┬────────────┘          ║
╚══════════════════════════════════════════╪═══════════════════════╝
                                           │
                                           │ 802.15.4 Radio
                                           │ (Wireless, NOT Ethernet!)
                                           │
╔══════════════════════════════════════════╧═══════════════════════╗
║              THREAD NETWORK (Isolated IPv6)                      ║
║              Prefix: fd11:22::/64 (example)                      ║
║          NOT connected to your LAN or Internet!                  ║
║                                                                   ║
║  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐          ║
║  │ Light Bulb   │  │ Temp Sensor  │  │  Door Lock   │          ║
║  │fd11:22::0001 │  │fd11:22::0002 │  │fd11:22::0003 │          ║
║  └──────────────┘  └──────────────┘  └──────────────┘          ║
║                                                                   ║
║  All devices only reachable via chip-tool through USB dongle     ║
╚═══════════════════════════════════════════════════════════════════╝
```

## Communication Paths

### Path 1: Matter Device → MQTT (The Bridge)

```
┌─────────────────────────────────────────────────────────────────┐
│                    End-to-End Flow                               │
└─────────────────────────────────────────────────────────────────┘

Step 1: Matter device sends data
   Matter Device (fd11:22::1234)
        ↓ [802.15.4 Radio - IPv6 packet over Thread]
   USB Dongle receives radio packet

Step 2: USB dongle communicates via serial
   USB Dongle
        ↓ [USB Serial - NOT network, just UART bytes]
   matter2mqtt reads from /dev/ttyACM0

Step 3: matter2mqtt publishes to MQTT
   matter2mqtt (chip-tool)
        ↓ [TCP/IP IPv4 - localhost:1883]
   Mosquitto MQTT Broker

Step 4: MQTT subscribers receive
   Mosquitto
        ↓ [TCP/IP IPv4 - over LAN]
   Home Assistant / Other subscribers on LAN

═══════════════════════════════════════════════════════════════════
NOTICE: IPv6 is only used in Step 1 (Thread radio network)
        Your LAN only sees Step 3-4 (IPv4 MQTT traffic)
═══════════════════════════════════════════════════════════════════
```

### Path 2: Command to Matter Device

```
┌─────────────────────────────────────────────────────────────────┐
│              Sending Command (Reverse Flow)                      │
└─────────────────────────────────────────────────────────────────┘

Step 1: MQTT command arrives
   Home Assistant publishes MQTT
        ↓ [TCP/IP IPv4 - over LAN]
   Mosquitto receives on matter/light/set

Step 2: matter2mqtt processes command
   matter2mqtt subscribes to topic
        ↓ [Execute chip-tool command]
   chip-tool onoff on 12345 1

Step 3: chip-tool sends via USB
   chip-tool process
        ↓ [USB Serial - UART bytes]
   USB Dongle receives command

Step 4: Dongle transmits via Thread
   USB Dongle
        ↓ [802.15.4 Radio - IPv6 packet]
   Matter Device (fd11:22::1234) executes command

═══════════════════════════════════════════════════════════════════
NOTICE: Your LAN only carries MQTT (IPv4) in Steps 1-2
        Thread network (IPv6) is isolated in Step 4
═══════════════════════════════════════════════════════════════════
```

## What IPv6 is Needed Where?

| Layer | IPv6 Required? | Why / Why Not |
|-------|----------------|---------------|
| **Your LAN** | ❌ NO | Only MQTT traffic (IPv4) passes through LAN |
| **Your Router** | ❌ NO | Doesn't route Thread traffic (isolated radio network) |
| **Your ISP** | ❌ NO | Matter devices don't connect to Internet |
| **Host Computer** | ⚠️ Link-Local | Just needs fe80:: for Docker (enabled by default) |
| **Docker Containers** | ✅ YES | Needs IPv6 stack for chip-tool (Docker provides this) |
| **Thread Network** | ✅ YES | Thread IS an IPv6 network (6LoWPAN) |
| **Matter Devices** | ✅ YES | Matter protocol requires IPv6 |

## Common Misconceptions

### ❌ Myth 1: "My router needs to route IPv6"
**Reality:** Your router never sees Thread traffic. Thread uses 802.15.4 radio, not Ethernet/Wi-Fi.

### ❌ Myth 2: "My ISP needs to provide IPv6"
**Reality:** Matter devices don't connect to the Internet. They're local-only via Thread.

### ❌ Myth 3: "All devices on my LAN need IPv6"
**Reality:** Only the computer running Docker needs link-local IPv6 (enabled by default on Linux/Mac/Windows).

### ❌ Myth 4: "MQTT broker needs IPv6"
**Reality:** MQTT runs over IPv4 (localhost:1883 or LAN IP).

### ✅ Truth: "Thread network is a separate IPv6 network"
Thread creates its own isolated IPv6 network using radio (802.15.4), completely separate from your Ethernet/Wi-Fi LAN.

## Technical Details

### Thread Network Prefix

OTBR creates a ULA prefix for the Thread network:

```bash
# Example Thread network prefix
fd11:22:33:44::/64

# Devices get addresses from this prefix
Light Bulb:   fd11:22:33:44::1234:5678
Temp Sensor:  fd11:22:33:44::abcd:ef01
Door Lock:    fd11:22:33:44::9876:5432
```

**This prefix is NOT routed to your LAN** (unless you explicitly configure OTBR to do so, which matter2mqtt doesn't need).

### Border Router's Job

OTBR acts as a "border" between two networks:

```
   LAN Side                OTBR (Border)           Thread Side
(Ethernet/Wi-Fi)                                 (802.15.4 Radio)
     │                         │                        │
     │                    ┌────┴────┐                   │
     │                    │ Routing │                   │
     │                    │ Decision│                   │
     │                    └────┬────┘                   │
     │                         │                        │
IPv4 │◄───────MQTT─────────────┤                        │
     │                         │                        │
     │                         ├──────────Thread────────►│ IPv6
     │                         │      (via USB)         │
     │                         │                        │
```

**For matter2mqtt:** OTBR only forwards Thread → USB → matter2mqtt, NOT Thread → LAN.

### Why Link-Local IPv6 is Sufficient

**Link-local (fe80::) scope:**
- Allows IPv6 communication on the same link (same network segment)
- Docker containers can use link-local to communicate within Docker
- chip-tool uses link-local for Matter commissioning

**You don't need:**
- Global IPv6 address (2001::/3)
- Unique Local Address beyond Thread network (fc00::/7)
- IPv6 routing on your LAN

### Verification Commands

**Check what IPv6 you have:**

```bash
# On host (Linux/Mac)
ip -6 addr show  # Linux
ifconfig | grep inet6  # Mac

# Expected minimum:
# inet6 ::1/128 scope host (loopback) ← Always present
# inet6 fe80::xxxx/64 scope link      ← Link-local, all you need!

# You DON'T need:
# inet6 2001:xxxx/64 scope global     ← Global (not needed)
```

**Verify Thread network is isolated:**

```bash
# From your computer, try to ping Thread device
ping6 fd11:22::1234
# This will FAIL (timeout) - that's correct!
# Thread network is isolated from your LAN

# But matter2mqtt CAN reach it via chip-tool:
docker exec matter2mqtt chip-tool onoff read on-off 12345 1
# This WORKS - uses USB serial, not network ping
```

**Check MQTT traffic on LAN:**

```bash
# Capture traffic on LAN interface
sudo tcpdump -i eth0 port 1883

# You'll see:
# - IPv4 packets only
# - MQTT protocol (CONNECT, PUBLISH, SUBSCRIBE)
# - NO IPv6 packets
# - NO Thread network traffic (it's on radio, not Ethernet!)
```

## Scenarios

### Scenario 1: IPv4-only LAN (Your Case)

```
Configuration:
- Router: IPv4 only, no IPv6 routing
- LAN: All devices use IPv4
- Computer: Has link-local IPv6 (fe80::)
- Thread: Isolated IPv6 network (fd11:22::/64)

Result: ✅ Everything works!
- MQTT uses IPv4 on LAN
- Thread uses IPv6 on radio (isolated)
- matter2mqtt bridges between them
```

### Scenario 2: Dual-stack LAN (IPv4 + IPv6)

```
Configuration:
- Router: IPv4 + IPv6 routing
- LAN: All devices have both IPv4 and IPv6
- Computer: Has global IPv6 (2001:xxxx)
- Thread: Isolated IPv6 network (fd11:22::/64)

Result: ✅ Everything works!
- MQTT can use IPv4 or IPv6 on LAN
- Thread still uses its own IPv6 prefix (isolated)
- matter2mqtt still bridges between them
- Thread devices NOT accessible from LAN directly
```

### Scenario 3: IPv6 disabled on host

```
Configuration:
- Computer: IPv6 completely disabled
- sysctl net.ipv6.conf.all.disable_ipv6=1

Result: ❌ Won't work!
- chip-tool requires IPv6 stack
- Docker needs IPv6 for Thread network
- Fix: Enable link-local IPv6 (see IPv6 Requirements doc)
```

### Scenario 4: Want Thread devices on LAN

```
Configuration:
- You WANT Thread devices reachable from LAN
- Example: Ping Matter devices from other computers

Requirement:
- LAN needs IPv6 routing
- OTBR needs to advertise Thread prefix to LAN
- Configure OTBR with BACKBONE_ROUTER=1

This is OPTIONAL and NOT needed for matter2mqtt!
```

## Summary: Do I Need IPv6 Routing?

**Quick decision tree:**

```
Do you need IPv6 routing on your LAN/router?
│
├─ Are you using matter2mqtt? → NO, link-local IPv6 is enough
│
├─ Do you only access Matter devices via MQTT? → NO
│
├─ Do you want to ping/access Matter devices directly from LAN? → YES
│
└─ Do you want Matter devices to access Internet? → YES (rare)
```

**For 99% of matter2mqtt users:**
- ✅ Keep your LAN IPv4-only
- ✅ Enable link-local IPv6 on host (default)
- ✅ Let Thread network stay isolated
- ✅ Use MQTT to control devices (IPv4)

**Your router config:**
```
IPv6: Disabled  ← This is fine!
Thread network stays isolated from LAN
```
