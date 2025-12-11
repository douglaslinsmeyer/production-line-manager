#include "wifi_manager.h"

// Static instance pointer for event callback
WiFiManager* WiFiManager::instance = nullptr;

WiFiManager::WiFiManager()
    : currentMode(MODE_OFF),
      connected(false),
      connCallback(nullptr),
      lastReconnectAttempt(0),
      reconnectDelay(WIFI_RECONNECT_INITIAL_DELAY),
      reconnectAttempts(0) {

    instance = this;
}

bool WiFiManager::begin() {
    Serial.println("Initializing WiFi...");

    // Register event handler
    WiFi.onEvent(onWiFiEvent);

    // WiFi will be configured by calling connectSTA() or startAP()
    // based on configuration in main.cpp

    return true;
}

bool WiFiManager::connectSTA(const char* ssid, const char* password, uint32_t timeout) {
    Serial.printf("Connecting to WiFi: %s\n", ssid);

    // Stop any existing connection
    if (currentMode != MODE_OFF) {
        WiFi.disconnect(true);
        delay(100);
    }

    // Store credentials for reconnection
    staSsid = String(ssid);
    staPassword = String(password);

    // Set mode and begin connection
    WiFi.mode(WIFI_STA);
    WiFi.begin(ssid, password);

    currentMode = MODE_STA;
    connected = false;

    // Wait for connection with timeout
    unsigned long startTime = millis();
    while (!connected && (millis() - startTime) < timeout) {
        delay(100);

        if (WiFi.status() == WL_CONNECTED) {
            connected = true;
            Serial.println("✓ WiFi connected");
            Serial.printf("  IP Address: %s\n", WiFi.localIP().toString().c_str());
            Serial.printf("  RSSI: %d dBm\n", WiFi.RSSI());

            // Reset reconnection state
            reconnectAttempts = 0;
            reconnectDelay = WIFI_RECONNECT_INITIAL_DELAY;

            if (connCallback) {
                connCallback(true);
            }

            return true;
        }
    }

    // Connection failed
    Serial.println("✗ WiFi connection timeout");
    Serial.println("  Entering AP mode for setup...");

    // Enter AP mode as fallback
    enterAPMode();

    return false;
}

bool WiFiManager::startAP(const char* ssid, const char* password) {
    // Stop any existing connection
    if (currentMode != MODE_OFF) {
        WiFi.disconnect(true);
        delay(100);
    }

    // Generate AP SSID from MAC if not provided
    String apSSID;
    if (ssid == nullptr || strlen(ssid) == 0) {
        String mac = WiFi.macAddress();
        mac.replace(":", "");
        apSSID = "ESP32-Setup-" + mac.substring(6);  // Last 6 chars of MAC
    } else {
        apSSID = String(ssid);
    }

    Serial.printf("Starting Access Point: %s\n", apSSID.c_str());

    // Configure AP
    WiFi.mode(WIFI_AP);

    bool success;
    if (password == nullptr || strlen(password) == 0) {
        // Open network (no password)
        success = WiFi.softAP(apSSID.c_str());
        Serial.println("  Security: Open (no password)");
    } else {
        // WPA2 protected
        success = WiFi.softAP(apSSID.c_str(), password, WIFI_AP_CHANNEL, 0, WIFI_AP_MAX_CONNECTIONS);
        Serial.println("  Security: WPA2");
    }

    if (success) {
        currentMode = MODE_AP;
        connected = true;  // AP mode is "connected" from device perspective

        IPAddress apIP = WiFi.softAPIP();
        Serial.println("✓ Access Point started");
        Serial.printf("  AP IP: %s\n", apIP.toString().c_str());
        Serial.printf("  Connect to '%s' to configure WiFi\n", apSSID.c_str());

        if (connCallback) {
            connCallback(true);
        }

        return true;
    } else {
        Serial.println("✗ Failed to start Access Point");
        currentMode = MODE_OFF;
        return false;
    }
}

