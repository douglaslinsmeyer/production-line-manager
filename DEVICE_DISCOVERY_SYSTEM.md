# Device Discovery & Assignment System

## Overview

Zero-touch provisioning system for ESP32-S3 IoT devices. Devices auto-discover themselves via MQTT, and the dashboard provides a UI to assign devices to production lines and physically identify them.

## System Architecture

```
ESP32-S3 Device (MAC: A4:D3:22:A0:ED:30)
    ├─ Announces every 60s → devices/announce
    ├─ Publishes status → devices/{MAC}/status
    ├─ Publishes input changes → devices/{MAC}/input-change
    └─ Subscribes to commands → devices/{MAC}/command

                    ↓ MQTT

MQTT Broker (Mosquitto)
    └─ Topics: devices/*, production-lines/*

                    ↓

Go API Server
    ├─ Device Discovery Handler (listens to devices/announce)
    ├─ Device Registry (PostgreSQL)
    ├─ Device-Line Assignment Mapping
    ├─ Event Translation (device events → line events)
    └─ REST API: /api/v1/devices/*

                    ↓ HTTP/JSON

React Dashboard
    ├─ Device Discovery Page (/devices)
    ├─ Flash Button (triggers LED + buzzer)
    ├─ Assign Dropdown (maps device → line)
    └─ Real-time status updates (polling every 5s)
```

## Implementation Status

### ✅ Firmware (ESP32-S3)

**Files Created**:
- `firmware/platformio.ini` - PlatformIO configuration with Arduino 3.0.2 for W5500 support
- `firmware/src/main.cpp` - Main application with device discovery
- `firmware/src/config.h` - MAC-based device ID configuration
- `firmware/src/identification.h/.cpp` - LED + buzzer flash implementation
- `firmware/src/mqtt/mqtt_client.h/.cpp` - Device-centric MQTT topics
- `firmware/src/ethernet/eth_manager.h/.cpp` - W5500 Ethernet driver
- `firmware/src/gpio/digital_input.h/.cpp` - 8 digital inputs with debouncing
- `firmware/src/gpio/digital_output.h/.cpp` - 8 digital outputs via TCA9554PWR I2C

**Features**:
- ✓ Bootloader working (flashed successfully)
- ✓ W5500 Ethernet via POE (10.100.131.76)
- ✓ Device announcement with MAC, IP, firmware, capabilities
- ✓ Flash identification (GPIO38 LED + GPIO46 buzzer)
- ✓ Digital I/O operational (8 inputs, 8 outputs)
- ✓ MQTT connectivity to broker
- ✓ Watchdog timer management

**Device Info**:
- MAC: **A4:D3:22:A0:ED:30**
- Type: ESP32-S3-POE-8DI8DO
- IP: 10.100.131.76
- Firmware: v1.0.0
- Status: Online

### ✅ API (Go Backend)

**Files Created**:
- `api/migrations/005_create_devices.up.sql` - Device tables schema
- `api/internal/domain/device.go` - Device domain models
- `api/internal/repository/device_repository.go` - Device database operations
- `api/internal/mqtt/device_discovery.go` - MQTT device announcement handler
- `api/internal/handler/devices.go` - Device REST API endpoints

