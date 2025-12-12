# ESP32-S3-POE-ETH Firmware

PlatformIO-based firmware for Waveshare ESP32-S3-POE-ETH-8DI-8DO board with W5500 Ethernet.

## Hardware

- **Board**: Waveshare ESP32-S3-POE-ETH-8DI-8DO
- **MCU**: ESP32-S3 (Xtensa LX7 dual-core @ 240MHz)
- **Flash**: 16MB
- **PSRAM**: 8MB Octal
- **Ethernet**: W5500 (POE 802.3af)
- **WiFi**: 2.4GHz 802.11 b/g/n (Station + AP modes)
- **Digital Inputs**: 8 channels (GPIO4-11) with optocoupler isolation
- **Digital Outputs**: 8 channels via TCA9554PWR I2C expander

## Network Connectivity

The firmware supports three connection modes:

1. **Ethernet (Primary)**: W5500 with POE support, auto-connects on boot
2. **WiFi Station**: Connect to existing WiFi network, credentials stored in NVS
3. **WiFi AP Mode**: Hosts configuration portal at http://192.168.4.1

See [Network Configuration Documentation](docs/network-configuration.md) for details.

## Quick Start

### 1. Configure MQTT Broker IP

Edit `src/config.h` and update:
```cpp
#define MQTT_BROKER "192.168.68.123"  // Your MQTT broker IP
```

### 2. Build Firmware

```bash
cd /Users/douglasl/Projects/assembly-line-manager/firmware
~/.platformio/penv/bin/pio run
```

### 3. Flash to ESP32-S3 (First Time)

**Enter Download Mode**:
1. Connect USB-C cable to the board
2. Hold BOOT button (pulls GPIO0 LOW)
3. Press and release RESET button
4. Release BOOT button

**Flash**:
```bash
~/.platformio/penv/bin/pio run --target upload
```

**Monitor**:
```bash
~/.platformio/penv/bin/pio device monitor --baud 115200
```

### 4. Expected Boot Sequence

```
==============================================
  Waveshare ESP32-S3-POE-ETH-8DI-8DO
  Firmware Version: 1.0.0
  Device ID: esp32-s3-001
==============================================

Waiting for boot stabilization...
Boot stabilization complete

Initializing I2C...
  I2C SDA: GPIO42 (MTMS - JTAG pin)
  I2C SCL: GPIO41 (MTDI - JTAG pin)
  Note: Hardware JTAG debugging not available
  Use USB Serial/JTAG on GPIO19/20 for debugging

Initializing digital outputs...
TCA9554PWR initialized - all outputs OFF
✓ Digital outputs ready (all OFF)

Initializing digital inputs...
WARNING: Waiting for boot stabilization due to ESP32-S3 power-up glitches
Digital inputs initialized (GPIO4-11)
Digital inputs ready - boot stabilization complete
✓ Digital inputs configured

PSRAM Size: 8388608 bytes
Free PSRAM: ...

Initializing W5500 Ethernet...
  W5500 CS: GPIO16
  W5500 RST: GPIO39
  SPI SCK: GPIO15, MISO: GPIO14, MOSI: GPIO13
W5500 initialized - waiting for connection...
ETH Started
ETH Cable Connected
ETH Got IP: 192.168.1.xxx
  Gateway: 192.168.1.1
  Subnet: 255.255.255.0
  MAC: XX:XX:XX:XX:XX:XX

✓ Ethernet connection established
   IP Address: 192.168.1.xxx

Initializing MQTT client...
MQTT configured: broker=192.168.68.123:1883 client_id=esp32-s3-poe-eth-001
Connecting to MQTT broker...
MQTT connected!
Subscribed to: production-lines/commands/status
```

## Pin Mappings

### W5500 Ethernet (SPI)
- SCK: GPIO15
- MISO: GPIO14
- MOSI: GPIO13
- CS: GPIO16
- IRQ: GPIO12
- RST: GPIO39

### Digital Inputs
- DIN1-8: GPIO4, GPIO5, GPIO6, GPIO7, GPIO8, GPIO9, GPIO10, GPIO11

### Digital Outputs (I2C)
- I2C SCL: GPIO41 (MTDI)
- I2C SDA: GPIO42 (MTMS)
- TCA9554PWR Address: 0x20

### Other
- RGB LED: GPIO38
- Buzzer: GPIO46

## MQTT Integration

### Topics

**Subscribe** (commands from API):
- `production-lines/commands/status`

**Publish** (events to API):
- `production-lines/events/status`
- `production-lines/events/heartbeat`

### Command Format

```json
{
  "line_code": "LINE-001",
  "command": "set_output",
  "channel": 0,
  "state": true
}
```

### Status Event Format

```json
{
  "device_id": "esp32-s3-001",
  "line_code": "LINE-001",
  "ethernet_connected": true,
  "ip_address": "192.168.1.100",
  "digital_inputs": 15,
  "digital_outputs": 160,
  "uptime_seconds": 3600,
  "firmware_version": "1.0.0"
}
```

## Testing

### Test Digital Outputs

Use MQTTX Web Client (http://192.168.68.123:8090):

```json
// Turn ON output channel 1 (0-indexed)
Topic: production-lines/commands/status
Payload: {"line_code": "LINE-001", "command": "set_output", "channel": 0, "state": true}

// Turn OFF output channel 1
Payload: {"line_code": "LINE-001", "command": "set_output", "channel": 0, "state": false}
```

### Monitor Status Events

Subscribe to: `production-lines/events/status`

You should see heartbeat messages every 30 seconds with current input/output states.

## Troubleshooting

### Firmware Won't Flash
- Ensure you entered Download Boot mode (GPIO0=LOW)
- Try: Hold BOOT, press RESET, release BOOT
- Check USB cable connection

### No Serial Output
- Baud rate should be 115200
- ESP32-S3 uses USB CDC - may need driver on some systems
- Try different USB port

### Ethernet Not Connecting
- Check POE power or use 7-36V screw terminal
- Verify Ethernet cable is connected
- Check router DHCP settings
- W5500 reset happens 220ms after boot (100ms stabilization + 20ms reset + 100ms init)

### I2C Errors (TCA9554PWR not found)
- Check I2C connections
- Run I2C scanner to verify address 0x20
- GPIO41/42 should not have external pull-ups (conflicts with JTAG)

### MQTT Connection Fails
- Verify MQTT broker IP in config.h
- Check broker is running: `docker ps | grep mqtt`
- Test broker: `mosquitto_sub -h 192.168.68.123 -t "#" -v`

## Development

- **Platform**: PlatformIO
- **Framework**: Arduino ESP32 3.0.2 (with W5500 support)
- **Language**: C++17
- **Build Time**: ~130 seconds

## Critical Notes (ESP32-S3 Specific)

1. **Power-Up Glitches**: GPIO4-11 have 60µs low-level glitches at boot - firmware waits 100ms before reading inputs
2. **Strapping Pins**: GPIO0 (boot mode), GPIO3 (JTAG source), GPIO45/46 (VDD_SPI) - see ESP32-S3 datasheet
3. **JTAG Disabled**: GPIO41/42 used for I2C - use USB Serial/JTAG on GPIO19/20 for debugging
4. **Watchdogs**: RWDT and MWDT0 auto-enabled - firmware feeds with `delay(10)` in loop

## License

Proprietary