void WiFiManager::stop() {
    Serial.println("Stopping WiFi...");

    if (currentMode != MODE_OFF) {
        WiFi.disconnect(true);
        WiFi.mode(WIFI_OFF);
        currentMode = MODE_OFF;
        connected = false;

        if (connCallback) {
            connCallback(false);
        }
    }
}

void WiFiManager::update() {
    // Only handle reconnection in STA mode
    if (currentMode == MODE_STA && !connected) {
        // Check if it's time to attempt reconnection
        if (millis() - lastReconnectAttempt > reconnectDelay) {
            lastReconnectAttempt = millis();
            reconnectAttempts++;

            Serial.printf("WiFi reconnection attempt %d/%d\n",
                         reconnectAttempts, WIFI_RECONNECT_MAX_ATTEMPTS);

            // Check if max attempts reached
            if (reconnectAttempts >= WIFI_RECONNECT_MAX_ATTEMPTS) {
                Serial.println("✗ Max reconnection attempts reached");
                Serial.println("  Entering AP mode...");
                enterAPMode();
                return;
            }

            // Attempt reconnection
            WiFi.reconnect();

            // Increase backoff delay
            reconnectDelay = calculateBackoffDelay();
        }
    }
}

bool WiFiManager::isConnected() {
    if (currentMode == MODE_STA) {
        return (WiFi.status() == WL_CONNECTED);
    } else if (currentMode == MODE_AP) {
        return true;  // AP mode is always "connected"
    }
    return false;
}

IPAddress WiFiManager::getIP() {
    if (currentMode == MODE_STA) {
        return WiFi.localIP();
    } else if (currentMode == MODE_AP) {
        return WiFi.softAPIP();
    }
    return IPAddress(0, 0, 0, 0);
}

int WiFiManager::getRSSI() {
    if (currentMode == MODE_STA && connected) {
        return WiFi.RSSI();
    }
    return 0;
}

void WiFiManager::setConnectionCallback(WiFiConnectionCallback callback) {
    connCallback = callback;
}

String WiFiManager::getMACAddress() {
    return WiFi.macAddress();
}

void WiFiManager::onWiFiEvent(WiFiEvent_t event, WiFiEventInfo_t info) {
    if (instance == nullptr) return;

    switch (event) {
        case ARDUINO_EVENT_WIFI_STA_CONNECTED:
            Serial.println("WiFi event: CONNECTED");
            break;

        case ARDUINO_EVENT_WIFI_STA_GOT_IP:
            Serial.println("WiFi event: GOT_IP");
            instance->connected = true;
            instance->reconnectAttempts = 0;
            instance->reconnectDelay = WIFI_RECONNECT_INITIAL_DELAY;

            if (instance->connCallback) {
                instance->connCallback(true);
            }
            break;

        case ARDUINO_EVENT_WIFI_STA_DISCONNECTED:
            Serial.println("WiFi event: DISCONNECTED");
            instance->handleDisconnection();
            break;

        case ARDUINO_EVENT_WIFI_STA_LOST_IP:
            Serial.println("WiFi event: LOST_IP");
            instance->connected = false;
            break;

        default:
            break;
    }
}

void WiFiManager::handleDisconnection() {
    connected = false;

    if (connCallback) {
        connCallback(false);
    }

    // Start reconnection process
    lastReconnectAttempt = millis();
}

void WiFiManager::enterAPMode() {
    // Clear stored credentials (user needs to reconfigure)
    staSsid = "";
    staPassword = "";

    // Reset reconnection state
    reconnectAttempts = 0;
    reconnectDelay = WIFI_RECONNECT_INITIAL_DELAY;

    // Start AP mode
    startAP();
}

uint32_t WiFiManager::calculateBackoffDelay() {
    // Exponential backoff: 5s, 10s, 20s, 40s, 60s (capped)
    uint8_t shift = (reconnectAttempts < 3) ? reconnectAttempts : 3;
    uint32_t delay = WIFI_RECONNECT_INITIAL_DELAY * (1 << shift);
    return (delay < WIFI_RECONNECT_MAX_DELAY) ? delay : WIFI_RECONNECT_MAX_DELAY;
}
