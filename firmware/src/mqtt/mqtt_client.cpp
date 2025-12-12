#include "mqtt_client.h"
#include "config.h"
#include "device_config.h"
#include "network/connection_manager.h"
#include <ETH.h>

// External references
extern DeviceConfig deviceConfig;
extern ConnectionManager networkManager;

// Static instance pointer for callback
MQTTClientManager* MQTTClientManager::instance = nullptr;

MQTTClientManager::MQTTClientManager()
    : mqttClient(ethClient),
      cmdCallback(nullptr),
      flashCallback(nullptr),
      networkManagerPtr(nullptr),
      lastReconnectAttempt(0),
      reconnectInterval(5000) {

    instance = this;
    deviceMAC[0] = '\0';
}

void MQTTClientManager::begin(const char* macAddress) {
    // Store MAC address
    strncpy(deviceMAC, macAddress, sizeof(deviceMAC) - 1);
    deviceMAC[sizeof(deviceMAC) - 1] = '\0';

    // Build device-specific topics
    snprintf(deviceTopicCommand, sizeof(deviceTopicCommand),
             "%s%s%s", MQTT_TOPIC_DEVICE_PREFIX, deviceMAC, MQTT_TOPIC_COMMAND_SUFFIX);
    snprintf(deviceTopicStatus, sizeof(deviceTopicStatus),
             "%s%s%s", MQTT_TOPIC_DEVICE_PREFIX, deviceMAC, MQTT_TOPIC_STATUS_SUFFIX);

    mqttClient.setServer(MQTT_BROKER, MQTT_PORT);
    mqttClient.setCallback(onMessage);
    mqttClient.setBufferSize(MQTT_MAX_PACKET_SIZE);

    Serial.printf("MQTT configured:\n");
    Serial.printf("  Broker: %s:%d\n", MQTT_BROKER, MQTT_PORT);
    Serial.printf("  Device ID (MAC): %s\n", deviceMAC);
    Serial.printf("  Command topic: %s\n", deviceTopicCommand);
    Serial.printf("  Status topic: %s\n", deviceTopicStatus);
}

bool MQTTClientManager::connect() {
    Serial.println("Connecting to MQTT broker...");

    // Use MAC as client ID for uniqueness
    bool success = mqttClient.connect(
        deviceMAC,
        MQTT_USER,
        MQTT_PASSWORD
    );

    if (success) {
        Serial.println("MQTT connected!");

        // Subscribe to device-specific command topic
        if (mqttClient.subscribe(deviceTopicCommand)) {
            Serial.printf("✓ Subscribed to: %s\n", deviceTopicCommand);
        } else {
            Serial.println("✗ Failed to subscribe to command topic");
        }

        // Publish device announcement
        publishAnnouncement();

    } else {
        Serial.printf("MQTT connection failed, rc=%d\n", mqttClient.state());
    }

    return success;
}

void MQTTClientManager::disconnect() {
    if (mqttClient.connected()) {
        mqttClient.disconnect();
        Serial.println("MQTT disconnected");
    }
}

void MQTTClientManager::setNetworkManager(ConnectionManager* manager) {
    networkManagerPtr = manager;
}

void MQTTClientManager::update() {
    // Skip MQTT operations if in AP mode (no internet connectivity)
    if (networkManagerPtr && networkManagerPtr->isInAPMode()) {
        return;  // Don't attempt MQTT in AP mode
    }

    // Skip if not actually connected to internet
    if (networkManagerPtr && !networkManagerPtr->isConnected()) {
        return;  // No network connection at all
    }

    if (!mqttClient.connected()) {
        // Auto-reconnect logic
        if (millis() - lastReconnectAttempt > reconnectInterval) {
            lastReconnectAttempt = millis();
            Serial.println("MQTT reconnecting...");
            connect();
        }
    } else {
        mqttClient.loop();
    }
}

bool MQTTClientManager::isConnected() {
    return mqttClient.connected();
}

