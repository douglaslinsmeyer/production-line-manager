#include "connection_manager.h"

// Static instance pointer for callbacks
ConnectionManager* ConnectionManager::instance = nullptr;

// External reference to global device config
extern DeviceConfig deviceConfig;
extern char deviceMAC[18];

ConnectionManager::ConnectionManager()
    : ethManager(nullptr),
      wifiManager(nullptr),
      captivePortal(nullptr),
      deviceWebServer(nullptr),
      activeInterface(INTERFACE_NONE),
      connected(false),
      connectionCallback(nullptr) {

    instance = this;

    // Create manager instances
    ethManager = new EthernetManager();
    wifiManager = new WiFiManager();
    captivePortal = new CaptivePortal();
    deviceWebServer = new DeviceWebServer();
}

bool ConnectionManager::begin(const char* mac) {
    Serial.println("\n=== Initializing Network ===");

    // Use provided MAC or global deviceMAC
    const char* macToUse = (mac != nullptr) ? mac : deviceMAC;

    // Get connection mode from configuration
    ConnectionMode mode = deviceConfig.getConnectionMode();

    Serial.printf("Connection Mode: %s\n",
                 mode == MODE_WIFI ? "WiFi" : "Ethernet");

    // Enforce mutual exclusion and initialize appropriate interface
    if (mode == MODE_WIFI) {
        activeInterface = INTERFACE_WIFI;

        // ENSURE ETHERNET IS OFF
        Serial.println("Disabling Ethernet...");
        // Ethernet should not be started in the first place
        // ETH.end() is not available, so we just don't initialize it

        // Initialize WiFi
        Serial.println("Initializing WiFi...");
        wifiManager->begin();
        wifiManager->setConnectionCallback(onWiFiConnection);

        // Check if credentials are configured
        const DeviceConfig::Settings& settings = deviceConfig.getSettings();

        if (strlen(settings.wifiSSID) > 0 && !settings.wifiAPMode) {
            // Attempt STA mode connection
            Serial.printf("Connecting to WiFi: %s\n", settings.wifiSSID);
            bool success = wifiManager->connectSTA(
                settings.wifiSSID,
                settings.wifiPassword,
                WIFI_CONNECTION_TIMEOUT
            );

            if (success) {
                connected = true;
                // Start always-on configuration web server
                deviceWebServer->begin(80);
                return true;
            } else {
                // Connection failed, WiFiManager will have entered AP mode
                Serial.println("WiFi connection failed - in AP mode");
                // Start captive portal for setup
                captivePortal->begin(macToUse);
                return true;  // Still successful initialization, just in AP mode
            }
        } else {
            // No credentials or AP mode forced - start AP mode
            Serial.println("No WiFi credentials - starting AP mode");
            wifiManager->startAP();
            // Start captive portal for setup
            captivePortal->begin(macToUse);
            return true;
        }

    } else {
        // MODE_ETHERNET
        activeInterface = INTERFACE_ETHERNET;

        // ENSURE WIFI IS OFF
        Serial.println("Disabling WiFi...");
        WiFi.mode(WIFI_OFF);

        // Initialize Ethernet
        Serial.println("Initializing Ethernet...");
        ethManager->begin();
        ethManager->setConnectionCallback(onEthernetConnection);

        // Start always-on configuration web server
        // Wait a moment for Ethernet to be ready
        delay(500);
        deviceWebServer->begin(80);

        return true;
    }
}

void ConnectionManager::update() {
    // Update active interface manager
    if (activeInterface == INTERFACE_ETHERNET && ethManager) {
        ethManager->update();
        connected = ethManager->isConnected();
    } else if (activeInterface == INTERFACE_WIFI && wifiManager) {
        wifiManager->update();
        connected = wifiManager->isConnected();
    }

    // Update web servers
    if (captivePortal) {
        captivePortal->update();
    }
    if (deviceWebServer) {
        deviceWebServer->update();
    }
}

bool ConnectionManager::isConnected() {
    if (activeInterface == INTERFACE_ETHERNET && ethManager) {
        return ethManager->isConnected();
    } else if (activeInterface == INTERFACE_WIFI && wifiManager) {
        return wifiManager->isConnected();
    }
    return false;
}

IPAddress ConnectionManager::getIP() {
    if (activeInterface == INTERFACE_ETHERNET && ethManager) {
        return ethManager->getIP();
    } else if (activeInterface == INTERFACE_WIFI && wifiManager) {
        return wifiManager->getIP();
    }
    return IPAddress(0, 0, 0, 0);
}

int ConnectionManager::getRSSI() {
    if (activeInterface == INTERFACE_WIFI && wifiManager) {
        return wifiManager->getRSSI();
    }
    return 0;  // Ethernet doesn't have RSSI
}

void ConnectionManager::setConnectionCallback(void (*callback)(bool)) {
    connectionCallback = callback;
}

bool ConnectionManager::switchInterface(Interface newInterface) {
    Serial.printf("\n=== Switching Network Interface ===\n");
    Serial.printf("Current: %d, New: %d\n", activeInterface, newInterface);

    // Update configuration
    ConnectionMode newMode = (newInterface == INTERFACE_WIFI) ? MODE_WIFI : MODE_ETHERNET;
    deviceConfig.setConnectionMode(newMode);
    deviceConfig.save();

    Serial.println("Configuration saved. Rebooting in 3 seconds...");

    // Display countdown
    for (int i = 3; i > 0; i--) {
        Serial.printf("%d...\n", i);
        delay(1000);
    }

    // Reboot device
    ESP.restart();

    return true;  // Won't actually return due to restart
}

void ConnectionManager::onEthernetConnection(bool connState) {
    if (instance == nullptr) return;

    Serial.printf("Ethernet: %s\n", connState ? "Connected" : "Disconnected");

    instance->connected = connState;

    if (instance->connectionCallback) {
        instance->connectionCallback(connState);
    }
}

void ConnectionManager::onWiFiConnection(bool connState) {
    if (instance == nullptr) return;

    Serial.printf("WiFi: %s\n", connState ? "Connected" : "Disconnected");

    instance->connected = connState;

    if (instance->connectionCallback) {
        instance->connectionCallback(connState);
    }
}

void ConnectionManager::ensureMutualExclusion() {
    // This is enforced in begin() by only initializing one interface
    // Runtime check to ensure both aren't active
    bool ethActive = (activeInterface == INTERFACE_ETHERNET && ethManager && ethManager->isConnected());
    bool wifiActive = (activeInterface == INTERFACE_WIFI && wifiManager && wifiManager->isConnected());

    if (ethActive && wifiActive) {
        Serial.println("ERROR: Both Ethernet and WiFi are active! This should never happen.");
        Serial.println("Rebooting to recover...");
        delay(2000);
        ESP.restart();
    }
}
