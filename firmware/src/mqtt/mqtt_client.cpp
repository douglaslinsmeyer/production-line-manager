#include "mqtt_client.h"
#include "config.h"
#include "device_config.h"
#include "network/connection_manager.h"
#include <ETH.h>

// External references
extern DeviceConfig deviceConfig;
extern ConnectionManager networkManager;
extern LineStateManager lineState;

// Static instance pointer for callback
MQTTClientManager* MQTTClientManager::instance = nullptr;

MQTTClientManager::MQTTClientManager()
    : mqttClient(ethClient),
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

    // Get broker configuration from device settings
    const DeviceConfig::Settings& settings = deviceConfig.getSettings();
    const char* broker = (strlen(settings.mqttBroker) > 0) ? settings.mqttBroker : MQTT_BROKER;
    uint16_t port = (settings.mqttPort > 0) ? settings.mqttPort : MQTT_PORT;

    mqttClient.setServer(broker, port);
    mqttClient.setCallback(onMessage);
    mqttClient.setBufferSize(MQTT_MAX_PACKET_SIZE);

    Serial.printf("MQTT configured:\n");
    Serial.printf("  Broker: %s:%d\n", broker, port);
    Serial.printf("  Device ID (MAC): %s\n", deviceMAC);
    Serial.printf("  Command topic: %s\n", deviceTopicCommand);
    Serial.printf("  Status topic: %s\n", deviceTopicStatus);
}

bool MQTTClientManager::connect() {
    Serial.println("Connecting to MQTT broker...");

    // Get MQTT credentials from device configuration
    const DeviceConfig::Settings& settings = deviceConfig.getSettings();

    // Use MAC as client ID for uniqueness
    // Use stored credentials if available, otherwise fall back to compiled defaults
    const char* user = (strlen(settings.mqttUser) > 0) ? settings.mqttUser : MQTT_USER;
    const char* password = (strlen(settings.mqttPassword) > 0) ? settings.mqttPassword : MQTT_PASSWORD;

    bool success = mqttClient.connect(
        deviceMAC,
        user,
        password
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

bool MQTTClientManager::publishStatus(uint8_t inputs, uint8_t outputs, bool networkConnected, LineState lineState) {
    if (!mqttClient.connected()) {
        return false;
    }

    const DeviceConfig::Settings& settings = deviceConfig.getSettings();

    // Create JSON status message
    JsonDocument doc;
    doc["device_id"] = deviceMAC;
    doc["line_state"] = LineStateManager::stateToString(lineState);
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
        Serial.printf("Published status: line_state=%s inputs=0x%02X outputs=0x%02X\n",
                     LineStateManager::stateToString(lineState), inputs, outputs);
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

    // Handle get_status command
    if (strcmp(command, "get_status") == 0) {
        Serial.println("Get status command - publishing current status");
        // Status will be published in main loop's next heartbeat
        return;
    }

    // Handle set_line_state command (from API)
    if (strcmp(command, "set_line_state") == 0) {
        const char* stateStr = doc["state"] | "";

        Serial.printf("Set line state command: %s\n", stateStr);

        // Parse state string to enum
        LineState newState = LINE_STATE_UNKNOWN;
        if (strcmp(stateStr, "ON") == 0) {
            newState = LINE_STATE_ON;
        } else if (strcmp(stateStr, "OFF") == 0) {
            newState = LINE_STATE_OFF;
        } else if (strcmp(stateStr, "MAINTENANCE") == 0) {
            newState = LINE_STATE_MAINTENANCE;
        } else if (strcmp(stateStr, "ERROR") == 0) {
            newState = LINE_STATE_ERROR;
        } else {
            Serial.printf("Invalid state: %s\n", stateStr);
            return;
        }

        // Update line state (will trigger callback which publishes status)
        lineState.setState(newState, "mqtt");

        return;
    }

    Serial.printf("Unknown command: %s\n", command);
}
