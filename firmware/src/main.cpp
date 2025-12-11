#include <Arduino.h>
#include "config.h"
#include "device_config.h"
#include "network/connection_manager.h"
#include "gpio/boot_button.h"
#include "gpio/digital_input.h"
#include "gpio/digital_output.h"
#include "mqtt/mqtt_client.h"
#include "identification.h"

// Global managers
ConnectionManager networkManager;
BootButton bootButton;
DigitalInputManager inputs;
DigitalOutputManager outputs;
MQTTClientManager mqtt;
DeviceIdentification deviceID;

// Device identification (MAC address)
char deviceMAC[18];  // Format: "XX:XX:XX:XX:XX:XX"

// State tracking
unsigned long lastHeartbeat = 0;
unsigned long lastAnnouncement = 0;

void onInputChange(uint8_t channel, bool state);
void onNetworkConnection(bool connected);
void onMQTTCommand(const char* command, uint8_t channel, bool state);
void onFlashIdentify();
void onBootButtonLongPress(uint32_t duration);
String getMACAddress();

void setup() {
    // ===================================================================
    // STEP 1: Initialize Serial Communication (USB CDC)
    // ===================================================================
    Serial.begin(115200);
    delay(1000);  // Wait for USB enumeration

    Serial.println("\n\n==============================================");
    Serial.println("  Waveshare ESP32-S3-POE-ETH-8DI-8DO");
    Serial.printf("  Firmware Version: %s\n", FIRMWARE_VERSION);
    Serial.printf("  Device Type: %s\n", DEVICE_TYPE);
    Serial.println("==============================================\n");

    // ===================================================================
    // STEP 2: Wait for Boot Stabilization
    // CRITICAL: ESP32-S3 power-up glitches on GPIO1-20 (60µs low-level)
    // ===================================================================
    Serial.println("Waiting for boot stabilization...");
    delay(BOOT_STABILIZATION_DELAY);
    Serial.println("Boot stabilization complete\n");

    // ===================================================================
    // STEP 3: Get MAC Address for Device Identification
    // ===================================================================
    String macStr = getMACAddress();
    macStr.toCharArray(deviceMAC, sizeof(deviceMAC));
    Serial.printf("Device ID (MAC): %s\n\n", deviceMAC);

    // ===================================================================
    // STEP 4: Initialize Boot Button Handler
    // ===================================================================
    Serial.println("Initializing boot button handler...");
    bootButton.begin();
    bootButton.setLongPressCallback(onBootButtonLongPress);

    // Check for long press during boot (15s hold to enter AP mode)
    Serial.println("Checking for AP mode trigger (hold BOOT for 15s)...");
    unsigned long bootCheckStart = millis();
    while (millis() - bootCheckStart < 15100) {  // Check for 15.1 seconds
        bootButton.update();
        if (bootButton.longPressDetected()) {
            Serial.println("\n!!! BOOT BUTTON HELD 15 SECONDS !!!");
            Serial.println("Entering AP Mode - Clearing WiFi credentials");

            // Clear WiFi credentials and force AP mode
            deviceConfig.clearWiFiCredentials();
            deviceConfig.save();

            // Visual feedback
            Serial.println("Configuration cleared. Device will enter AP mode on boot.\n");

            bootButton.resetLongPress();
            break;
        }
        delay(10);
    }
    Serial.println("Boot button check complete\n");

    // ===================================================================
    // STEP 5: Load Device Configuration from NVS
    // ===================================================================
    Serial.println("Loading device configuration from NVS...");
    deviceConfig.begin();
    deviceConfig.printSettings();

    // ===================================================================
    // STEP 6: Initialize I2C for TCA9554PWR
    // Note: GPIO41/42 are JTAG pins - hardware JTAG will be disabled
    // ===================================================================
    Serial.println("Initializing I2C...");
    Serial.printf("  I2C SDA: GPIO%d (MTMS - JTAG pin)\n", I2C_SDA_PIN);
    Serial.printf("  I2C SCL: GPIO%d (MTDI - JTAG pin)\n", I2C_SCL_PIN);
    Serial.println("  Note: Hardware JTAG debugging not available");
    Serial.println("  Use USB Serial/JTAG on GPIO19/20 for debugging\n");

    // ===================================================================
    // STEP 7: Initialize Digital Outputs
    // ===================================================================
    Serial.println("Initializing digital outputs...");
    if (outputs.begin()) {
        Serial.println("✓ Digital outputs ready (all OFF)\n");
    } else {
        Serial.println("✗ ERROR: Digital outputs initialization FAILED\n");
    }

    // ===================================================================
    // STEP 8: Initialize Digital Inputs
    // Note: Will wait additional 50ms before reading to avoid glitches
    // ===================================================================
    Serial.println("Initializing digital inputs...");
    inputs.begin();
    inputs.setCallback(onInputChange);
    Serial.println("✓ Digital inputs configured\n");

    // ===================================================================
    // STEP 9: Initialize Device Identification (LED + Buzzer)
    // ===================================================================
    Serial.println("Initializing device identification...");
    deviceID.begin();
    Serial.println("✓ Device identification ready\n");

    // ===================================================================
    // STEP 10: Display PSRAM Info
    // ===================================================================
    Serial.printf("PSRAM Size: %d bytes\n", ESP.getPsramSize());
    Serial.printf("Free PSRAM: %d bytes\n\n", ESP.getFreePsram());

    // ===================================================================
    // STEP 11: Initialize Network (WiFi OR Ethernet)
    // Unified network manager handles interface selection
    // ===================================================================
    Serial.println("Initializing network...");
    if (networkManager.begin(deviceMAC)) {
        networkManager.setConnectionCallback(onNetworkConnection);

        // Wait for connection (30 second timeout)
        Serial.println("Waiting for network connection (30s timeout)...");
        unsigned long timeout = millis() + 30000;
        while (!networkManager.isConnected() && millis() < timeout) {
            delay(100);
            networkManager.update();
            bootButton.update();  // Continue monitoring boot button
        }

        if (networkManager.isConnected()) {
            Serial.println("✓ Network connected!");
            Serial.printf("   Interface: %s\n",
                         networkManager.getActiveInterface() == ConnectionManager::INTERFACE_WIFI ? "WiFi" : "Ethernet");
            Serial.printf("   IP Address: %s\n", networkManager.getIP().toString().c_str());

            if (networkManager.getActiveInterface() == ConnectionManager::INTERFACE_WIFI) {
                Serial.printf("   RSSI: %d dBm\n", networkManager.getRSSI());

                // Check if in AP mode
                WiFiManager* wifi = networkManager.getWiFiManager();
                if (wifi && wifi->getMode() == WiFiManager::MODE_AP) {
                    Serial.println("   MODE: Access Point (setup mode)");
                    Serial.println("   Connect to device's WiFi network to configure");
                }
            }
            Serial.println();
        } else {
            Serial.println("✗ Network connection timeout");

            // If WiFi mode, check if in AP mode
            if (deviceConfig.getConnectionMode() == MODE_WIFI) {
                WiFiManager* wifi = networkManager.getWiFiManager();
                if (wifi && wifi->getMode() == WiFiManager::MODE_AP) {
                    Serial.println("✓ Access Point mode active");
                    Serial.printf("   AP SSID: ESP32-Setup-XXXXXX\n");
                    Serial.printf("   AP IP: %s\n", networkManager.getIP().toString().c_str());
                    Serial.println("   Connect to configure WiFi\n");
                }
            }
        }
    } else {
        Serial.println("✗ ERROR: Network initialization FAILED\n");
    }

    // ===================================================================
    // STEP 12: Initialize MQTT with Device Discovery
    // ===================================================================
    Serial.println("Initializing MQTT client...");
    mqtt.begin(deviceMAC);  // Use MAC address as device ID
    mqtt.setCommandCallback(onMQTTCommand);
    mqtt.setFlashCallback(onFlashIdentify);

    if (networkManager.isConnected()) {
        mqtt.connect();
    }

    Serial.println("\n==============================================");
    Serial.println("  Initialization Complete");
    Serial.println("==============================================\n");

    // Print system info
    Serial.printf("Chip Model: %s\n", ESP.getChipModel());
    Serial.printf("Chip Revision: %d\n", ESP.getChipRevision());
    Serial.printf("CPU Frequency: %d MHz\n", ESP.getCpuFreqMHz());
    Serial.printf("Flash Size: %d bytes\n", ESP.getFlashChipSize());
    Serial.printf("Free Heap: %d bytes\n\n", ESP.getFreeHeap());
}

