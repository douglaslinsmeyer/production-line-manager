#pragma once

// Device Configuration
// Device ID is MAC address (set dynamically at runtime)
#define DEVICE_TYPE "ESP32-S3-POE-8DI8DO"
#define FIRMWARE_VERSION "1.0.0"

// MQTT Configuration
#define MQTT_BROKER "192.168.68.123"  // Host machine IP where MQTT broker is running
#define MQTT_PORT 1883
#define MQTT_USER ""  // No auth by default
#define MQTT_PASSWORD ""

// MQTT Topics - Device Discovery Architecture
#define MQTT_TOPIC_ANNOUNCE "devices/announce"
#define MQTT_TOPIC_DEVICE_PREFIX "devices/"
#define MQTT_TOPIC_COMMAND_SUFFIX "/command"
#define MQTT_TOPIC_STATUS_SUFFIX "/status"
#define MQTT_TOPIC_INPUT_SUFFIX "/input-change"

// Legacy topics (for backward compatibility during migration)
#define MQTT_TOPIC_LEGACY_COMMAND "production-lines/commands/status"
#define MQTT_TOPIC_LEGACY_EVENT "production-lines/events/status"

// Network Configuration
#define USE_DHCP true
#define STATIC_IP "192.168.1.100"
#define GATEWAY "192.168.1.1"
#define SUBNET "255.255.255.0"
#define DNS_SERVER "8.8.8.8"

// Timing Configuration
#define HEARTBEAT_INTERVAL 30000  // 30 seconds
#define DEBOUNCE_DELAY 50         // 50ms debounce for inputs
#define BOOT_STABILIZATION_DELAY 100  // 100ms wait after boot for glitches to settle
#define INPUT_READY_DELAY 50      // Hardware stabilization before first read
                                  // Note: INPUT_GRACE_PERIOD (2s) in digital_input.cpp suppresses callbacks

// WiFi Configuration
#define WIFI_AP_CHANNEL 6                 // WiFi channel for AP mode
#define WIFI_AP_MAX_CONNECTIONS 4         // Max clients in AP mode
#define WIFI_CONNECTION_TIMEOUT 30000     // 30 seconds to connect
#define WIFI_RECONNECT_MAX_ATTEMPTS 10    // Max reconnection attempts before entering AP mode
#define WIFI_RECONNECT_INITIAL_DELAY 5000 // 5 seconds initial delay
#define WIFI_RECONNECT_MAX_DELAY 60000    // 60 seconds max delay

// Boot Button Configuration
#define BOOT_BUTTON_PIN 0                 // GPIO0 is the BOOT button on ESP32-S3
#define BOOT_BUTTON_LONG_PRESS 15000      // 15 seconds for long press (AP mode trigger)
#define BOOT_BUTTON_WARNING_TIME 10000    // 10 seconds warning (beep)

// Control Button Configuration
#define CONTROL_BUTTON_CHANNEL 0           // DIN1 (first digital input, 0-indexed)
#define CONTROL_BUTTON_GPIO 4              // GPIO4 (DIN1)
#define CONTROL_BUTTON_LONG_PRESS 5000     // 5 seconds for long press (maintenance mode)
#define BUTTON_LED_CHANNEL 4               // EXIO5 (TCA9554PWR channel 4, 0-indexed)

// Tower Light Configuration (Stack Light D01/D02/D03)
#define TOWER_LIGHT_RED_CHANNEL    0       // D01 (DOUT1) - Red
#define TOWER_LIGHT_YELLOW_CHANNEL 1       // D02 (DOUT2) - Yellow
#define TOWER_LIGHT_GREEN_CHANNEL  2       // D03 (DOUT3) - Green

// Status LED Configuration (Network/MQTT indicator on DO4)
#define STATUS_LED_CHANNEL 3               // D04 (DOUT4, Channel 3, 0-indexed)

// Button LED Pattern Timing
#define BUTTON_LED_MAINTENANCE_PERIOD 500  // 500ms blink for maintenance
#define BUTTON_LED_ERROR_PERIOD 200        // 200ms fast blink for error

// MQTT Buffer Configuration
#define MQTT_MAX_PACKET_SIZE 512          // Maximum MQTT packet size

// mDNS Configuration
#define MDNS_ENABLED true                 // Enabled by default
#define MDNS_SERVICE_NAME "_mqtt"         // Standard MQTT service name
#define MDNS_PROTOCOL "_tcp"              // TCP protocol
#define MDNS_DISCOVERY_TIMEOUT 5000       // 5 second discovery timeout
#define MDNS_CACHE_ENABLED true           // Cache discovered brokers
#define MDNS_CACHE_EXPIRY 3600000         // 1 hour cache validity (milliseconds)

// Display Configuration
#define DISPLAY_I2C_ADDRESS 0x3C          // SSD1306 I2C address (0x3C or 0x3D)
#define DISPLAY_WIDTH 128                 // OLED display width in pixels
#define DISPLAY_HEIGHT 64                 // OLED display height in pixels
#define DISPLAY_REFRESH_INTERVAL 2000     // Display refresh interval (2s)

// Hardware Configuration (from platformio.ini build_flags)
// Pin definitions are in build_flags - no need to redefine here
