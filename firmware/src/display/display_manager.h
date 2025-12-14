#pragma once

#include <Arduino.h>
#include <Adafruit_GFX.h>
#include <Adafruit_SSD1306.h>

// Forward declarations
class ConnectionManager;
class MQTTClientManager;

// WiFi signal strength bitmaps (8x8 pixels)
// Traditional WiFi icon with radiating arcs
// 0 bars (no signal - X through icon)
static const uint8_t WIFI_ICON_0[] PROGMEM = {
    0b10000001,
    0b01000010,
    0b00100100,
    0b00011000,
    0b00011000,
    0b00100100,
    0b01000010,
    0b10000001
};

// 1 bar (poor signal - just dot)
static const uint8_t WIFI_ICON_1[] PROGMEM = {
    0b00000000,
    0b00000000,
    0b00000000,
    0b00000000,
    0b00000000,
    0b00000000,
    0b00011000,
    0b00011000
};

// 2 bars (fair signal - dot + small arc)
static const uint8_t WIFI_ICON_2[] PROGMEM = {
    0b00000000,
    0b00000000,
    0b00000000,
    0b00000000,
    0b00111100,
    0b01000010,
    0b00011000,
    0b00011000
};

// 3 bars (good signal - dot + 2 arcs)
static const uint8_t WIFI_ICON_3[] PROGMEM = {
    0b00000000,
    0b00000000,
    0b01111110,
    0b10000001,
    0b00111100,
    0b01000010,
    0b00011000,
    0b00011000
};

// 4 bars (excellent signal - dot + 3 arcs)
static const uint8_t WIFI_ICON_4[] PROGMEM = {
    0b11111111,
    0b10000001,
    0b01111110,
    0b10000001,
    0b00111100,
    0b01000010,
    0b00011000,
    0b00011000
};

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

    /**
     * Show OTA update progress on display
     * @param percent Progress percentage (0-100)
     */
    void showOTAProgress(uint8_t percent);

    /**
     * Show OTA update complete message
     */
    void showOTAComplete();

    /**
     * Show OTA error message
     * @param error Error message string
     */
    void showOTAError(const char* error);

    /**
     * Check if OTA is in progress
     * @return true if OTA update is being displayed
     */
    bool isOTAInProgress() const { return otaInProgress; }

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

    // OTA state tracking
    bool otaInProgress;
    uint8_t lastOTAPercent;
    unsigned long lastOTAUpdate;

    // Refresh display content based on current state
    void refreshDisplay();

    // Helper functions for rendering specific information
    void drawIPAddress(const char* ip);
    void drawNetworkStatus();
    void drawMQTTStatus();
    void drawUptime();
    void drawAPMode();
    void drawNoNetwork();
    void drawWiFiIcon(int16_t x, int16_t y, uint8_t bars);

    // Utility functions
    String formatUptime(unsigned long seconds);
    bool stateHasChanged();
};
