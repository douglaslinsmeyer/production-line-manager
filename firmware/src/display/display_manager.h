#pragma once

#include <Arduino.h>
#include <Adafruit_GFX.h>
#include <Adafruit_SSD1306.h>

// Forward declarations
class ConnectionManager;
class MQTTClientManager;

/**
 * Display Manager
 *
 * Controls SSD1306 OLED display (128x64) to show real-time network
 * and connectivity status information.
 *
 * Display Layout:
 * - Line 0: IP address (large font, 16px)
 * - Line 2: Network type + signal strength + status
 * - Line 3: MQTT connection status
 * - Line 4: System uptime
 *
 * Uses non-blocking update pattern consistent with other modules.
 * Display refreshes every DISPLAY_REFRESH_INTERVAL (2000ms by default).
 */
class DisplayManager {
public:
    DisplayManager();
    ~DisplayManager();

    /**
     * Initialize display hardware
     * @return true if display initialized successfully
     */
    bool begin();

    /**
     * Update display (call in main loop)
     * Non-blocking with timed refresh to avoid slowing loop
     */
    void update();

    /**
     * Set references to network and MQTT managers
     * Allows display to query current connection state
     */
    void setNetworkManager(ConnectionManager* manager);
    void setMQTTManager(MQTTClientManager* manager);

    /**
     * Force immediate display refresh (bypass timed refresh)
     * Useful for critical state changes
     */
    void forceRefresh();

    /**
     * Clear display and show message
     * @param message Text to display (centered)
     */
    void showMessage(const char* message);

private:
    Adafruit_SSD1306* display;
    ConnectionManager* networkManager;
    MQTTClientManager* mqttManager;

    unsigned long lastRefresh;
    unsigned long bootTime;
    bool displayInitialized;

    // Display state tracking (for change detection)
    String lastIPAddress;
    bool lastNetworkConnected;
    bool lastMQTTConnected;
    int lastRSSI;

    // Refresh display content based on current state
    void refreshDisplay();

    // Helper functions for rendering specific information
    void drawIPAddress(const char* ip);
    void drawNetworkStatus();
    void drawMQTTStatus();
    void drawUptime();
    void drawAPMode();
    void drawNoNetwork();

    // Utility functions
    String formatUptime(unsigned long seconds);
    String formatRSSI(int rssi);
    bool stateHasChanged();
};