void loop() {
    // ===================================================================
    // Main Loop - runs continuously
    // ===================================================================

    // Update network manager (WiFi or Ethernet)
    networkManager.update();

    // Update boot button handler
    bootButton.update();

    // Handle long press during runtime (force AP mode)
    if (bootButton.longPressDetected()) {
        Serial.println("\n!!! BOOT BUTTON HELD - ENTERING AP MODE !!!");
        deviceConfig.clearWiFiCredentials();
        deviceConfig.save();

        Serial.println("Rebooting to AP mode in 3 seconds...");
        delay(3000);
        ESP.restart();
    }

    // Update MQTT client (handles reconnection)
    mqtt.update();

    // Update digital inputs (debouncing + change detection)
    inputs.update();

    // Periodic device announcement (every 60 seconds)
    if (millis() - lastAnnouncement > 60000) {
        lastAnnouncement = millis();

        if (mqtt.isConnected()) {
            mqtt.publishAnnouncement();
        }
    }

    // Periodic status/heartbeat (every 30 seconds)
    if (millis() - lastHeartbeat > HEARTBEAT_INTERVAL) {
        lastHeartbeat = millis();

        if (mqtt.isConnected()) {
            mqtt.publishStatus(
                inputs.getAllInputs(),
                outputs.getAllOutputs(),
                networkManager.isConnected()
            );
        }
    }

    // Feed watchdog timer
    // ESP32-S3 has auto-enabled watchdogs (RWDT and MWDT0)
    delay(10);
}

