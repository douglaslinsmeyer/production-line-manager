#include "mqtt_client.h"
#include "config.h"
#include <ETH.h>

// Static instance pointer for callback
MQTTClientManager* MQTTClientManager::instance = nullptr;

MQTTClientManager::MQTTClientManager()
    : mqttClient(ethClient),
      cmdCallback(nullptr),
      flashCallback(nullptr),
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

void MQTTClientManager::update() {
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

    // Create device announcement message
    JsonDocument doc;
    doc["device_id"] = deviceMAC;
    doc["device_type"] = DEVICE_TYPE;
    doc["firmware_version"] = FIRMWARE_VERSION;
    doc["ip_address"] = ETH.localIP().toString();
    doc["mac_address"] = deviceMAC;

    JsonObject caps = doc["capabilities"].to<JsonObject>();
    caps["digital_inputs"] = 8;
    caps["digital_outputs"] = 8;
    caps["ethernet"] = true;
    caps["wifi"] = false;

    JsonObject status = doc["status"].to<JsonObject>();
    status["uptime_seconds"] = millis() / 1000;
    status["free_heap"] = ESP.getFreeHeap();
    status["rssi"] = nullptr;

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

bool MQTTClientManager::publishStatus(uint8_t inputs, uint8_t outputs, bool ethConnected) {
    if (!mqttClient.connected()) {
        return false;
    }

    // Create JSON status message
    JsonDocument doc;
    doc["device_id"] = deviceMAC;
    doc["digital_inputs"] = inputs;
    doc["digital_outputs"] = outputs;
    doc["ethernet_connected"] = ethConnected;
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

    Serial.printf("Unknown command: %s\n", command);
}
