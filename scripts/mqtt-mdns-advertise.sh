#!/bin/bash
#
# MQTT mDNS Advertisement for macOS Development
#
# This script advertises the MQTT broker (running in Docker) via mDNS
# so ESP32 devices on your LAN can auto-discover it.
#
# Usage:
#   ./mqtt-mdns-advertise.sh        # Run in foreground
#   ./mqtt-mdns-advertise.sh &      # Run in background
#

SERVICE_NAME="Assembly Line MQTT Broker"
SERVICE_TYPE="_mqtt._tcp"
PORT="1883"

echo "🚀 Starting mDNS advertisement for MQTT broker"
echo "   Service: $SERVICE_NAME"
echo "   Type: $SERVICE_TYPE"
echo "   Port: $PORT"
echo ""
echo "ℹ️  This bridges Docker to your physical LAN for ESP32 discovery"
echo "   Press Ctrl+C to stop, or close this terminal to keep running"
echo ""

# Advertise using macOS's built-in dns-sd
dns-sd -R "$SERVICE_NAME" "$SERVICE_TYPE" . "$PORT"
