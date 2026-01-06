#pragma once

#include <Arduino.h>
#include <WebServer.h>
#include "../device_config.h"

// Forward declarations
class ConnectionManager;
class OTAManager;
class ProfileManager;
class ProfileStorage;
class LineStateManager;
class OutputController;

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

    /**
     * Set connection manager reference
     * Allows web server to access network status and RSSI
     * @param manager Pointer to ConnectionManager instance
     */
    void setConnectionManager(ConnectionManager* manager);

    /**
     * Set OTA manager reference
     * Allows web server to handle firmware updates
     * @param manager Pointer to OTAManager instance
     */
    void setOTAManager(OTAManager* manager);

    /**
     * Set profile manager references
     * Allows web server to access profile and state management
     * @param profileMgr Pointer to ProfileManager instance
     * @param profileStore Pointer to ProfileStorage instance
     * @param stateMgr Pointer to LineStateManager instance
     * @param outputCtrl Pointer to OutputController instance
     */
    void setProfileComponents(ProfileManager* profileMgr, ProfileStorage* profileStore,
                             LineStateManager* stateMgr, OutputController* outputCtrl);

private:
    WebServer* webServer;
    ConnectionManager* connectionManager;
    OTAManager* otaManager;
    ProfileManager* profileManager;
    ProfileStorage* profileStorage;
    LineStateManager* lineState;
    OutputController* outputController;
    bool running;
    uint16_t serverPort;
    unsigned long otaStartTime;

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

    // OTA handlers
    void handleOTAPage();
    void handleOTAUpload();
    void handleOTAUploadData();
    void handleOTAProgress();

    // Signal profile API handlers
    void handleProfileView();
    void handleStateView();
    void handleStateSet();
    void handleOverrideClear();
    void handleOutputTest();

    // Signal profile HTML page handlers
    void handleProfilePage();
    void handleStateControlPage();
    void handleOutputTestPage();

    // HTML page generation
    String generateHomePage();
    String generateConfigPage();
    String generateWiFiPage();
    String generateEthernetPage();
    String generateMQTTPage();
    String generateDevicePage();
    String generateOTAPage();
    String generateProfilePage();
    String generateStateControlPage();
    String generateOutputTestPage();

    // Shared HTML components
    String getHTMLHeader(const char* title);
    String getHTMLFooter();
    String getNavigation();
    String getCSS();
};