bool MQTTClientManager::publishAnnouncement() {
    if (!mqttClient.connected()) {
        return false;
    }

    const DeviceConfig::Settings& settings = deviceConfig.getSettings();

    // Create device announcement message
    JsonDocument doc;
    doc["device_id"] = deviceMAC;
    doc["device_type"] = DEVICE_TYPE;
    doc["firmware_version"] = FIRMWARE_VERSION;
    doc["ip_address"] = networkManager.getIP().toString();
    doc["mac_address"] = deviceMAC;

    // Capabilities
    JsonObject caps = doc["capabilities"].to<JsonObject>();
    caps["digital_inputs"] = 8;
    caps["digital_outputs"] = 8;
    caps["ethernet"] = true;
    caps["wifi"] = true;  // Device now supports WiFi

    // Connection information
    JsonObject conn = doc["connection"].to<JsonObject>();
    conn["mode"] = networkManager.getActiveInterface() == ConnectionManager::INTERFACE_WIFI ? "wifi" : "ethernet";
    conn["wifi_enabled"] = settings.wifiEnabled;

    if (networkManager.getActiveInterface() == ConnectionManager::INTERFACE_WIFI) {
        WiFiManager* wifi = networkManager.getWiFiManager();
        conn["wifi_ssid"] = settings.wifiSSID;
        conn["wifi_rssi"] = networkManager.getRSSI();
        conn["ap_mode"] = (wifi && wifi->getMode() == WiFiManager::MODE_AP);
    }

    // Status
    JsonObject status = doc["status"].to<JsonObject>();
    status["uptime_seconds"] = millis() / 1000;
    status["free_heap"] = ESP.getFreeHeap();
    if (networkManager.getActiveInterface() == ConnectionManager::INTERFACE_WIFI) {
        status["rssi"] = networkManager.getRSSI();
    } else {
        status["rssi"] = nullptr;
    }

    doc["timestamp"] = millis();

    char buffer[MQTT_MAX_PACKET_SIZE];
    size_t len = serializeJson(doc, buffer);

    bool success = mqttClient.publish(MQTT_TOPIC_ANNOUNCE, (const uint8_t*)buffer, len, true);  // Retained message

    if (success) {
        Serial.printf("Published device announcement to: %s\n", MQTT_TOPIC_ANNOUNCE);
    } else {
        Serial.println("ERROR: Failed to publish announcement");
    }

    return success;
}

bool MQTTClientManager::publishStatus(uint8_t inputs, uint8_t outputs, bool networkConnected) {
    if (!mqttClient.connected()) {
        return false;
    }

    const DeviceConfig::Settings& settings = deviceConfig.getSettings();

    // Create JSON status message
    JsonDocument doc;
    doc["device_id"] = deviceMAC;
    doc["digital_inputs"] = inputs;
    doc["digital_outputs"] = outputs;
    doc["network_connected"] = networkConnected;
    doc["connection_type"] = networkManager.getActiveInterface() == ConnectionManager::INTERFACE_WIFI ? "wifi" : "ethernet";

    // Add WiFi-specific information
    if (networkManager.getActiveInterface() == ConnectionManager::INTERFACE_WIFI) {
        doc["wifi_rssi"] = networkManager.getRSSI();
        doc["wifi_ssid"] = settings.wifiSSID;
    }

    doc["assigned_line"] = nullptr;  // API will translate via assignment table
    doc["timestamp"] = millis();

    char buffer[MQTT_MAX_PACKET_SIZE];
    size_t len = serializeJson(doc, buffer);

    bool success = mqttClient.publish(deviceTopicStatus, buffer, len);

    if (success) {
        Serial.printf("Published status: inputs=0x%02X outputs=0x%02X\n", inputs, outputs);
    } else {
        Serial.println("ERROR: Failed to publish status");
    }

    return success;
}

bool MQTTClientManager::publishInputChange(uint8_t channel, bool state, uint8_t allInputs) {
    if (!mqttClient.connected()) {
        return false;
    }

    // Create JSON input change event
    char topicBuffer[80];
    snprintf(topicBuffer, sizeof(topicBuffer),
             "%s%s%s", MQTT_TOPIC_DEVICE_PREFIX, deviceMAC, MQTT_TOPIC_INPUT_SUFFIX);

    JsonDocument doc;
    doc["device_id"] = deviceMAC;
    doc["channel"] = channel;
    doc["state"] = state;
    doc["all_inputs"] = allInputs;
    doc["timestamp"] = millis();

    char buffer[256];
    size_t len = serializeJson(doc, buffer);

    bool success = mqttClient.publish(topicBuffer, buffer, len);

    if (success) {
        Serial.printf("Published input change: CH%d=%s\n", channel + 1, state ? "HIGH" : "LOW");
    }

    return success;
}

void MQTTClientManager::setCommandCallback(MQTTCommandCallback callback) {
    cmdCallback = callback;
}

void MQTTClientManager::setFlashCallback(MQTTFlashCallback callback) {
    flashCallback = callback;
}

void MQTTClientManager::onMessage(char* topic, byte* payload, unsigned int length) {
    Serial.printf("MQTT message received on topic: %s\n", topic);

    // Null-terminate payload
    char message[length + 1];
    memcpy(message, payload, length);
    message[length] = '\0';

    Serial.printf("Payload: %s\n", message);

    if (instance != nullptr) {
        instance->handleCommand(message);
    }
}

