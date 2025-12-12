# Network Configuration

## Overview

The ESP32-S3 firmware supports three network connection modes, providing flexibility for various deployment scenarios. The device prioritizes Ethernet for reliability but can fall back to WiFi or operate in Access Point mode for configuration.

## Connection Modes

### 1. Ethernet (Primary Mode)

**Hardware**: W5500 Ethernet controller with POE support (802.3af)

**Features**:
- Auto-connection on boot
- Reliable, wired network connection
- Power over Ethernet (POE) support
- Preferred mode for production environments

**Configuration**:
```cpp
// Pin mappings (in config.h)
#define W5500_CS_PIN   16
#define W5500_RST_PIN  39
#define W5500_IRQ_PIN  12
#define W5500_MOSI_PIN 11
#define W5500_MISO_PIN 13
#define W5500_SCLK_PIN 12
```

**Boot Sequence**:
1. Initialize W5500 SPI interface
2. Reset W5500 chip (220ms delay)
3. Obtain IP via DHCP
4. Connect to MQTT broker
5. Begin device announcement

**LED Indicator**: Solid Green

**Troubleshooting**:
- Ensure Ethernet cable is connected (link light on)
- Verify POE power or 7-36V DC supply
- Check router/switch POE compatibility
- Serial log shows "ETH Cable Not Connected" if no link

### 2. WiFi Station Mode

**Description**: Connect to an existing WiFi network

**Features**:
- Automatic connection on boot
- Credentials stored in NVS (persistent)
- Falls back to this mode if Ethernet unavailable
- Supports WPA/WPA2 security

**Configuration Storage**:
Credentials stored in NVS namespace `device_config`:
- `wifi_ssid` (string, max 32 bytes)
- `wifi_password` (string, max 64 bytes)
- `wifi_enabled` (uint8, 0=disabled, 1=enabled)

**Boot Sequence**:
1. Check if Ethernet connected
2. If not, read WiFi credentials from NVS
3. Connect to WiFi network
4. Obtain IP via DHCP
5. Connect to MQTT broker
6. Begin device announcement

**LED Indicators**:
- Blinking Green: Connecting to WiFi
- Solid Blue: WiFi connected