// ===================================================================
// Callback Functions
// ===================================================================

void onInputChange(uint8_t channel, bool state) {
    // Publish input change to MQTT
    if (mqtt.isConnected()) {
        mqtt.publishInputChange(channel, state, inputs.getAllInputs());
    }
}

void onFlashIdentify() {
    Serial.println("\n========================================");
    Serial.println("  FLASH IDENTIFY TRIGGERED");
    Serial.println("========================================\n");

    // Flash device for 10 seconds
    deviceID.flashIdentify(10);
}

String getMACAddress() {
    // Get MAC address from ESP32 eFuse
    uint64_t mac64 = ESP.getEfuseMac();

    // ESP.getEfuseMac() returns MAC in reverse byte order
    uint8_t mac[6];
    mac[0] = (mac64 >> 40) & 0xFF;
    mac[1] = (mac64 >> 32) & 0xFF;
    mac[2] = (mac64 >> 24) & 0xFF;
    mac[3] = (mac64 >> 16) & 0xFF;
    mac[4] = (mac64 >> 8) & 0xFF;
    mac[5] = mac64 & 0xFF;

    char macStr[18];
    snprintf(macStr, sizeof(macStr),
             "%02X:%02X:%02X:%02X:%02X:%02X",
             mac[0], mac[1], mac[2], mac[3], mac[4], mac[5]);

    return String(macStr);
}

void onNetworkConnection(bool connected) {
    if (connected) {
        Serial.println("\n✓ Network connection established");
        Serial.printf("   Interface: %s\n",
                     networkManager.getActiveInterface() == ConnectionManager::INTERFACE_WIFI ? "WiFi" : "Ethernet");
        Serial.printf("   IP Address: %s\n", networkManager.getIP().toString().c_str());

        if (networkManager.getActiveInterface() == ConnectionManager::INTERFACE_WIFI) {
            Serial.printf("   RSSI: %d dBm\n", networkManager.getRSSI());
        }

        // Connect to MQTT when network comes up
        mqtt.connect();
    } else {
        Serial.println("\n✗ Network connection lost");

        // Disconnect MQTT when network goes down
        mqtt.disconnect();
    }
}

void onBootButtonLongPress(uint32_t duration) {
    Serial.printf("\n=== BOOT BUTTON LONG PRESS DETECTED (%lu ms) ===\n", duration);
    Serial.println("AP mode reset will be triggered");

    // Visual/audio feedback will be added when integrated with DeviceIdentification
}

void onMQTTCommand(const char* command, uint8_t channel, bool state) {
    Serial.printf("Executing command: %s (CH%d=%s)\n",
                 command, channel + 1, state ? "ON" : "OFF");

    // Handle set_output command
    if (strcmp(command, "set_output") == 0) {
        if (outputs.setOutput(channel, state)) {
            Serial.printf("✓ Output CH%d set to %s\n", channel + 1, state ? "ON" : "OFF");

            // Publish updated status
            mqtt.publishStatus(
                inputs.getAllInputs(),
                outputs.getAllOutputs(),
                networkManager.isConnected()
            );
        } else {
            Serial.printf("✗ Failed to set output CH%d\n", channel + 1);
        }
    }
    // Add more commands here as needed
    else {
        Serial.printf("Unknown command: %s\n", command);
    }
}
