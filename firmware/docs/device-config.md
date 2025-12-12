# Device Configuration (NVS)

## Overview

ESP32-S3 firmware stores configuration in Non-Volatile Storage (NVS), a flash-based key-value storage system. Configuration persists across reboots and power cycles, allowing devices to remember WiFi credentials, MQTT settings, and device metadata.

## NVS Namespace

**Namespace**: `device_config`

All device configuration is stored within this namespace to avoid conflicts with other firmware components or libraries.

## Configuration Parameters

### Network Settings

**WiFi SSID**
- **Key**: `wifi_ssid`
- **Type**: String (NVS_TYPE_STR)
- **Max Size**: 32 bytes
- **Default**: Empty (not configured)
- **Description**: WiFi network name to connect to

**WiFi Password**
- **Key**: `wifi_password`
- **Type**: String (NVS_TYPE_STR)
- **Max Size**: 64 bytes
- **Default**: Empty (not configured)
- **Description**: WiFi network password
- **Security**: Stored encrypted in NVS

**WiFi Enabled**
- **Key**: `wifi_enabled`
- **Type**: uint8 (NVS_TYPE_U8)
- **Values**: 0=disabled, 1=enabled
- **Default**: 1 (enabled)
- **Description**: Enable/disable WiFi functionality

**AP Password**
- **Key**: `ap_password`
- **Type**: String (NVS_TYPE_STR)
- **Max Size**: 64 bytes
- **Default**: "esp32config"
- **Description**: Password for Access Point mode

**Prefer Ethernet**
- **Key**: `prefer_ethernet`
- **Type**: uint8 (NVS_TYPE_U8)
- **Values**: 0=no, 1=yes
- **Default**: 1 (yes)
- **Description**: Prefer Ethernet over WiFi when both available

### MQTT Settings

**MQTT Broker**
- **Key**: `mqtt_broker`
- **Type**: String (NVS_TYPE_STR)
- **Max Size**: 128 bytes
- **Default**: "192.168.1.100" (from config.h)
- **Description**: MQTT broker IP address or hostname

**MQTT Port**
- **Key**: `mqtt_port`
- **Type**: uint16 (NVS_TYPE_U16)
- **Default**: 1883
- **Description**: MQTT broker TCP port

**MQTT Username**
- **Key**: `mqtt_username`
- **Type**: String (NVS_TYPE_STR)
- **Max Size**: 64 bytes
- **Default**: Empty (anonymous)
- **Description**: MQTT authentication username

**MQTT Password**
- **Key**: `mqtt_password`
- **Type**: String (NVS_TYPE_STR)
- **Max Size**: 64 bytes
- **Default**: Empty (anonymous)
- **Description**: MQTT authentication password
- **Security**: Stored encrypted in NVS

**MQTT Use TLS**
- **Key**: `mqtt_use_tls`
- **Type**: uint8 (NVS_TYPE_U8)
- **Values**: 0=no, 1=yes
- **Default**: 0 (no TLS)
- **Description**: Enable TLS/SSL for MQTT connection

### Device Identity

**Device Name**
- **Key**: `device_name`
- **Type**: String (NVS_TYPE_STR)
- **Max Size**: 64 bytes
- **Default**: "ESP32-S3-{MAC}"
- **Description**: Friendly device name

**Line Assignment**
- **Key**: `line_assignment`
- **Type**: String (NVS_TYPE_STR)
- **Max Size**: 64 bytes
- **Default**: Empty (unassigned)
- **Description**: Production line code this device is assigned to

## Default Values

Defaults are defined in `config.h`:

```cpp
#define DEFAULT_MQTT_BROKER "192.168.1.100"
#define DEFAULT_MQTT_PORT 1883
#define DEFAULT_AP_PASSWORD "esp32config"
#define DEFAULT_DEVICE_NAME "ESP32-S3-Device"
```

## Configuration Methods

### Method 1: Web Configuration Portal (Primary)

**When**: Device is in Access Point mode

**Steps**:
1. Connect to device AP: `ESP32-DEVICE-{MAC}`
2. Browse to http://192.168.4.1
3. Fill out configuration form:
   - WiFi SSID (scan and select)
   - WiFi Password
   - MQTT Broker IP
   - MQTT Port
   - Device Name (optional)
4. Click "Save and Reboot"
5. Device saves to NVS and restarts

**Behind the scenes**:
```cpp
// Web server receives POST request
nvs_handle_t handle;
nvs_open("device_config", NVS_READWRITE, &handle);
nvs_set_str(handle, "wifi_ssid", ssid);
nvs_set_str(handle, "wifi_password", password);
nvs_set_str(handle, "mqtt_broker", broker);
nvs_set_u16(handle, "mqtt_port", port);
nvs_commit(handle);
nvs_close(handle);

ESP.restart();  // Reboot to apply
```

### Method 2: MQTT Command (Remote)

**When**: Device is already connected to network and MQTT

**Topic**: `devices/{MAC}/command`

**Payload**:
```json
{
  "command": "configure",
  "config": {
    "wifi_ssid": "FactoryWiFi",
    "wifi_password": "newpassword123",
    "mqtt_broker": "10.0.0.50",
    "mqtt_port": 1883,
    "device_name": "Assembly Line 1 Controller"
  }
}
```

**Response**: Device updates NVS and reboots to apply changes

### Method 3: Factory Reset (Erase All)

**Trigger**: Hold BOOT button for 10 seconds

**Effect**:
```cpp
nvs_handle_t handle;
nvs_open("device_config", NVS_READWRITE, &handle);
nvs_erase_all(handle);  // Erase all keys in namespace
nvs_commit(handle);
nvs_close(handle);

ESP.restart();  // Reboot in AP mode with defaults
```

