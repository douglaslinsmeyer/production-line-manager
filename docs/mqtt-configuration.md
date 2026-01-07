# MQTT Configuration Guide

## Overview

This project uses RabbitMQ as the MQTT broker for communication between ESP32-S3 devices, the Go API, and potentially web clients. RabbitMQ 3.13 provides robust MQTT support with multiple protocol versions.

## Protocol Versions

RabbitMQ 3.13 supports **MQTT 3.1, 3.1.1, and 5.0** simultaneously. Clients can connect using any supported version.

### Current Client Implementations

- **ESP32-S3 Firmware**: MQTT 3.1.1 (using PubSubClient library)
- **Go API**: MQTT with QoS 1 (version negotiated automatically)
- **Web UI**: Not currently using MQTT (uses HTTP REST API only)

### MQTT 5.0 Support and Limitations

RabbitMQ 3.13 introduced MQTT 5.0 support, but with some limitations:

#### ✅ Supported MQTT 5.0 Features

- Session expiry intervals (configured: 86400 seconds / 24 hours)
- Retained messages
- QoS 0 and QoS 1 message delivery
- Will messages
- Clean session / Clean start
- Topic aliases
- User properties
- Message expiry intervals

#### ❌ Unsupported MQTT 5.0 Features

- **QoS 2 (Exactly-once delivery)**: If an MQTT 5.0 client publishes with QoS 2, RabbitMQ will disconnect the client with reason code 155: "QoS not supported"
- **Shared subscriptions**: Not implemented in RabbitMQ's MQTT plugin
- **AUTH packet**: Re-authentication during an active session is not supported
- **Some subscription options**: Certain advanced subscription options may not be available

#### Impact on This Project

- ✅ Current firmware uses QoS 0/1 → No issues
- ✅ API publishes with QoS 1 → No issues
- ✅ No shared subscriptions required → No issues
- ⚠️ **Important**: If upgrading to MQTT 5.0 client libraries, avoid using QoS 2

## Connection Information

### MQTT TCP (for embedded devices/firmware)

**Development Environment:**
```
Host: 192.168.68.123 (or Docker host IP)
Port: 1883
Protocol: MQTT 3.1.1 or MQTT 5.0
Auth: None (anonymous enabled for development)
TLS: Disabled
```

**Example firmware configuration:**
```cpp
#define MQTT_BROKER "192.168.68.123"
#define MQTT_PORT 1883
#define MQTT_USER ""  // No auth in development
#define MQTT_PASSWORD ""
```

### MQTT WebSocket (for web browsers)

**Development Environment:**
```
URL: ws://localhost:9001/
Protocol: ws:// (unencrypted)
MQTT Version: 3.1.1 or 5.0
Auth: None (anonymous enabled for development)
Compression: Enabled
```

**⚠️ Non-Standard Configuration:**
- This project uses **port 9001** instead of RabbitMQ's default **15675**
- This project uses **path /** instead of RabbitMQ's default **/ws**
- Standard connection would be: `ws://localhost:15675/ws`

**Why the custom configuration?**
The custom port/path configuration works perfectly but deviates from RabbitMQ standards. When implementing WebSocket clients, use `ws://localhost:9001/` not the standard `ws://localhost:15675/ws`.

### WebSocket Usage Example (JavaScript)

If you want to add real-time MQTT connections to the web UI:

1. **Install MQTT.js library:**
   ```bash
   cd web
   npm install mqtt
   ```

2. **Connect to broker:**
   ```typescript
   import mqtt from 'mqtt';

   // MQTT 5.0 connection
   const client = mqtt.connect('ws://localhost:9001/', {
     protocol: 'ws',
     protocolVersion: 5,  // MQTT 5.0
     clean: true,
     reconnectPeriod: 1000,
   });

   // Or MQTT 3.1.1 connection
   const client = mqtt.connect('ws://localhost:9001/', {
     protocol: 'ws',
     protocolVersion: 4,  // MQTT 3.1.1
     clean: true,
     reconnectPeriod: 1000,
   });

   client.on('connect', () => {
     console.log('Connected to MQTT broker');
     client.subscribe('production-lines/#', { qos: 1 });
   });

   client.on('message', (topic, message) => {
     console.log('Received:', topic, message.toString());
   });

   // Publish message
   client.publish('production-lines/commands/status',
     JSON.stringify({ command: 'get_status' }),
     { qos: 1 }
   );
   ```

