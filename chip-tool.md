# Building chip-tool

chip-tool is the reference Matter controller from the connectedhomeip project. This guide covers obtaining pre-built binaries or building from source.

## Pre-built Binaries (Recommended)

**Official releases:**
- Download from [matter2mqtt releases](https://github.com/your-org/matter2mqtt/releases)
- Includes chip-tool bundled with matter2mqtt

**Supported architectures:**
- `linux-arm64` - Raspberry Pi 3/4/5, ARM64 servers
- `linux-amd64` - x86_64 Linux systems
- `darwin-arm64` - Apple Silicon Macs
- `darwin-amd64` - Intel Macs

Extract and verify:
```bash
tar xzf matter2mqtt-v0.1.0-linux-arm64.tar.gz
cd matter2mqtt
./chip-tool version
```

## Building from Source

**Warning:** Building chip-tool requires ~25GB of disk space and 30+ minutes.

### Prerequisites

**macOS:**
```bash
brew install openssl pkg-config cmake ninja
```

**Linux (Ubuntu/Debian):**
```bash
sudo apt-get update
sudo apt-get install -y git gcc g++ pkg-config libssl-dev libdbus-1-dev \
    libglib2.0-dev libavahi-client-dev ninja-build python3-venv \
    python3-dev python3-pip unzip libgirepository1.0-dev libcairo2-dev \
    libreadline-dev
```

### Build Steps

**1. Clone the repository**
```bash
git clone --depth 1 https://github.com/project-chip/connectedhomeip.git
cd connectedhomeip
```

**2. Initialize submodules**
```bash
# This downloads ~20GB+ of dependencies
git submodule update --init --recursive

# Or use their script for a shallow clone
./scripts/checkout_submodules.py --shallow --platform darwin  # or linux
```

**3. Activate the build environment**
```bash
source scripts/activate.sh

# This sets up a Python virtual environment and installs build tools
# First time takes several minutes
```

**4. Build chip-tool**

**For macOS ARM64:**
```bash
./scripts/build/build_examples.py --target darwin-arm64-chip-tool build
# Binary will be at: out/darwin-arm64-chip-tool/chip-tool
```

**For macOS x86_64:**
```bash
./scripts/build/build_examples.py --target darwin-x64-chip-tool build
# Binary will be at: out/darwin-x64-chip-tool/chip-tool
```

**For Linux ARM64:**
```bash
./scripts/build/build_examples.py --target linux-arm64-chip-tool build
# Binary will be at: out/linux-arm64-chip-tool/chip-tool
```

**For Linux x86_64:**
```bash
./scripts/build/build_examples.py --target linux-x64-chip-tool build
# Binary will be at: out/linux-x64-chip-tool/chip-tool
```

**5. Install the binary**
```bash
# System-wide
sudo cp out/darwin-arm64-chip-tool/chip-tool /usr/local/bin/

# Or in your matter2mqtt project
mkdir -p ~/path/to/matter2mqtt/bin
cp out/darwin-arm64-chip-tool/chip-tool ~/path/to/matter2mqtt/bin/
```

**6. Verify installation**
```bash
chip-tool version
# Should show version information
```

## Build Time Expectations

- **First time setup:** 10-15 minutes (submodules + activate.sh)
- **Compilation:** 1-2 minutes on modern hardware
- **Total disk usage:** ~25GB for source + build artifacts
- **Final binary size:** ~50MB

## Troubleshooting

**Submodule errors:**
```bash
# If submodules fail, try again with force
git submodule update --init --recursive --force
```

**Python environment issues:**
```bash
# Deactivate and reactivate
deactivate  # if in venv
source scripts/activate.sh
```

**Build fails with missing dependencies:**
```bash
# Ensure all prerequisites installed
# Check connectedhomeip docs for platform-specific requirements
```

**Out of disk space:**
- Need at least 30GB free for build
- Consider building in Docker and extracting just the binary

## Docker Build (Alternative)

Build without installing dependencies on your system:

```bash
docker run -it --rm \
  -v $(pwd)/bin:/output \
  ubuntu:22.04 bash -c '
    apt-get update && \
    apt-get install -y git gcc g++ pkg-config libssl-dev libdbus-1-dev \
      libglib2.0-dev libavahi-client-dev ninja-build python3-venv \
      python3-dev python3-pip unzip libgirepository1.0-dev && \
    git clone --depth 1 https://github.com/project-chip/connectedhomeip.git && \
    cd connectedhomeip && \
    git submodule update --init --recursive && \
    source scripts/activate.sh && \
    ./scripts/build/build_examples.py --target linux-x64-chip-tool build && \
    cp out/linux-x64-chip-tool/chip-tool /output/
  '
```

This produces a Linux binary without affecting your host system.

## Cleanup

After building, you can safely delete the source tree to reclaim disk space:

```bash
# Keep only the binary
cp out/darwin-arm64-chip-tool/chip-tool ~/bin/
cd ..
rm -rf connectedhomeip  # Removes ~25GB
```

The chip-tool binary is self-contained and doesn't need the source tree to run.

## Version Management

Track which version of chip-tool you're using:

```bash
chip-tool version
# Or check the git commit it was built from
```

For production use, tag your chip-tool binaries with the connectedhomeip commit hash for reproducibility.
```