# MQTT Message Formats

This document defines the JSON schemas for all MQTT messages in the Assembly Line Manager system.

## Device Messages

### Device Announcement

**Topic**: `devices/announce`

Devices publish this message every 60 seconds for discovery and heartbeat.

```json
{
  "device_id": "A4:D3:22:A0:ED:30",
  "device_type": "ESP32-S3-POE-8DI8DO",
  "firmware_version": "1.0.0",
  "ip_address": "10.100.131.76",
  "mac_address": "A4:D3:22:A0:ED:30",
  "capabilities": {
    "digital_inputs": 8,
    "digital_outputs": 8,
    "ethernet": true,
    "wifi": true
  },
  "status": {
    "uptime_seconds": 3600,
    "free_heap": 327868
  }
}
```

**Fields**:
- `device_id` (string): Unique device identifier (typically MAC address)
- `device_type` (string): Hardware model
- `firmware_version` (string): Firmware version (semver format)
- `ip_address` (string): Current IP address
- `mac_address` (string): MAC address
- `capabilities` (object): Device capabilities
  - `digital_inputs` (number): Number of digital inputs
  - `digital_outputs` (number): Number of digital outputs
  - `ethernet` (boolean): Ethernet support
  - `wifi` (boolean): WiFi support
- `status` (object): Current status
  - `uptime_seconds` (number): Time since boot
  - `free_heap` (number): Available memory in bytes

### Device Status Update

**Topic**: `devices/{MAC}/status`

Regular status updates published every 30 seconds.

```json
{
  "uptime_seconds": 7200,
  "free_heap": 325432,
  "wifi_rssi": -45,
  "connected": true
}
```

**Fields**:
- `uptime_seconds` (number): Time since boot
- `free_heap` (number): Available heap memory
- `wifi_rssi` (number, optional): WiFi signal strength (dBm)
- `connected` (boolean): Network connection status

### Input Change Event

**Topic**: `devices/{MAC}/input-change`

Published when a digital input state changes.

```json
{
  "channel": 0,
  "state": true,
  "timestamp": 1734567890
}
```

**Fields**:
- `channel` (number): Input channel (0-7)
- `state` (boolean): New input state (true=high, false=low)
- `timestamp` (number): Unix timestamp of change

## Device Commands

**Topic**: `devices/{MAC}/command`

Commands sent from API to devices.

### Flash Identify Command

Triggers LED and buzzer for physical device identification.

```json
{
  "command": "flash_identify",
  "duration": 10
}
```

**Fields**:
- `command` (string): "flash_identify"
- `duration` (number): Duration in seconds (default: 10)

### Set Output Command

Sets the state of a digital output channel.

```json
{
  "command": "set_output",
  "channel": 0,
  "state": true
}
```

**Fields**:
- `command` (string): "set_output"
- `channel` (number): Output channel (0-7)
- `state` (boolean): Desired state (true=on, false=off)

### Configure Device Command

Updates device configuration (WiFi, MQTT, etc.).

```json
{
  "command": "configure",
  "config": {
    "wifi_ssid": "FactoryWiFi",
    "wifi_password": "password123",
    "mqtt_broker": "192.168.1.100",
    "mqtt_port": 1883,
    "device_name": "Production Line 1 Controller"
  }
}
```

**Fields**:
- `command` (string): "configure"
- `config` (object): Configuration parameters
  - `wifi_ssid` (string, optional): WiFi SSID
  - `wifi_password` (string, optional): WiFi password
  - `mqtt_broker` (string, optional): MQTT broker address
  - `mqtt_port` (number, optional): MQTT broker port
  - `device_name` (string, optional): Friendly device name

### Reboot Command

Restarts the device.

```json
{
  "command": "reboot"
}
```

**Fields**:
- `command` (string): "reboot"

### Factory Reset Command

Resets device to factory defaults.

```json
{
  "command": "factory_reset"
}
```

**Fields**:
- `command` (string): "factory_reset"

