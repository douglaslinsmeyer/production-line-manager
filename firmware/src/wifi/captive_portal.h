#pragma once

#include <Arduino.h>
#include <WebServer.h>
#include <DNSServer.h>
#include <WiFi.h>

// Forward declaration
class DeviceConfig;

// Callback type for when credentials are saved
typedef void (*CredentialsSavedCallback)(const char* ssid, const char* password);

/**
 * Captive Portal
 *
 * Provides a web interface for WiFi configuration in AP mode.
 * Includes DNS server for captive portal redirection and
 * web server for configuration page with network scanner.
 */
class CaptivePortal {
public:
    CaptivePortal();
    ~CaptivePortal();

    /**
     * Start captive portal (AP mode must be active)
     * @param deviceMAC MAC address for display
     * @return true if started successfully
     */
    bool begin(const char* deviceMAC);

    /**
     * Stop captive portal
     */
    void stop();

    /**
     * Update (call in main loop)
     * Handles DNS and web server requests
     */
    void update();

    /**
     * Check if credentials have been submitted
     * @return true if new credentials are available
     */
    bool hasCredentials() const { return credentialsSet; }

    /**
     * Get submitted credentials
     * @param ssid Output parameter for SSID
     * @param password Output parameter for password
     */
    void getCredentials(String& ssid, String& password);

    /**
     * Set callback for when credentials are saved
     * @param callback Function to call with SSID and password
     */
    void setCredentialsSavedCallback(CredentialsSavedCallback callback);

    /**
     * Set device config reference (for NVS credential storage)
     * @param config Device configuration manager
     */
    void setDeviceConfig(DeviceConfig* config);

private:
    WebServer* webServer;
    DNSServer* dnsServer;
    DeviceConfig* deviceConfig;

    bool credentialsSet;
    String submittedSSID;
    String submittedPassword;
    String deviceMAC;

    CredentialsSavedCallback savedCallback;

    // HTTP request handlers
    void handleRoot();
    void handleScan();
    void handleSave();
    void handleNotFound();

    // HTML page generation
    String generateSetupPage();
    String generateSuccessPage();
    String getEncryptionType(wifi_auth_mode_t encryption);

    static const uint8_t DNS_PORT = 53;
    static const uint16_t WEB_PORT = 80;
    static const IPAddress AP_IP;
};
