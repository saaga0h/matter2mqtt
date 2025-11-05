#!/bin/bash
# IPv6 Verification Script for matter2mqtt
# This script checks if your system is properly configured for Matter/Thread

set -e

echo "======================================"
echo "Matter/Thread IPv6 Verification"
echo "======================================"
echo

# Color codes
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Check functions
check_pass() {
    echo -e "${GREEN}✓${NC} $1"
}

check_fail() {
    echo -e "${RED}✗${NC} $1"
}

check_warn() {
    echo -e "${YELLOW}⚠${NC} $1"
}

# 1. Check if IPv6 is enabled in kernel
echo "1. Checking IPv6 kernel support..."
if [ -f /proc/net/if_inet6 ]; then
    check_pass "IPv6 is enabled in kernel"
else
    check_fail "IPv6 is NOT enabled in kernel"
    echo "   Fix: Enable IPv6 in kernel or boot parameters"
    exit 1
fi
echo

# 2. Check sysctl IPv6 settings
echo "2. Checking sysctl IPv6 configuration..."
ipv6_disabled=$(cat /proc/sys/net/ipv6/conf/all/disable_ipv6 2>/dev/null || echo "1")
if [ "$ipv6_disabled" = "0" ]; then
    check_pass "IPv6 is enabled (disable_ipv6=0)"
else
    check_fail "IPv6 is disabled (disable_ipv6=1)"
    echo "   Fix: sudo sysctl -w net.ipv6.conf.all.disable_ipv6=0"
    exit 1
fi
echo

# 3. Check for IPv6 addresses
echo "3. Checking for IPv6 addresses..."
ipv6_addrs=$(ip -6 addr show scope global 2>/dev/null | grep "inet6" | wc -l)
ipv6_link_local=$(ip -6 addr show scope link 2>/dev/null | grep "inet6" | wc -l)

if [ "$ipv6_addrs" -gt 0 ]; then
    check_pass "Found $ipv6_addrs global IPv6 address(es)"
elif [ "$ipv6_link_local" -gt 0 ]; then
    check_warn "No global IPv6, but found $ipv6_link_local link-local address(es)"
    echo "   Note: Link-local is sufficient for Matter commissioning"
else
    check_fail "No IPv6 addresses found"
    echo "   This may prevent Matter commissioning"
fi

echo "   IPv6 addresses:"
ip -6 addr show | grep "inet6" | sed 's/^/   /'
echo

# 4. Check IPv6 forwarding (required for OTBR)
echo "4. Checking IPv6 forwarding..."
ipv6_forward=$(cat /proc/sys/net/ipv6/conf/all/forwarding 2>/dev/null || echo "0")
if [ "$ipv6_forward" = "1" ]; then
    check_pass "IPv6 forwarding is enabled"
else
    check_warn "IPv6 forwarding is disabled"
    echo "   OTBR may enable this automatically"
    echo "   Manual fix: sudo sysctl -w net.ipv6.conf.all.forwarding=1"
fi
echo

# 5. Check for IPv6 default route
echo "5. Checking IPv6 routing..."
ipv6_routes=$(ip -6 route show 2>/dev/null | wc -l)
if [ "$ipv6_routes" -gt 0 ]; then
    check_pass "IPv6 routing table has $ipv6_routes route(s)"
    echo "   Routes:"
    ip -6 route show | head -n 5 | sed 's/^/   /'
else
    check_warn "No IPv6 routes found"
    echo "   Note: OTBR will create Thread routes"
fi
echo

# 6. Check Docker IPv6 support
echo "6. Checking Docker IPv6 configuration..."
if command -v docker &> /dev/null; then
    if docker network inspect bridge | grep -q "EnableIPv6.*true"; then
        check_pass "Docker bridge network has IPv6 enabled"
    else
        check_warn "Docker bridge network does NOT have IPv6 enabled"
        echo "   Note: matter2mqtt uses host networking, so this is OK"
        echo "   Only affects containers using bridge networking"
    fi
else
    check_warn "Docker not found or not running"
fi
echo

# 7. Check for Avahi/mDNS (required for Matter discovery)
echo "7. Checking mDNS/Avahi..."
if systemctl is-active --quiet avahi-daemon 2>/dev/null; then
    check_pass "Avahi daemon is running"
elif command -v avahi-daemon &> /dev/null; then
    check_warn "Avahi is installed but not running"
    echo "   Start with: sudo systemctl start avahi-daemon"
else
    check_warn "Avahi not found"
    echo "   Install with: sudo apt install avahi-daemon (Debian/Ubuntu)"
    echo "   Note: OTBR container includes its own Avahi"
fi
echo

# 8. Test IPv6 connectivity
echo "8. Testing IPv6 connectivity..."
if ping6 -c 1 -W 2 2001:4860:4860::8888 &>/dev/null; then
    check_pass "Can ping Google DNS (2001:4860:4860::8888)"
elif ping6 -c 1 -W 2 ::1 &>/dev/null; then
    check_warn "IPv6 loopback works, but no external connectivity"
    echo "   Note: External IPv6 not required for Matter"
else
    check_fail "IPv6 ping failed"
fi
echo

# 9. Check USB devices
echo "9. Checking for USB serial devices..."
usb_devices=$(ls -1 /dev/serial/by-id/ 2>/dev/null | wc -l)
if [ "$usb_devices" -gt 0 ]; then
    check_pass "Found $usb_devices USB serial device(s):"
    ls -1 /dev/serial/by-id/ | sed 's/^/   /'
else
    check_warn "No USB serial devices found in /dev/serial/by-id/"
    echo "   Checking /dev/tty*..."
    ls -1 /dev/tty{ACM,USB}* 2>/dev/null | sed 's/^/   /' || echo "   None found"
fi
echo

# Summary
echo "======================================"
echo "Summary"
echo "======================================"
echo
echo "Your system is ready for Matter/Thread if:"
echo "  ✓ IPv6 is enabled in kernel"
echo "  ✓ At least link-local IPv6 addresses exist"
echo "  ✓ USB device is detected"
echo
echo "The following are optional but recommended:"
echo "  • IPv6 forwarding (OTBR enables this)"
echo "  • Global IPv6 address (not required for local Matter)"
echo "  • Avahi running (OTBR container includes it)"
echo

# Exit code based on critical checks
if [ "$ipv6_disabled" = "0" ] && [ "$ipv6_link_local" -gt 0 ]; then
    echo -e "${GREEN}✓ Your system appears ready for Matter/Thread!${NC}"
    exit 0
else
    echo -e "${RED}✗ Please fix the issues above before running matter2mqtt${NC}"
    exit 1
fi
