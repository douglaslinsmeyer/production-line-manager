#pragma once

#include <Arduino.h>
#include <WiFi.h>
#include "config.h"

// Callback type for WiFi connection events
typedef void (*WiFiConnectionCallback)(bool connected);

/**
 * WiFi Manager
 *
 * Manages WiFi connection in both Station (STA) and Access Point (AP) modes.
 * Handles auto-reconnection with exponential backoff and fallback to AP mode.
 */
class WiFiManager {
public:
    enum Mode {
        MODE_OFF,   // WiFi disabled
        MODE_STA,   // Station mode (client)
        MODE_AP     // Access Point mode (setup)
    };

    WiFiManager();

    /**
     * Initialize WiFi based on configuration
     * Attempts STA connection if credentials exist, otherwise enters AP mode
     * @return true if initialization successful
     */
    bool begin();

    /**
     * Connect to WiFi network in Station mode
     * @param ssid Network SSID
     * @param password Network password (empty for open networks)
     * @param timeout Connection timeout in milliseconds
     * @return true if connected successfully
     */
    bool connectSTA(const char* ssid, const char* password, uint32_t timeout = WIFI_CONNECTION_TIMEOUT);

    /**
     * Start Access Point for initial setup
     * Creates AP with SSID: "ESP32-Setup-{MAC}"
     * @param ssid AP SSID (if null, auto-generated from MAC)
     * @param password AP password (if null, open network)
     * @return true if AP started successfully
     */
    bool startAP(const char* ssid = nullptr, const char* password = nullptr);

    /**
     * Stop WiFi completely
     */
    void stop();

    /**
     * Update WiFi state (call in main loop)
     * Handles reconnection logic and state monitoring
     */
    void update();

    /**
     * Check if WiFi is connected
     * @return true if connected in STA mode or AP is active
     */
    bool isConnected();

    /**
     * Get IP address
     * @return Current IP address (STA or AP)
     */
    IPAddress getIP();

    /**
     * Get current WiFi mode
     * @return MODE_OFF, MODE_STA, or MODE_AP
     */
    Mode getMode() const { return currentMode; }

    /**
     * Get RSSI (signal strength) in STA mode
     * @return Signal strength in dBm (negative number, closer to 0 is better)
     */
    int getRSSI();

    /**
     * Set callback for connection state changes
     * @param callback Function to call on connect/disconnect
     */
    void setConnectionCallback(WiFiConnectionCallback callback);

    /**
     * Get MAC address as string
     * @return MAC address in format XX:XX:XX:XX:XX:XX
     */
    String getMACAddress();

private:
    Mode currentMode;
    bool connected;
    WiFiConnectionCallback connCallback;

    // Reconnection state
    unsigned long lastReconnectAttempt;
    unsigned long reconnectDelay;
    uint8_t reconnectAttempts;

    // Stored credentials for auto-reconnect
    String staSsid;
    String staPassword;

    // WiFi event handling
    static void onWiFiEvent(WiFiEvent_t event, WiFiEventInfo_t info);
    static WiFiManager* instance;

    void handleDisconnection();
    void enterAPMode();
    uint32_t calculateBackoffDelay();
};