3. **Update docker-compose environment:**
   ```yaml
   web:
     environment:
       VITE_MQTT_WS_URL: ws://localhost:9001/
   ```

## RabbitMQ Management UI

Access the RabbitMQ management interface at:
```
URL: http://localhost:15672
Username: guest
Password: guest
```

The management UI allows you to:
- Monitor MQTT connections in real-time
- View message rates and queue statistics
- Inspect topics and subscriptions
- Debug connection issues

## Topic Architecture

### Device Topics

Devices use a hierarchical topic structure:

```
devices/announce                    # Device announces presence
devices/{device_id}/command         # Commands sent to device
devices/{device_id}/status          # Device status updates
devices/{device_id}/input-change    # Input state changes
```

### Production Line Topics (Legacy)

Legacy topics for backward compatibility:

```
production-lines/commands/status    # System commands
production-lines/events/status      # System events
```

## Quality of Service (QoS) Levels

### QoS 0 - At most once
- Fastest delivery, no guarantees
- Message may be lost if connection drops
- Use for: Non-critical status updates, high-frequency sensor data

### QoS 1 - At least once
- Message guaranteed to arrive at least once
- May receive duplicates
- **Currently used by this project** for commands and status
- Use for: Commands, important events, device control

### QoS 2 - Exactly once
- **NOT SUPPORTED by RabbitMQ MQTT plugin**
- Attempting to use will disconnect client
- **Do not use in this project**

## Configuration Files

### rabbitmq.conf

Location: `/rabbitmq.conf`

Key settings:
```conf
# MQTT TCP
mqtt.listeners.tcp.default = 1883
mqtt.allow_anonymous = true  # DEV ONLY
mqtt.max_session_expiry_interval_seconds = 86400
mqtt.prefetch = 10

# MQTT WebSocket
web_mqtt.tcp.port = 9001
web_mqtt.ws_path = /
web_mqtt.ws_opts.compress = true
```

### docker-compose.yml

MQTT service configuration:
```yaml
mqtt:
  image: rabbitmq:3.13-management-alpine
  ports:
    - "1883:1883"      # MQTT TCP
    - "9001:9001"      # MQTT WebSocket
    - "15672:15672"    # Management UI
```

## Debugging and Testing

### Using MQTTX Web Client

The project includes MQTTX Web for testing MQTT connections:

```bash
# Access MQTTX at:
http://localhost:8090

# Connection settings:
Host: localhost (or 192.168.68.123)
Port: 9001
Protocol: ws:// or 1883 for TCP
MQTT Version: 5.0 or 3.1.1
```

### Command Line Testing

Test MQTT connection using mosquitto_pub/mosquitto_sub:

```bash
# Subscribe to all topics
mosquitto_sub -h localhost -p 1883 -t '#' -v

# Publish test message
mosquitto_pub -h localhost -p 1883 -t 'test/topic' -m 'Hello MQTT'

# Subscribe to device topics
mosquitto_sub -h localhost -p 1883 -t 'devices/#' -v
```

### Check RabbitMQ Status

```bash
# Check RabbitMQ version (should show 3.13.x)
docker exec production-line-mqtt rabbitmqctl version

# List enabled plugins (should show rabbitmq_mqtt and rabbitmq_web_mqtt)
docker exec production-line-mqtt rabbitmq-plugins list

# Show MQTT connections
docker exec production-line-mqtt rabbitmqctl list_connections

# Show current queues
docker exec production-line-mqtt rabbitmqctl list_queues
```

## Production Deployment Considerations

⚠️ **The current configuration is FOR DEVELOPMENT ONLY**

### Critical Security Requirements for Production

1. **Disable Anonymous Access**
   ```conf
   mqtt.allow_anonymous = false
   # Remove mqtt.default_user and mqtt.default_pass
   ```

2. **Create Dedicated MQTT Users**
   ```bash
   # Create users with strong passwords
   rabbitmqctl add_user firmware_device <strong-password>
   rabbitmqctl add_user web_ui_client <strong-password>
   rabbitmqctl add_user api_client <strong-password>

   # Set minimal permissions
   rabbitmqctl set_permissions -p / firmware_device "devices/.*" "devices/.*" "devices/.*|production-lines/.*"
   rabbitmqctl set_permissions -p / web_ui_client "production-lines/.*" "" "production-lines/.*"
   rabbitmqctl set_permissions -p / api_client "production-lines/.*" "production-lines/.*" "production-lines/.*"
   ```

