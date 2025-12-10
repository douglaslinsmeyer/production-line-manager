#include <Arduino.h>
#include "config.h"
#include "ethernet/eth_manager.h"
#include "gpio/digital_input.h"
#include "gpio/digital_output.h"
#include "mqtt/mqtt_client.h"
#include "identification.h"

// Global managers
EthernetManager ethernet;
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
void onEthernetConnection(bool connected);
void onMQTTCommand(const char* command, uint8_t channel, bool state);
void onFlashIdentify();
String getMACAddress();

void setup() {
    // ===================================================================
    // STEP 1: Initialize Serial Communication (USB CDC)
    // ===================================================================
    Serial.begin(115200);
    delay(1000);  // Wait for USB enumeration

    // Get MAC address for device identification
    String macStr = getMACAddress();
    macStr.toCharArray(deviceMAC, sizeof(deviceMAC));

    Serial.println("\n\n==============================================");
    Serial.println("  Waveshare ESP32-S3-POE-ETH-8DI-8DO");
    Serial.printf("  Firmware Version: %s\n", FIRMWARE_VERSION);
    Serial.printf("  Device ID (MAC): %s\n", deviceMAC);
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
    // STEP 3: Initialize I2C for TCA9554PWR
    // Note: GPIO41/42 are JTAG pins - hardware JTAG will be disabled
    // ===================================================================
    Serial.println("Initializing I2C...");
    Serial.printf("  I2C SDA: GPIO%d (MTMS - JTAG pin)\n", I2C_SDA_PIN);
    Serial.printf("  I2C SCL: GPIO%d (MTDI - JTAG pin)\n", I2C_SCL_PIN);
    Serial.println("  Note: Hardware JTAG debugging not available");
    Serial.println("  Use USB Serial/JTAG on GPIO19/20 for debugging\n");

    // ===================================================================
    // STEP 4: Initialize Digital Outputs
    // ===================================================================
    Serial.println("Initializing digital outputs...");
    if (outputs.begin()) {
        Serial.println("✓ Digital outputs ready (all OFF)\n");
    } else {
        Serial.println("✗ ERROR: Digital outputs initialization FAILED\n");
    }

    // ===================================================================
    // STEP 5: Initialize Digital Inputs
    // Note: Will wait additional 50ms before reading to avoid glitches
    // ===================================================================
    Serial.println("Initializing digital inputs...");
    inputs.begin();
    inputs.setCallback(onInputChange);
    Serial.println("✓ Digital inputs configured\n");

    // ===================================================================
    // STEP 6: Initialize Device Identification (LED + Buzzer)
    // ===================================================================
    Serial.println("Initializing device identification...");
    deviceID.begin();
    Serial.println("✓ Device identification ready\n");

    // ===================================================================
    // STEP 7: Display PSRAM Info
    // ===================================================================
    Serial.printf("PSRAM Size: %d bytes\n", ESP.getPsramSize());
    Serial.printf("Free PSRAM: %d bytes\n\n", ESP.getFreePsram());

    // ===================================================================
    // STEP 8: Initialize Ethernet
    // Includes proper W5500 reset sequence accounting for glitches
    // ===================================================================
    if (ethernet.begin()) {
        ethernet.setConnectionCallback(onEthernetConnection);

        // Wait for Ethernet connection (30 second timeout)
        Serial.println("Waiting for Ethernet connection (30s timeout)...");
        unsigned long timeout = millis() + 30000;
        while (!ethernet.isConnected() && millis() < timeout) {
            delay(100);
            ethernet.update();
        }

        if (ethernet.isConnected()) {
            Serial.println("✓ Ethernet connected!\n");
        } else {
            Serial.println("✗ Ethernet connection timeout\n");
        }
    } else {
        Serial.println("✗ ERROR: Ethernet initialization FAILED\n");
    }

    // ===================================================================
    // STEP 9: Initialize MQTT with Device Discovery
    // ===================================================================
    Serial.println("Initializing MQTT client...");
    mqtt.begin(deviceMAC);  // Use MAC address as device ID
    mqtt.setCommandCallback(onMQTTCommand);
    mqtt.setFlashCallback(onFlashIdentify);

    if (ethernet.isConnected()) {
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

    // Update Ethernet manager
    ethernet.update();

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
                ethernet.isConnected()
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

void onEthernetConnection(bool connected) {
    if (connected) {
        Serial.println("\n✓ Ethernet connection established");
        Serial.printf("   IP Address: %s\n", ethernet.getIP().toString().c_str());

        // Connect to MQTT when Ethernet comes up
        mqtt.connect();
    } else {
        Serial.println("\n✗ Ethernet connection lost");

        // Disconnect MQTT when Ethernet goes down
        mqtt.disconnect();
    }
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
                ethernet.isConnected()
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
