#pragma once

#include <Arduino.h>
#include <WebServer.h>
#include "../device_config.h"

/**
 * Device Web Server
 *
 * Always-on web server for device configuration accessible at device IP.
 * Provides web-based UI for configuring WiFi, Ethernet, MQTT, and device settings.
 * Runs on port 80 when not in AP mode, or port 8080 to avoid conflicts.
 */
class DeviceWebServer {
public:
    DeviceWebServer();
    ~DeviceWebServer();

    /**
     * Start the web server
     * @param port Port to listen on (default 80, use 8080 if captive portal active)
     * @return true if started successfully
     */
    bool begin(uint16_t port = 80);

    /**
     * Stop the web server
     */
    void stop();

    /**
     * Update (call in main loop)
     * Handles HTTP requests
     */
    void update();

    /**
     * Check if server is running
     * @return true if server is active
     */
    bool isRunning() const { return running; }

private:
    WebServer* webServer;
    bool running;
    uint16_t serverPort;

    // HTTP request handlers
    void handleRoot();
    void handleConfig();
    void handleWiFiConfig();
    void handleEthernetConfig();
    void handleMQTTConfig();
    void handleDeviceConfig();
    void handleSaveWiFi();
    void handleSaveEthernet();
    void handleSaveMQTT();
    void handleSaveDevice();
    void handleReboot();
    void handleReset();
    void handleStatus();
    void handleNotFound();

    // HTML page generation
    String generateHomePage();
    String generateConfigPage();
    String generateWiFiPage();
    String generateEthernetPage();
    String generateMQTTPage();
    String generateDevicePage();

    // Shared HTML components
    String getHTMLHeader(const char* title);
    String getHTMLFooter();
    String getNavigation();
    String getCSS();
};
