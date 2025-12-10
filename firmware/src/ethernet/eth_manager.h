#pragma once

#include <Arduino.h>
#include <ETH.h>
#include <SPI.h>

// Callback function type for connection state changes
typedef void (*ConnectionCallback)(bool connected);

// Ethernet Manager for W5500
class EthernetManager {
public:
    EthernetManager();

    // Initialize W5500 Ethernet
    bool begin();

    // Update ethernet state (call in loop)
    void update();

    // Check if connected
    bool isConnected();

    // Get IP address
    IPAddress getIP();

    // Set connection state callback
    void setConnectionCallback(ConnectionCallback callback);

private:
    bool connected;
    ConnectionCallback connCallback;
    unsigned long lastStatusCheck;

    static void onEvent(arduino_event_id_t event, arduino_event_info_t info);
    static EthernetManager* instance;  // For static callback
};