3. **Enable TLS/SSL Encryption**
   ```conf
   # MQTT over TLS (port 8883)
   mqtt.listeners.ssl.default = 8883
   ssl_options.cacertfile = /etc/rabbitmq/certs/ca-cert.pem
   ssl_options.certfile = /etc/rabbitmq/certs/server-cert.pem
   ssl_options.keyfile = /etc/rabbitmq/certs/server-key.pem
   ssl_options.verify = verify_peer

   # WebSocket over SSL (port 15676)
   web_mqtt.ssl.port = 15676
   web_mqtt.ssl.cacertfile = /etc/rabbitmq/certs/ca-cert.pem
   web_mqtt.ssl.certfile = /etc/rabbitmq/certs/server-cert.pem
   web_mqtt.ssl.keyfile = /etc/rabbitmq/certs/server-key.pem
   ```

4. **Change Admin Credentials**
   ```yaml
   environment:
     RABBITMQ_DEFAULT_USER: admin
     RABBITMQ_DEFAULT_PASS: <strong-generated-password>
   ```

5. **Restrict Management UI**
   ```conf
   # Only allow management UI from localhost (use SSH tunnel)
   management.tcp.ip = 127.0.0.1
   ```

6. **Disable Management Metrics Collector** (for high connection count)
   ```conf
   management_agent.disable_metrics_collector = true
   ```

7. **Enable Persistent Logging**
   ```conf
   log.file = /var/log/rabbitmq/rabbitmq.log
   log.file.level = info
   log.file.rotation.date = $D0  # Daily rotation
   ```

8. **Use Internal Docker Network**
   - Don't expose MQTT ports directly to the internet
   - Use reverse proxy/load balancer with TLS termination
   - Enable proxy protocol if behind ALB/NLB

## Performance Optimization

### For High Connection Count

If you have hundreds or thousands of concurrent MQTT connections:

1. **Increase Erlang process limit:**
   ```yaml
   environment:
     RABBITMQ_SERVER_ADDITIONAL_ERL_ARGS: "+P 1048576"
   ```

2. **Disable management metrics:**
   ```conf
   management_agent.disable_metrics_collector = true
   ```

3. **Use Prometheus for monitoring instead of Management UI**

4. **Optimize TCP buffers:**
   ```conf
   mqtt.tcp_listen_options.sndbuf = 32768
   mqtt.tcp_listen_options.recbuf = 32768
   ```

### QoS Performance

- **QoS 0 significantly outperforms QoS 1** (up to 4x faster)
- Use QoS 0 for high-frequency sensor data
- Use QoS 1 only for critical commands/events

## Troubleshooting

### Connection Refused

- Check if RabbitMQ is running: `docker ps | grep mqtt`
- Check port bindings: `docker port production-line-mqtt`
- Check firewall rules
- Verify IP address in firmware config matches Docker host

### Client Disconnects Immediately

- Check if using QoS 2 (not supported - will disconnect)
- Verify MQTT version compatibility
- Check credentials if anonymous access disabled
- Review RabbitMQ logs: `docker logs production-line-mqtt`

### Messages Not Received

- Verify topic subscription patterns match published topics
- Check QoS levels on both publisher and subscriber
- Verify client connected before subscribing
- Use MQTTX to debug topic flow

### High Memory Usage

- Check `vm_memory_high_watermark` setting
- Reduce `mqtt.prefetch` value
- Verify retained messages aren't accumulating
- Check for zombie connections

## References

- [RabbitMQ MQTT Plugin Documentation](https://www.rabbitmq.com/docs/mqtt)
- [RabbitMQ Web MQTT Plugin Documentation](https://www.rabbitmq.com/docs/web-mqtt)
- [MQTT 5.0 in RabbitMQ 3.13 Blog Post](https://www.rabbitmq.com/blog/2023/07/21/mqtt5)
- [RabbitMQ 3.13 Release Notes](https://www.rabbitmq.com/blog/2024/03/11/rabbitmq-3.13.0-announcement)
- [MQTT.js Documentation](https://github.com/mqttjs/MQTT.js)
- [PubSubClient Library (Arduino)](https://github.com/knolleary/pubsubclient)