**Troubleshooting**:
- Verify SSID and password are correct (case-sensitive)
- Check WiFi signal strength (must be adequate)
- Ensure WiFi network is 2.4GHz (ESP32 doesn't support 5GHz)
- Avoid DFS channels (52-144)

### 3. WiFi Access Point Mode

**Description**: Device hosts its own WiFi network for configuration

**Features**:
- Triggered when no network available
- Hosts web configuration portal
- Captive portal (redirects all requests)
- Manual trigger via boot button (15 seconds)

**Default Configuration**:
- **SSID**: `ESP32-DEVICE-{MAC_LAST_4}`
- **Password**: Configured in NVS or default "esp32config"
- **IP Address**: 192.168.4.1
- **Web Portal**: http://192.168.4.1

**Trigger Conditions**:
1. No Ethernet cable connected
2. AND no WiFi credentials stored
3. OR WiFi connection failed (timeout)
4. OR boot button held for 15 seconds

**LED Indicator**: Blinking Blue

**Web Configuration Portal**:
- **URL**: http://192.168.4.1
- **Features**:
  - WiFi SSID scanner and selection
  - Password input
  - MQTT broker configuration
  - Device name and settings
  - Network status display
  - Reboot and factory reset options

**Timeout**: Returns to normal operation after 10 minutes of inactivity

**Exiting AP Mode**:
1. Configure WiFi credentials via web portal
2. Click "Save and Reboot"
3. Device restarts and connects to configured network

## Connection Priority

The firmware follows this priority when establishing network connection:

```
Boot
 ├─ Check Ethernet cable
 │   ├─ Connected? → Use Ethernet (DONE)
 │   └─ Not connected → Continue
 │
 ├─ Check WiFi credentials in NVS
 │   ├─ Found? → Try WiFi connection
 │   │   ├─ Success? → Use WiFi (DONE)
 │   │   └─ Failed? → Continue
 │   └─ Not found → Continue
 │
 └─ Start Access Point mode
     └─ Wait for configuration
```

## LED Status Indicators

| LED Color | Blink Pattern | Network Status |
|-----------|--------------|----------------|
| Solid Green | Always On | Ethernet connected |
| Blinking Green | 500ms on/off | WiFi connecting |
| Solid Blue | Always On | WiFi connected |
| Blinking Blue | 250ms on/off | AP mode active |
| Red | Always On | Network error |
| Off | - | No power or initialization |

## Network Configuration Methods

### Method 1: Web Portal (Primary)

1. Connect to device AP: `ESP32-DEVICE-{MAC}`
2. Browse to http://192.168.4.1
3. Select WiFi network from scanner
4. Enter password
5. Configure MQTT broker (optional)
6. Click "Save and Reboot"

### Method 2: MQTT Command (Remote)

Send configuration via MQTT when device is already connected:

**Topic**: `devices/{MAC}/command`

**Payload**:
```json
{
  "command": "configure",
  "config": {
    "wifi_ssid": "FactoryWiFi",
    "wifi_password": "password123",
    "mqtt_broker": "192.168.1.100",
    "mqtt_port": 1883
  }
}
```

### Method 3: Factory Reset

Hold BOOT button for 10 seconds:
1. All NVS configuration erased
2. Device reboots
3. Starts in AP mode with defaults

## Advanced Configuration

### Prefer Ethernet Over WiFi

Even if WiFi credentials are configured, Ethernet takes priority when available.

```cpp
// In device_config.h
#define PREFER_ETHERNET true
```

### Static IP (Optional)

By default, device uses DHCP. For static IP:

```cpp
// In eth_manager.cpp
IPAddress static_ip(192, 168, 1, 100);
IPAddress gateway(192, 168, 1, 1);
IPAddress subnet(255, 255, 255, 0);
IPAddress dns(8, 8, 8, 8);

Ethernet.begin(mac, static_ip, dns, gateway, subnet);
```

### WiFi Power Saving

Disable WiFi power saving for better reliability:

```cpp
WiFi.setSleep(false);  // Disable power saving
```

## Network Security Considerations

**Credential Storage**:
- WiFi passwords stored encrypted in NVS
- MQTT credentials stored encrypted in NVS
- Factory reset erases all credentials

**AP Mode Security**:
- AP password required (WPA2)
- 10-minute timeout to prevent unauthorized access
- Captive portal requires local network access

**MQTT Security**:
- Supports username/password authentication
- TLS/SSL support (optional, requires certificate)

## Troubleshooting Guide

### Device Not Connecting to WiFi

**Symptom**: Device stuck blinking green, never connects

**Checks**:
1. Verify SSID and password are correct
2. Check WiFi signal strength (should be > -70dBm)
3. Ensure 2.4GHz network (ESP32-S3 doesn't support 5GHz)
4. Avoid DFS channels
5. Factory reset and reconfigure

### Device Not Appearing in AP Mode

**Symptom**: Can't find `ESP32-DEVICE-{MAC}` network

**Checks**:
1. Wait 30 seconds after boot
2. Ensure no Ethernet cable connected
3. Verify no WiFi credentials configured (factory reset)
4. Hold boot button for 15 seconds to force AP mode
5. Check serial logs for errors

### Ethernet Not Working

**Symptom**: "ETH Cable Not Connected" in serial logs

**Checks**:
1. Verify Ethernet cable is connected
2. Check link light on RJ45 connector
3. Verify POE power or DC power supply
4. W5500 reset timing (220ms after boot, already configured)
5. Check SPI pin connections

### Frequent Disconnections

**Symptom**: Device connects then disconnects repeatedly

**Checks**:
1. WiFi signal strength (move closer to AP or use Ethernet)
2. Router/AP capacity (may be overloaded)
3. MQTT broker connectivity (separate issue)
4. Power supply stability (POE or DC must be stable)

## See Also

- [Device Configuration (NVS)](./device-config.md)
- [MQTT Topics Reference](../../docs/mqtt/topics.md)
- [MQTT Message Formats](../../docs/mqtt/message-formats.md)
- [Hardware Pinout](./Esp32-s3_datasheet_en.pdf)