## Production Line Messages

### Line Created Event

**Topic**: `production-lines/events/created`

Published when a new production line is created.

```json
{
  "line_id": "550e8400-e29b-41d4-a716-446655440000",
  "line_code": "LINE-001",
  "name": "Assembly Line 1",
  "description": "Main assembly line",
  "status": "off",
  "created_at": "2025-12-11T10:30:00Z"
}
```

**Fields**:
- `line_id` (string): UUID of the line
- `line_code` (string): Human-readable code
- `name` (string): Line name
- `description` (string): Line description
- `status` (string): Initial status (on/off/maintenance/error)
- `created_at` (string): ISO 8601 timestamp

### Line Updated Event

**Topic**: `production-lines/events/updated`

Published when a production line is modified.

```json
{
  "line_id": "550e8400-e29b-41d4-a716-446655440000",
  "line_code": "LINE-001",
  "name": "Assembly Line 1 - Updated",
  "description": "Main assembly line (updated)",
  "status": "on",
  "updated_at": "2025-12-11T11:00:00Z"
}
```

**Fields**: Same as created event, plus `updated_at` timestamp

### Line Deleted Event

**Topic**: `production-lines/events/deleted`

Published when a production line is soft-deleted.

```json
{
  "line_id": "550e8400-e29b-41d4-a716-446655440000",
  "line_code": "LINE-001",
  "deleted_at": "2025-12-11T12:00:00Z"
}
```

**Fields**:
- `line_id` (string): UUID of the deleted line
- `line_code` (string): Line code
- `deleted_at` (string): ISO 8601 timestamp

### Line Status Changed Event

**Topic**: `production-lines/events/status`

Published when a production line status changes.

```json
{
  "line_id": "550e8400-e29b-41d4-a716-446655440000",
  "line_code": "LINE-001",
  "status": "maintenance",
  "previous_status": "on",
  "timestamp": "2025-12-11T13:00:00Z",
  "source": "api",
  "reason": "Scheduled maintenance"
}
```

**Fields**:
- `line_id` (string): UUID of the line
- `line_code` (string): Line code
- `status` (string): New status (on/off/maintenance/error)
- `previous_status` (string): Previous status
- `timestamp` (string): ISO 8601 timestamp
- `source` (string): Change source (api/device/schedule)
- `reason` (string, optional): Reason for change

## Production Line Commands

### Status Change Command

**Topic**: `production-lines/commands/status`

Command to change a production line status (from shop floor controllers or devices).

```json
{
  "line_code": "LINE-001",
  "status": "error",
  "source": "device",
  "reason": "Motor overheating detected"
}
```

**Fields**:
- `line_code` (string): Line code to update
- `status` (string): Desired status (on/off/maintenance/error)
- `source` (string): Command source (device/controller/manual)
- `reason` (string, optional): Reason for change

## Status Values

Production line status can be one of:
- `on`: Line is running normally
- `off`: Line is stopped
- `maintenance`: Line is under maintenance
- `error`: Line has encountered an error

## Timestamp Format

All timestamps use ISO 8601 format with UTC timezone:
```
2025-12-11T10:30:00Z
```

## JSON Schema Notes

- All JSON messages must be valid UTF-8
- Boolean values are `true` or `false` (lowercase)
- Numbers are JSON numbers (no quotes)
- Strings use double quotes
- Optional fields may be omitted
- Unknown fields are ignored

## Error Handling

If a device receives an invalid command:
1. Device logs error to serial output
2. Device may publish error message to `devices/{MAC}/error` topic
3. Command is not executed

If API receives an invalid message:
1. API logs error with message details
2. Message is acknowledged and discarded
3. No action is taken

## See Also

- [MQTT Topics Reference](./topics.md) - Topic structure and patterns
- [Architecture Overview](../architecture.md) - System architecture
- [API Documentation](../../api/README.md) - API endpoints
- [Firmware Documentation](../../firmware/README.md) - Device implementation
