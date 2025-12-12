#pragma once

#include <Arduino.h>
#include <Preferences.h>

// Connection mode for network interface
enum ConnectionMode {
    MODE_ETHERNET = 0,
    MODE_WIFI = 1
};

// Device Configuration Manager using NVS (Non-Volatile Storage)
class DeviceConfig {
public:
    struct Settings {
        char deviceID[32];
        char lineCode[32];
        char mqttBroker[64];
        uint16_t mqttPort;
        char mqttUser[32];
        char mqttPassword[32];
        bool useDHCP;
        char staticIP[16];
        char gateway[16];
        char subnet[16];
        char dnsServer[16];

        // WiFi Configuration
        ConnectionMode connectionMode;   // Which interface to use (Ethernet or WiFi)
        bool wifiEnabled;                // WiFi enable flag
        char wifiSSID[64];              // WiFi network name (max 32 for ESP32, but allow 64)
        char wifiPassword[64];          // WiFi password (WPA2 max 63 chars)
        bool wifiAPMode;                // Currently in AP mode?
    };

    DeviceConfig();

    // Initialize and load settings from NVS
    void begin();

    // Get current settings
    const Settings& getSettings() const { return settings; }

    // Update settings (and save to NVS)
    bool setDeviceID(const char* id);
    bool setLineCode(const char* code);
    bool setMQTTBroker(const char* broker, uint16_t port = 1883);
    bool setMQTTAuth(const char* user, const char* password);
    bool setNetworkMode(bool dhcp);
    bool setStaticIP(const char* ip, const char* gateway, const char* subnet, const char* dns = "8.8.8.8");

    // WiFi configuration methods
    bool setConnectionMode(ConnectionMode mode);
    bool setWiFiCredentials(const char* ssid, const char* password);
    bool clearWiFiCredentials();
    bool enableWiFi(bool enable);
    bool isWiFiEnabled() const;
    bool isWiFiAPMode() const;
    ConnectionMode getConnectionMode() const;

    // Reset to factory defaults
    void resetToDefaults();

    // Save all current settings to NVS
    bool save();

    // Interactive configuration via serial console
    void interactiveSetup();

    // Print current configuration
    void printSettings();

private:
    Preferences prefs;
    Settings settings;

    void loadSettings();
    void loadDefaults();
};

// Global configuration instance
extern DeviceConfig deviceConfig;
