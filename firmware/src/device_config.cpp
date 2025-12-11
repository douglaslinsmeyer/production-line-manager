#include "device_config.h"

// Global instance
DeviceConfig deviceConfig;

DeviceConfig::DeviceConfig() {
    memset(&settings, 0, sizeof(settings));
}

void DeviceConfig::begin() {
    prefs.begin("device_cfg", false);  // false = read/write mode
    loadSettings();
}

void DeviceConfig::loadSettings() {
    // Load existing network settings
    prefs.getString("device_id", settings.deviceID, sizeof(settings.deviceID));
    prefs.getString("line_code", settings.lineCode, sizeof(settings.lineCode));
    prefs.getString("mqtt_broker", settings.mqttBroker, sizeof(settings.mqttBroker));
    settings.mqttPort = prefs.getUShort("mqtt_port", 1883);
    prefs.getString("mqtt_user", settings.mqttUser, sizeof(settings.mqttUser));
    prefs.getString("mqtt_pass", settings.mqttPassword, sizeof(settings.mqttPassword));
    settings.useDHCP = prefs.getBool("use_dhcp", true);
    prefs.getString("static_ip", settings.staticIP, sizeof(settings.staticIP));
    prefs.getString("gateway", settings.gateway, sizeof(settings.gateway));
    prefs.getString("subnet", settings.subnet, sizeof(settings.subnet));
    prefs.getString("dns", settings.dnsServer, sizeof(settings.dnsServer));

    // Load WiFi settings with defaults
    settings.connectionMode = (ConnectionMode)prefs.getUChar("conn_mode", MODE_ETHERNET);
    settings.wifiEnabled = prefs.getBool("wifi_en", false);
    prefs.getString("wifi_ssid", settings.wifiSSID, sizeof(settings.wifiSSID));
    prefs.getString("wifi_pass", settings.wifiPassword, sizeof(settings.wifiPassword));
    settings.wifiAPMode = prefs.getBool("wifi_ap", false);

    // Apply defaults if empty
    if (strlen(settings.deviceID) == 0) {
        loadDefaults();
    }
}

void DeviceConfig::loadDefaults() {
    strncpy(settings.deviceID, "ESP32-Device", sizeof(settings.deviceID) - 1);
    strncpy(settings.lineCode, "LINE-01", sizeof(settings.lineCode) - 1);
    strncpy(settings.mqttBroker, "10.221.21.100", sizeof(settings.mqttBroker) - 1);
    settings.mqttPort = 1883;
    settings.useDHCP = true;

    // WiFi defaults
    settings.connectionMode = MODE_ETHERNET;  // Default to Ethernet
    settings.wifiEnabled = false;
    settings.wifiAPMode = false;
    memset(settings.wifiSSID, 0, sizeof(settings.wifiSSID));
    memset(settings.wifiPassword, 0, sizeof(settings.wifiPassword));
}

bool DeviceConfig::save() {
    // Save existing settings
    prefs.putString("device_id", settings.deviceID);
    prefs.putString("line_code", settings.lineCode);
    prefs.putString("mqtt_broker", settings.mqttBroker);
    prefs.putUShort("mqtt_port", settings.mqttPort);
    prefs.putString("mqtt_user", settings.mqttUser);
    prefs.putString("mqtt_pass", settings.mqttPassword);
    prefs.putBool("use_dhcp", settings.useDHCP);
    prefs.putString("static_ip", settings.staticIP);
    prefs.putString("gateway", settings.gateway);
    prefs.putString("subnet", settings.subnet);
    prefs.putString("dns", settings.dnsServer);

    // Save WiFi settings
    prefs.putUChar("conn_mode", settings.connectionMode);
    prefs.putBool("wifi_en", settings.wifiEnabled);
    prefs.putString("wifi_ssid", settings.wifiSSID);
    prefs.putString("wifi_pass", settings.wifiPassword);
    prefs.putBool("wifi_ap", settings.wifiAPMode);

    return true;
}

bool DeviceConfig::setDeviceID(const char* id) {
    if (strlen(id) == 0 || strlen(id) >= sizeof(settings.deviceID)) {
        return false;
    }
    strncpy(settings.deviceID, id, sizeof(settings.deviceID) - 1);
    settings.deviceID[sizeof(settings.deviceID) - 1] = '\0';
    return save();
}

bool DeviceConfig::setLineCode(const char* code) {
    if (strlen(code) == 0 || strlen(code) >= sizeof(settings.lineCode)) {
        return false;
    }
    strncpy(settings.lineCode, code, sizeof(settings.lineCode) - 1);
    settings.lineCode[sizeof(settings.lineCode) - 1] = '\0';
    return save();
}

bool DeviceConfig::setMQTTBroker(const char* broker, uint16_t port) {
    if (strlen(broker) == 0 || strlen(broker) >= sizeof(settings.mqttBroker)) {
        return false;
    }
    strncpy(settings.mqttBroker, broker, sizeof(settings.mqttBroker) - 1);
    settings.mqttBroker[sizeof(settings.mqttBroker) - 1] = '\0';
    settings.mqttPort = port;
    return save();
}

bool DeviceConfig::setMQTTAuth(const char* user, const char* password) {
    if (strlen(user) >= sizeof(settings.mqttUser) || strlen(password) >= sizeof(settings.mqttPassword)) {
        return false;
    }
    strncpy(settings.mqttUser, user, sizeof(settings.mqttUser) - 1);
    settings.mqttUser[sizeof(settings.mqttUser) - 1] = '\0';
    strncpy(settings.mqttPassword, password, sizeof(settings.mqttPassword) - 1);
    settings.mqttPassword[sizeof(settings.mqttPassword) - 1] = '\0';
    return save();
}

