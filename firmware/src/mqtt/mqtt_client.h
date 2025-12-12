#pragma once

#include <Arduino.h>
#include <PubSubClient.h>
#include <WiFiClient.h>
#include <ArduinoJson.h>
#include "state/line_state.h"

// Forward declaration
class ConnectionManager;

// Callback function type for MQTT commands
typedef void (*MQTTCommandCallback)(const char* command, uint8_t channel, bool state);
typedef void (*MQTTFlashCallback)();

// MQTT Client Manager
class MQTTClientManager {
public:
    MQTTClientManager();

    // Initialize MQTT client with MAC address
    void begin(const char* macAddress);

    // Connect to MQTT broker
    bool connect();

    // Disconnect from broker
    void disconnect();

    // Update MQTT client (call in loop)
    void update();

    // Check if connected
    bool isConnected();

    // Publish device announcement (discovery)
    bool publishAnnouncement();

    // Publish status event (with line state)
    bool publishStatus(uint8_t inputs, uint8_t outputs, bool networkConnected, LineState lineState = LINE_STATE_UNKNOWN);

    // Publish input change event
    bool publishInputChange(uint8_t channel, bool state, uint8_t allInputs);

    // Set command callback
    void setCommandCallback(MQTTCommandCallback callback);

    // Set flash identification callback
    void setFlashCallback(MQTTFlashCallback callback);

    // Set network manager reference (for connectivity checks)
    void setNetworkManager(ConnectionManager* manager);

private:
    WiFiClient ethClient;
    PubSubClient mqttClient;
    MQTTCommandCallback cmdCallback;
    MQTTFlashCallback flashCallback;
    ConnectionManager* networkManagerPtr;
    unsigned long lastReconnectAttempt;
    unsigned long reconnectInterval;

    char deviceMAC[18];  // MAC address in format "XX:XX:XX:XX:XX:XX"
    char deviceTopicCommand[64];  // devices/{MAC}/command
    char deviceTopicStatus[64];   // devices/{MAC}/status

    // MQTT callback (static for PubSubClient)
    static void onMessage(char* topic, byte* payload, unsigned int length);
    static MQTTClientManager* instance;

    // Message handling
    void handleCommand(const char* payload);
};