**Result**:
- All configuration erased
- Device reboots
- Starts in AP mode
- Uses default settings

## Configuration API (Firmware Code)

### Load Configuration

```cpp
class DeviceConfig {
public:
  String wifi_ssid;
  String wifi_password;
  String mqtt_broker;
  uint16_t mqtt_port;
  String device_name;

  bool load() {
    nvs_handle_t handle;
    if (nvs_open("device_config", NVS_READONLY, &handle) != ESP_OK) {
      return false;
    }

    // Load each setting
    char buf[128];
    size_t len;

    len = sizeof(buf);
    if (nvs_get_str(handle, "wifi_ssid", buf, &len) == ESP_OK) {
      wifi_ssid = String(buf);
    }

    len = sizeof(buf);
    if (nvs_get_str(handle, "mqtt_broker", buf, &len) == ESP_OK) {
      mqtt_broker = String(buf);
    } else {
      mqtt_broker = DEFAULT_MQTT_BROKER;
    }

    if (nvs_get_u16(handle, "mqtt_port", &mqtt_port) != ESP_OK) {
      mqtt_port = DEFAULT_MQTT_PORT;
    }

    nvs_close(handle);
    return true;
  }
};
```

### Save Configuration

```cpp
bool DeviceConfig::save() {
  nvs_handle_t handle;
  if (nvs_open("device_config", NVS_READWRITE, &handle) != ESP_OK) {
    return false;
  }

  nvs_set_str(handle, "wifi_ssid", wifi_ssid.c_str());
  nvs_set_str(handle, "wifi_password", wifi_password.c_str());
  nvs_set_str(handle, "mqtt_broker", mqtt_broker.c_str());
  nvs_set_u16(handle, "mqtt_port", mqtt_port);
  nvs_set_str(handle, "device_name", device_name.c_str());

  esp_err_t err = nvs_commit(handle);
  nvs_close(handle);

  return (err == ESP_OK);
}
```

### Factory Reset

```cpp
bool DeviceConfig::factoryReset() {
  nvs_handle_t handle;
  if (nvs_open("device_config", NVS_READWRITE, &handle) != ESP_OK) {
    return false;
  }

  esp_err_t err = nvs_erase_all(handle);
  nvs_commit(handle);
  nvs_close(handle);

  return (err == ESP_OK);
}
```

## Configuration Backup and Restore

### Export Configuration (via API)

**Endpoint**: `GET /api/v1/devices/{mac}/config`

**Response**:
```json
{
  "wifi_ssid": "FactoryWiFi",
  "mqtt_broker": "192.168.1.100",
  "mqtt_port": 1883,
  "device_name": "Assembly Line 1"
}
```

Note: Passwords are **not** included in export for security.

### Import Configuration (via API)

**Endpoint**: `POST /api/v1/devices/{mac}/config`

**Body**: Same as export format

**Effect**: Sends configure command via MQTT

## Security Considerations

### Credential Storage

**WiFi Password**:
- Stored in NVS flash
- NVS partition can be encrypted (ESP32-S3 feature)
- Not transmitted in clear text except during initial setup

**MQTT Password**:
- Stored in NVS flash
- Enable NVS encryption for production deployments
- Use TLS for MQTT connections when possible

### Encryption

Enable NVS encryption in `platformio.ini`:

```ini
board_build.partitions = partitions_encrypted.csv
board_build.f_flash = 80000000L
board_build.flash_mode = qio
```

Then initialize NVS encryption on first boot:

```cpp
esp_err_t ret = nvs_flash_secure_init();
if (ret == ESP_ERR_NVS_NO_FREE_PAGES || ret == ESP_ERR_NVS_NEW_VERSION_FOUND) {
  ESP_ERROR_CHECK(nvs_flash_erase());
  ret = nvs_flash_secure_init();
}
ESP_ERROR_CHECK(ret);
```

### Access Control

- Web portal protected by AP password
- MQTT commands require device MAC address (prevents broadcast attacks)
- Factory reset requires physical access (boot button)

## Troubleshooting

### Configuration Not Persisting

**Symptom**: Settings reset after reboot

**Checks**:
1. Verify `nvs_commit()` is called after writes
2. Check NVS partition is not corrupted: `nvs_flash_init()`
3. Ensure flash wear leveling is working
4. Check for out-of-space errors in NVS

**Solution**: Erase and reinitialize NVS
```cpp
nvs_flash_erase();
nvs_flash_init();
```

### Can't Read Configuration

**Symptom**: `nvs_get_str()` returns `ESP_ERR_NVS_NOT_FOUND`

**Cause**: Key doesn't exist (never set)

**Solution**: Use default values
```cpp
if (nvs_get_str(handle, "mqtt_broker", buf, &len) != ESP_OK) {
  // Key not found, use default
  mqtt_broker = DEFAULT_MQTT_BROKER;
}
```

### Factory Reset Not Working

**Symptom**: Configuration persists after factory reset

**Checks**:
1. Verify `nvs_erase_all()` is called on correct namespace
2. Check return value for errors
3. Ensure `nvs_commit()` is called after erase
4. Verify device actually reboots after erase

## See Also

- [Network Configuration](./network-configuration.md)
- [ESP32 NVS Documentation](https://docs.espressif.com/projects/esp-idf/en/latest/esp32/api-reference/storage/nvs_flash.html)
- [MQTT Topics Reference](../../docs/mqtt/topics.md)
- [API Device Configuration Endpoint](../../api/README.md#devices)