bool DeviceConfig::setNetworkMode(bool dhcp) {
    settings.useDHCP = dhcp;
    return save();
}

bool DeviceConfig::setStaticIP(const char* ip, const char* gateway, const char* subnet, const char* dns) {
    if (strlen(ip) >= sizeof(settings.staticIP) ||
        strlen(gateway) >= sizeof(settings.gateway) ||
        strlen(subnet) >= sizeof(settings.subnet) ||
        strlen(dns) >= sizeof(settings.dnsServer)) {
        return false;
    }

    strncpy(settings.staticIP, ip, sizeof(settings.staticIP) - 1);
    settings.staticIP[sizeof(settings.staticIP) - 1] = '\0';
    strncpy(settings.gateway, gateway, sizeof(settings.gateway) - 1);
    settings.gateway[sizeof(settings.gateway) - 1] = '\0';
    strncpy(settings.subnet, subnet, sizeof(settings.subnet) - 1);
    settings.subnet[sizeof(settings.subnet) - 1] = '\0';
    strncpy(settings.dnsServer, dns, sizeof(settings.dnsServer) - 1);
    settings.dnsServer[sizeof(settings.dnsServer) - 1] = '\0';
    settings.useDHCP = false;
    return save();
}

// WiFi Configuration Methods

bool DeviceConfig::setConnectionMode(ConnectionMode mode) {
    settings.connectionMode = mode;
    return save();
}

bool DeviceConfig::setWiFiCredentials(const char* ssid, const char* password) {
    // Validate SSID (must be non-empty, max 32 chars for ESP32)
    if (strlen(ssid) == 0 || strlen(ssid) > 32) {
        Serial.println("ERROR: SSID must be 1-32 characters");
        return false;
    }

    // Validate password (0 for open, or 8-63 chars for WPA2)
    size_t passLen = strlen(password);
    if (passLen != 0 && (passLen < 8 || passLen > 63)) {
        Serial.println("ERROR: Password must be 0 (open) or 8-63 characters");
        return false;
    }

    // Copy credentials
    strncpy(settings.wifiSSID, ssid, sizeof(settings.wifiSSID) - 1);
    settings.wifiSSID[sizeof(settings.wifiSSID) - 1] = '\0';
    strncpy(settings.wifiPassword, password, sizeof(settings.wifiPassword) - 1);
    settings.wifiPassword[sizeof(settings.wifiPassword) - 1] = '\0';

    // Clear AP mode flag when credentials are set
    settings.wifiAPMode = false;

    Serial.println("WiFi credentials saved");
    return save();
}

bool DeviceConfig::clearWiFiCredentials() {
    memset(settings.wifiSSID, 0, sizeof(settings.wifiSSID));
    memset(settings.wifiPassword, 0, sizeof(settings.wifiPassword));
    settings.wifiAPMode = true;  // Force AP mode when credentials are cleared

    Serial.println("WiFi credentials cleared - AP mode will activate on next boot");
    return save();
}

bool DeviceConfig::enableWiFi(bool enable) {
    settings.wifiEnabled = enable;
    settings.connectionMode = enable ? MODE_WIFI : MODE_ETHERNET;

    Serial.printf("WiFi %s - connection mode set to %s\n",
                 enable ? "enabled" : "disabled",
                 enable ? "WiFi" : "Ethernet");
    return save();
}

bool DeviceConfig::isWiFiEnabled() const {
    return settings.wifiEnabled;
}

ConnectionMode DeviceConfig::getConnectionMode() const {
    return settings.connectionMode;
}

void DeviceConfig::resetToDefaults() {
    prefs.clear();
    loadDefaults();
    save();
    Serial.println("Configuration reset to factory defaults");
}

void DeviceConfig::printSettings() {
    Serial.println("\n=== Device Configuration ===");
    Serial.printf("Device ID:     %s\n", settings.deviceID);
    Serial.printf("Line Code:     %s\n", settings.lineCode);
    Serial.printf("MQTT Broker:   %s:%d\n", settings.mqttBroker, settings.mqttPort);
    Serial.printf("MQTT User:     %s\n", strlen(settings.mqttUser) > 0 ? settings.mqttUser : "(none)");
    Serial.printf("Network Mode:  %s\n", settings.useDHCP ? "DHCP" : "Static IP");

    if (!settings.useDHCP) {
        Serial.printf("Static IP:     %s\n", settings.staticIP);
        Serial.printf("Gateway:       %s\n", settings.gateway);
        Serial.printf("Subnet:        %s\n", settings.subnet);
        Serial.printf("DNS:           %s\n", settings.dnsServer);
    }

    // WiFi settings
    Serial.println("\n--- WiFi Configuration ---");
    Serial.printf("Connection Mode: %s\n", settings.connectionMode == MODE_WIFI ? "WiFi" : "Ethernet");
    Serial.printf("WiFi Enabled:    %s\n", settings.wifiEnabled ? "Yes" : "No");
    Serial.printf("WiFi SSID:       %s\n", strlen(settings.wifiSSID) > 0 ? settings.wifiSSID : "(not configured)");
    Serial.printf("WiFi Password:   %s\n", strlen(settings.wifiPassword) > 0 ? "****" : "(not set)");
    Serial.printf("AP Mode:         %s\n", settings.wifiAPMode ? "Yes" : "No");
    Serial.println("============================\n");
}

void DeviceConfig::interactiveSetup() {
    Serial.println("\n=== Interactive Device Configuration ===");
    Serial.println("This feature allows configuration via serial console.");
    Serial.println("Future implementation: Use commands to configure settings.");
    Serial.println("========================================\n");
    // TODO: Implement interactive serial configuration if needed
}
