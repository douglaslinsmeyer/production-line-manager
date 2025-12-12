# MQTT Topic Reference

## Broker: RabbitMQ 3.13

### Connection Details
- **MQTT TCP**: Port 1883
- **MQTT WebSocket**: Port 9001
- **Management UI**: http://localhost:15672
- **Default Credentials**: guest/guest
- **Protocol**: MQTT 3.1.1

### Configuration
RabbitMQ is configured with the MQTT plugin enabled. The broker handles both MQTT and AMQP protocols simultaneously.

## Topic Structure

### Device Topics

#### Published by Devices

**Device Announcement**
- **Topic**: `devices/announce`
- **QoS**: 1
- **Frequency**: Every 60 seconds
- **Purpose**: Device discovery and heartbeat
- **Payload**: Device information including MAC, IP, firmware version, capabilities

**Device Status**
- **Topic**: `devices/{MAC}/status`
- **QoS**: 1
- **Frequency**: Every 30 seconds
- **Purpose**: Regular status updates with uptime and memory info
- **Payload**: Device status metrics

**Input Change Events**
- **Topic**: `devices/{MAC}/input-change`
- **QoS**: 1
- **Frequency**: On change
- **Purpose**: Digital input state changes
- **Payload**: Channel number and new state

#### Subscribed by Devices

**Device Commands**
- **Topic**: `devices/{MAC}/command`
- **QoS**: 1
- **Purpose**: Receive commands from API
- **Commands**:
  - `flash_identify`: Blink LED and buzzer for identification
  - `set_output`: Set digital output state
  - `configure`: Update device configuration
  - `reboot`: Restart device
  - `factory_reset`: Reset to defaults

### Production Line Topics

#### Published by API

**Line Created Event**
- **Topic**: `production-lines/events/created`
- **QoS**: 1
- **Purpose**: Notify when new production line is created
- **Payload**: Full line details

**Line Updated Event**
- **Topic**: `production-lines/events/updated`
- **QoS**: 1
- **Purpose**: Notify when production line is modified
- **Payload**: Updated line details

**Line Deleted Event**
- **Topic**: `production-lines/events/deleted`
- **QoS**: 1
- **Purpose**: Notify when production line is soft-deleted
- **Payload**: Line ID and code

**Line Status Changed Event**
- **Topic**: `production-lines/events/status`
- **QoS**: 1
- **Purpose**: Notify when line status changes
- **Payload**: Line ID, code, new status, timestamp, source

#### Subscribed by API

**Status Change Command**
- **Topic**: `production-lines/commands/status`
- **QoS**: 1
- **Purpose**: Receive status change commands from shop floor controllers or devices
- **Payload**: Line code, new status, source

## Topic Patterns

### Device-Specific Topics
All device-specific topics use the MAC address as the unique identifier:
```
devices/{MAC}/*
```

Example for device `A4:D3:22:A0:ED:30`:
```
devices/A4:D3:22:A0:ED:30/status
devices/A4:D3:22:A0:ED:30/input-change
devices/A4:D3:22:A0:ED:30/command
```

### Wildcard Subscriptions
The API subscribes to multiple topics using wildcards:

- **All device announcements**: `devices/announce`
- **All device statuses**: `devices/+/status`
- **All input changes**: `devices/+/input-change`
- **Status commands**: `production-lines/commands/status`

## Quality of Service (QoS)

All topics use **QoS 1** (At least once delivery):
- Messages are delivered at least once
- Broker stores message until acknowledged
- Suitable for important state changes and commands
- Small overhead but ensures reliability

## Message Retention

- **Device Announcements**: Retained (latest announcement available to new subscribers)
- **Device Status**: Not retained (periodic updates, no need for history)
- **Line Events**: Not retained (state stored in database)
- **Commands**: Not retained (one-time execution)

## Usage Examples

### Subscribe to All Device Announcements
```bash
mosquitto_sub -h localhost -p 1883 -t "devices/announce" -v
```

### Subscribe to Specific Device Commands
```bash
mosquitto_sub -h localhost -p 1883 -t "devices/A4:D3:22:A0:ED:30/command" -v
```

### Publish Status Change Command
```bash
mosquitto_pub -h localhost -p 1883 -t "production-lines/commands/status" \
  -m '{"line_code":"LINE-001","status":"maintenance","source":"manual"}'
```

### Subscribe to All Line Events
```bash
mosquitto_sub -h localhost -p 1883 -t "production-lines/events/#" -v
```

## Testing with MQTTX

The system includes MQTTX Web Client for testing:
1. Open http://localhost:8090
2. Create new connection:
   - Host: `mqtt://localhost`
   - Port: `1883`
   - Username: `guest`
   - Password: `guest`
3. Subscribe to topics to monitor traffic
4. Publish test messages to trigger actions

## See Also

- [Message Formats](./message-formats.md) - Detailed payload schemas
- [Architecture Overview](../architecture.md) - System architecture
- [Device Discovery](../../firmware/docs/network-configuration.md) - Device announcement details
- [API MQTT Integration](../../api/README.md#mqtt-integration) - API pub/sub patterns
