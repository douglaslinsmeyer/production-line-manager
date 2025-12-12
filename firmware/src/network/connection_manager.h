#pragma once

#include <Arduino.h>
#include <ETH.h>
#include "../ethernet/eth_manager.h"
#include "../wifi/wifi_manager.h"
#include "../wifi/captive_portal.h"
#include "../wifi/device_webserver.h"
#include "../device_config.h"

/**
 * Network Manager
 *
 * Unified interface for both WiFi and Ethernet connections.
 * Enforces mutual exclusion: only ONE interface active at a time.
 * Handles switching between interfaces with proper cleanup.
 */
class ConnectionManager {
public:
    enum Interface {
        INTERFACE_NONE,
        INTERFACE_ETHERNET,
        INTERFACE_WIFI
    };

    ConnectionManager();

    /**
     * Initialize network based on device configuration
     * Ensures only one interface is active (WiFi OR Ethernet)
     * @param deviceMAC Device MAC address for web server display
     * @return true if initialization successful
     */
    bool begin(const char* deviceMAC = nullptr);

    /**
     * Update network state (call in main loop)
     * Updates active interface manager
     */
    void update();

    /**
     * Check if network is connected
     * @return true if active interface is connected
     */
    bool isConnected();

    /**
     * Get IP address from active interface
     * @return Current IP address
     */
    IPAddress getIP();

    /**
     * Get currently active interface
     * @return INTERFACE_NONE, INTERFACE_ETHERNET, or INTERFACE_WIFI
     */
    Interface getActiveInterface() const { return activeInterface; }

    /**
     * Get RSSI (WiFi only, returns 0 for Ethernet)
     * @return Signal strength in dBm
     */
    int getRSSI();

    /**
     * Set connection callback for interface changes
     * @param callback Function to call on connect/disconnect
     */
    void setConnectionCallback(void (*callback)(bool));

    /**
     * Get reference to Ethernet manager
     * @return Pointer to EthernetManager
     */
    EthernetManager* getEthernetManager() { return ethManager; }

    /**
     * Get reference to WiFi manager
     * @return Pointer to WiFiManager
     */
    WiFiManager* getWiFiManager() { return wifiManager; }

    /**
     * Get reference to captive portal
     * @return Pointer to CaptivePortal
     */
    CaptivePortal* getCaptivePortal() { return captivePortal; }

    /**
     * Get reference to device web server
     * @return Pointer to DeviceWebServer
     */
    DeviceWebServer* getWebServer() { return deviceWebServer; }

    /**
     * Check if WiFi is in AP mode
     * @return true if in AP mode, false otherwise
     */
    bool isInAPMode() const;

    /**
     * Switch to different interface (requires reboot in practice)
     * This method saves configuration and reboots device
     * @param newInterface Interface to switch to
     * @return true if switch initiated
     */
    bool switchInterface(Interface newInterface);

private:
    EthernetManager* ethManager;
    WiFiManager* wifiManager;
    CaptivePortal* captivePortal;
    DeviceWebServer* deviceWebServer;
    Interface activeInterface;
    bool connected;

    void (*connectionCallback)(bool);

    // Static callbacks for interface events
    static void onEthernetConnection(bool connected);
    static void onWiFiConnection(bool connected);
    static ConnectionManager* instance;

    // Ensure only one interface is active
    void ensureMutualExclusion();
};