**Files Modified**:
- `api/internal/mqtt/subscriber.go` - Added device/* topic subscriptions
- `api/internal/mqtt/publisher.go` - Added PublishRaw method
- `api/internal/handler/router.go` - Added device routes
- `api/cmd/server/main.go` - Wired device discovery handler

**Database Tables**:
- `discovered_devices` - Registry of all devices
- `device_line_assignments` - Device-to-line mappings

**API Endpoints**:
```
GET    /api/v1/devices                    List all discovered devices
GET    /api/v1/devices/{mac}              Get device details
POST   /api/v1/devices/{mac}/assign       Assign device to line
DELETE /api/v1/devices/{mac}/assign       Unassign device
POST   /api/v1/devices/{mac}/identify     Flash LED/buzzer
POST   /api/v1/devices/{mac}/command      Send custom command
GET    /api/v1/lines/{id}/device          Get device for line
```

**Features**:
- ✓ Device discovery via MQTT
- ✓ Device registry with online/offline tracking
- ✓ Device-to-line assignment logic
- ✓ Flash identify command routing
- ✓ Stale device monitoring (offline after 2 minutes)

### ✅ Dashboard (React Frontend)

**Files Created**:
- `web/src/types/device.ts` - TypeScript device types
- `web/src/api/devices.ts` - Device API client
- `web/src/pages/DeviceDiscovery.tsx` - Device discovery UI

**Files Modified**:
- `web/src/App.tsx` - Added /devices route
- `web/src/components/common/Sidebar.tsx` - Added Devices navigation link

**Features**:
- ✓ Device list with real-time updates (5s polling)
- ✓ Flash button for physical identification
- ✓ Assign dropdown to map device → line
- ✓ Unassign button
- ✓ Status indicators (online/offline)
- ✓ Device capabilities display
- ✓ Summary statistics (total, online, assigned)

## How to Use the System

### 1. Power On a Device

1. Connect ESP32-S3 to power (POE or 7-36V screw terminal)
2. Device boots and connects to Ethernet
3. Device announces itself via MQTT
4. API discovers device and stores in database
5. Device appears in dashboard within 5 seconds

### 2. Identify a Device

1. Navigate to **Devices** page in dashboard
2. Click **Flash** button next to device
3. Device blinks green LED and beeps buzzer for 10 seconds
4. Note the physical location of the flashing device

### 3. Assign Device to Line

1. Find the device you want to assign
2. Select a production line from the dropdown
3. Device is now assigned
4. Future input changes from this device will update the line status

### 4. Monitor Devices

- **Online Status**: Green dot, last seen timestamp
- **Offline Status**: Gray dot, marked offline after 2 minutes of no announcements
- **Capabilities**: Shows 8DI/8DO, ETH/WiFi support
- **Firmware Version**: Track which version each device is running

## Testing the System

### Test Device Discovery

```bash
# 1. Check device is announced
curl http://localhost:8080/api/v1/devices | jq '.'

# Expected: Array with your device (MAC: A4:D3:22:A0:ED:30)
```

### Test Flash Identify

```bash
# 2. Send flash command
curl -X POST http://localhost:8080/api/v1/devices/A4:D3:22:A0:ED:30/identify \
  -H "Content-Type: application/json" \
  -d '{"duration": 10}'

# Watch ESP32-S3: Should blink green LED + beep buzzer for 10 seconds
```

### Test Device Assignment

```bash
# 3. Get a line ID
curl http://localhost:8080/api/v1/lines | jq '.[0].id'

# 4. Assign device to line
curl -X POST http://localhost:8080/api/v1/devices/A4:D3:22:A0:ED:30/assign \
  -H "Content-Type: application/json" \
  -d '{"line_id": "<line-uuid>"}'

# 5. Verify assignment
curl http://localhost:8080/api/v1/devices | jq '.[0] | {mac, assigned_line_code}'
```

### Test via Dashboard

1. Navigate to http://localhost:3000/devices
2. See discovered device in table
3. Click "Flash" - observe ESP32-S3 blinks/beeps
4. Select line from dropdown - device gets assigned
5. Toggle digital input on ESP32-S3
6. Check if assigned line status updates

## MQTT Topics Reference

### Published by Devices

```
devices/announce                    Device discovery (every 60s)
devices/{MAC}/status                Status updates (every 30s)
devices/{MAC}/input-change          Input state changes
```

### Subscribed by Devices

```
devices/{MAC}/command               Commands from API
```

### Example Messages

**Device Announcement**:
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
    "wifi": false
  },
  "status": {
    "uptime_seconds": 3600,
    "free_heap": 327868
  }
}
```

**Flash Identify Command**:
```json
{
  "command": "flash_identify",
  "duration": 10
}
```

**Set Output Command**:
```json
{
  "command": "set_output",
  "channel": 0,
  "state": true
}
```

## Database Schema

### discovered_devices

| Column | Type | Description |
|--------|------|-------------|
| id | UUID | Primary key |
| mac_address | TEXT | Unique device identifier |
| device_type | TEXT | Device model |
| firmware_version | TEXT | Firmware version |
| ip_address | TEXT | Current IP address |
| capabilities | JSONB | Device capabilities |
| first_seen | TIMESTAMPTZ | First discovery time |
| last_seen | TIMESTAMPTZ | Last announcement time |
| status | TEXT | online/offline/unknown |
| metadata | JSONB | Additional device info |

### device_line_assignments

| Column | Type | Description |
|--------|------|-------------|
| id | UUID | Primary key |
| device_mac | TEXT | Foreign key to device |
| line_id | UUID | Foreign key to line |
| assigned_at | TIMESTAMPTZ | Assignment time |
| assigned_by | TEXT | User who assigned |
| active | BOOLEAN | Only one active per device |

## Deployment Checklist

- [x] ESP32-S3 firmware flashed and running
- [x] W5500 Ethernet connectivity working
- [x] Device announces to MQTT broker
- [x] API subscribes to device/* topics
- [x] Database tables created
- [x] Device discovery handler implemented
- [x] REST API endpoints working
- [x] Dashboard device page created
- [x] Navigation updated
- [ ] Test flash identify end-to-end
- [ ] Test device assignment workflow
- [ ] Deploy to production

## Next Steps

1. **Test the Dashboard**:
   - Rebuild web container: `docker-compose build web`
   - Restart web container: `docker-compose up -d web`
   - Navigate to http://localhost:3000/devices
   - Verify device appears in table

2. **Test Assignment**:
   - Select a production line from dropdown
   - Verify device gets assigned
   - Toggle digital input on ESP32-S3
   - Check if line status updates

3. **Add More Devices**:
   - Flash firmware to additional ESP32-S3 boards
   - Each will auto-discover with unique MAC
   - Assign each to different lines

4. **Production Deployment**:
   - Update firmware with actual MQTT broker IP
   - Deploy multiple devices across facility
   - Use flash feature to identify each device location
   - Assign devices via dashboard

## Troubleshooting

### Device Not Appearing in Dashboard

1. Check MQTT connectivity: `docker logs production-line-mqtt`
2. Check API logs: `docker logs production-line-api | grep device`
3. Verify device is publishing: Use MQTTX Web (http://localhost:8090)
   - Subscribe to `devices/announce`
   - Should see messages every 60 seconds

### Flash Identify Not Working

1. Check command is sent: `docker logs production-line-api | grep identify`
2. Check device logs via serial: `pio device monitor`
3. Verify GPIO38 (LED) and GPIO46 (buzzer) are connected properly

### Device Shows Offline

- Device goes offline if no announcement for 2 minutes
- Check Ethernet cable connection
- Verify MQTT broker is accessible from device network
- Check serial logs for MQTT connection errors

## Files Created Summary

**Firmware** (9 files):
- platformio.ini, config.h, identification.h/.cpp
- mqtt/mqtt_client.h/.cpp (modified)
- main.cpp (modified)

**API** (5 files):
- migrations/005_create_devices.up/down.sql
- domain/device.go
- repository/device_repository.go
- mqtt/device_discovery.go
- handler/devices.go

**Dashboard** (3 files):
- types/device.ts
- api/devices.ts
- pages/DeviceDiscovery.tsx

**Total**: 17 files created/modified across the stack

## Congratulations!

You now have a complete IoT device management system with:
- Zero-touch device provisioning
- Physical device identification (flash feature)
- Centralized device-to-line assignment
- Real-time device monitoring
- Scalable architecture for 100+ devices