void MQTTClientManager::handleCommand(const char* payload) {
    // Parse JSON command
    JsonDocument doc;
    DeserializationError error = deserializeJson(doc, payload);

    if (error) {
        Serial.printf("JSON parse error: %s\n", error.c_str());
        return;
    }

    // Extract command fields
    const char* command = doc["command"] | "";

    Serial.printf("Received command: %s\n", command);

    // Handle flash_identify command
    if (strcmp(command, "flash_identify") == 0) {
        uint16_t duration = doc["duration"] | 10;
        Serial.printf("Flash identify command: %d seconds\n", duration);

        if (flashCallback != nullptr) {
            flashCallback();
        }
        return;
    }

    // Handle set_output command
    if (strcmp(command, "set_output") == 0) {
        uint8_t channel = doc["channel"] | 0;
        bool state = doc["state"] | false;

        Serial.printf("Set output command: CH%d=%s\n", channel + 1, state ? "ON" : "OFF");

        if (cmdCallback != nullptr) {
            cmdCallback(command, channel, state);
        }
        return;
    }

    // Handle get_status command
    if (strcmp(command, "get_status") == 0) {
        Serial.println("Get status command - publishing current status");
        // Status will be published in main loop's next heartbeat
        return;
    }

    // ===================================================================
    // WiFi Configuration Commands
    // ===================================================================

    // Handle wifi_configure command
    if (strcmp(command, "wifi_configure") == 0) {
        const char* ssid = doc["ssid"] | "";
        const char* password = doc["password"] | "";
        bool enabled = doc["enabled"] | true;

        Serial.printf("WiFi configure command: SSID='%s', enabled=%s\n",
                     ssid, enabled ? "true" : "false");

        if (strlen(ssid) == 0) {
            Serial.println("ERROR: SSID is required");
            return;
        }

        // Save credentials to config
        if (deviceConfig.setWiFiCredentials(ssid, password)) {
            deviceConfig.enableWiFi(enabled);
            deviceConfig.save();

            Serial.println("WiFi configuration saved. Rebooting in 3 seconds...");
            delay(3000);
            ESP.restart();
        } else {
            Serial.println("ERROR: Failed to save WiFi configuration");
        }
        return;
    }

    // Handle wifi_enable command
    if (strcmp(command, "wifi_enable") == 0) {
        bool enabled = doc["enabled"] | true;

        Serial.printf("WiFi enable command: %s\n", enabled ? "enabled" : "disabled");

        // Check if credentials are configured
        const DeviceConfig::Settings& settings = deviceConfig.getSettings();
        if (enabled && strlen(settings.wifiSSID) == 0) {
            Serial.println("ERROR: WiFi credentials not configured");
            return;
        }

        deviceConfig.enableWiFi(enabled);
        deviceConfig.save();

        Serial.println("WiFi mode changed. Rebooting in 3 seconds...");
        delay(3000);
        ESP.restart();
        return;
    }

    // Handle wifi_disable command
    if (strcmp(command, "wifi_disable") == 0) {
        Serial.println("WiFi disable command - switching to Ethernet");

        deviceConfig.enableWiFi(false);
        deviceConfig.save();

        Serial.println("Switching to Ethernet mode. Rebooting in 3 seconds...");
        delay(3000);
        ESP.restart();
        return;
    }

    // Handle wifi_reset_ap command
    if (strcmp(command, "wifi_reset_ap") == 0) {
        Serial.println("WiFi reset AP command - clearing credentials");

        deviceConfig.clearWiFiCredentials();
        deviceConfig.save();

        Serial.println("WiFi credentials cleared. Rebooting to AP mode in 3 seconds...");
        delay(3000);
        ESP.restart();
        return;
    }

    // Handle get_wifi_status command
    if (strcmp(command, "get_wifi_status") == 0) {
        Serial.println("Get WiFi status command");

        const DeviceConfig::Settings& settings = deviceConfig.getSettings();

        JsonDocument statusDoc;
        statusDoc["device_id"] = deviceMAC;
        statusDoc["wifi_enabled"] = settings.wifiEnabled;
        statusDoc["connection_mode"] = settings.connectionMode == MODE_WIFI ? "wifi" : "ethernet";

        if (networkManager.getActiveInterface() == ConnectionManager::INTERFACE_WIFI) {
            WiFiManager* wifi = networkManager.getWiFiManager();
            statusDoc["wifi_connected"] = wifi && wifi->isConnected();
            statusDoc["wifi_ssid"] = settings.wifiSSID;
            statusDoc["wifi_rssi"] = networkManager.getRSSI();
            statusDoc["ap_mode"] = (wifi && wifi->getMode() == WiFiManager::MODE_AP);
        } else {
            statusDoc["wifi_connected"] = false;
            statusDoc["wifi_ssid"] = strlen(settings.wifiSSID) > 0 ? settings.wifiSSID : nullptr;
            statusDoc["wifi_rssi"] = nullptr;
            statusDoc["ap_mode"] = false;
        }

        statusDoc["ip_address"] = networkManager.getIP().toString();

        char buffer[MQTT_MAX_PACKET_SIZE];
        size_t len = serializeJson(statusDoc, buffer);

        mqttClient.publish(deviceTopicStatus, buffer, len);
        Serial.println("WiFi status published");
        return;
    }

    Serial.printf("Unknown command: %s\n", command);
}
