#include "display_manager.h"
#include "config.h"
#include "network/connection_manager.h"
#include "mqtt/mqtt_client.h"

DisplayManager::DisplayManager()
    : display(nullptr),
      networkManager(nullptr),
      mqttManager(nullptr),
      lastRefresh(0),
      bootTime(0),
      displayInitialized(false),
      lastIPAddress(""),
      lastNetworkConnected(false),
      lastMQTTConnected(false),
      lastRSSI(0) {
}

DisplayManager::~DisplayManager() {
    if (display != nullptr) {
        delete display;
        display = nullptr;
    }
}

bool DisplayManager::begin() {
    Serial.println("Initializing SSD1306 OLED display...");
    Serial.printf("  I2C Address: 0x%02X\n", DISPLAY_I2C_ADDRESS);
    Serial.printf("  Resolution: %dx%d\n", DISPLAY_WIDTH, DISPLAY_HEIGHT);

    // Create display object (using I2C bus already initialized in main.cpp)
    display = new Adafruit_SSD1306(DISPLAY_WIDTH, DISPLAY_HEIGHT, &Wire, -1);

    // Initialize display
    if (!display->begin(SSD1306_SWITCHCAPVCC, DISPLAY_I2C_ADDRESS)) {
        Serial.println("✗ ERROR: SSD1306 allocation failed");
        Serial.println("  Check I2C wiring and address");
        delete display;
        display = nullptr;
        return false;
    }

    // Clear display
    display->clearDisplay();
    display->display();

    // Set boot time for uptime calculation
    bootTime = millis();

    displayInitialized = true;
    Serial.println("✓ Display initialized successfully");

    return true;
}

void DisplayManager::update() {
    // Skip if display not initialized
    if (!displayInitialized || display == nullptr) {
        return;
    }

    // Only refresh every DISPLAY_REFRESH_INTERVAL (2000ms)
    if (millis() - lastRefresh < DISPLAY_REFRESH_INTERVAL) {
        return;
    }

    // Check if state has changed (optimize unnecessary refreshes)
    if (!stateHasChanged()) {
        lastRefresh = millis();
        return;
    }

    // Refresh display content
    refreshDisplay();
    lastRefresh = millis();
}

void DisplayManager::setNetworkManager(ConnectionManager* manager) {
    networkManager = manager;
}

void DisplayManager::setMQTTManager(MQTTClientManager* manager) {
    mqttManager = manager;
}

void DisplayManager::forceRefresh() {
    if (!displayInitialized || display == nullptr) {
        return;
    }

    refreshDisplay();
    lastRefresh = millis();
}

void DisplayManager::showMessage(const char* message) {
    if (!displayInitialized || display == nullptr) {
        return;
    }

    display->clearDisplay();
    display->setTextSize(2);
    display->setTextColor(SSD1306_WHITE);

    // Center text
    int16_t x1, y1;
    uint16_t w, h;
    display->getTextBounds(message, 0, 0, &x1, &y1, &w, &h);
    int x = (DISPLAY_WIDTH - w) / 2;
    int y = (DISPLAY_HEIGHT - h) / 2;

    display->setCursor(x, y);
    display->println(message);
    display->display();
}

void DisplayManager::refreshDisplay() {
    if (networkManager == nullptr || mqttManager == nullptr) {
        // Managers not set yet
        showMessage("Starting...");
        return;
    }

    display->clearDisplay();

    // Query current state from managers
    bool networkConnected = networkManager->isConnected();
    bool mqttConnected = mqttManager->isConnected();
    bool inAPMode = networkManager->isInAPMode();

    if (inAPMode) {
        drawAPMode();
    } else if (networkConnected) {
        drawIPAddress(networkManager->getIP().toString().c_str());
        drawNetworkStatus();
        drawMQTTStatus();
        drawUptime();
    } else {
        drawNoNetwork();
    }

    display->display();  // Push to hardware
}

void DisplayManager::drawIPAddress(const char* ip) {
    display->setTextSize(1);  // 8px tall (smaller to prevent wrapping)
    display->setTextColor(SSD1306_WHITE);
    display->setCursor(0, 0);
    display->print("IP: ");
    display->println(ip);
}

void DisplayManager::drawNetworkStatus() {
    display->setTextSize(1);  // 8px tall
    display->setTextColor(SSD1306_WHITE);
    display->setCursor(0, 12);

    if (networkManager->getActiveInterface() == ConnectionManager::INTERFACE_WIFI) {
        int rssi = networkManager->getRSSI();
        display->printf("WiFi: %ddBm ", rssi);
        display->write(0xFB);  // Checkmark symbol (√)
    } else {
        display->print("Ethernet ");
        display->write(0xFB);  // Checkmark symbol (√)
    }
}

void DisplayManager::drawMQTTStatus() {
    display->setTextSize(1);  // 8px tall
    display->setTextColor(SSD1306_WHITE);
    display->setCursor(0, 22);

    if (mqttManager->isConnected()) {
        display->print("MQTT: Connected");
    } else {
        display->print("MQTT: Disconnected");
    }
}

void DisplayManager::drawUptime() {
    display->setTextSize(1);  // 8px tall
    display->setTextColor(SSD1306_WHITE);
    display->setCursor(0, 32);

    unsigned long uptimeSeconds = (millis() - bootTime) / 1000;
    String uptimeStr = formatUptime(uptimeSeconds);
    display->printf("Up: %s", uptimeStr.c_str());
}

void DisplayManager::drawAPMode() {
    display->setTextSize(1);
    display->setTextColor(SSD1306_WHITE);

    display->setCursor(0, 0);
    display->println("*** SETUP MODE ***");

    display->setCursor(0, 12);
    display->println("SSID: ESP32-Setup");

    display->setCursor(0, 24);
    display->printf("IP: %s", networkManager->getIP().toString().c_str());

    display->setCursor(0, 36);
    display->println("Visit to configure");
}

void DisplayManager::drawNoNetwork() {
    display->setTextSize(2);
    display->setTextColor(SSD1306_WHITE);

    display->setCursor(0, 0);
    display->println("No Network");

    display->setTextSize(1);
    display->setCursor(0, 24);
    display->println("Connecting...");

    // Show connection mode
    display->setCursor(0, 36);
    if (networkManager->getActiveInterface() == ConnectionManager::INTERFACE_WIFI) {
        display->println("[WiFi mode]");
    } else {
        display->println("[Ethernet mode]");
    }

    // Show uptime
    display->setCursor(0, 48);
    unsigned long uptimeSeconds = (millis() - bootTime) / 1000;
    String uptimeStr = formatUptime(uptimeSeconds);
    display->printf("Up: %s", uptimeStr.c_str());
}

String DisplayManager::formatUptime(unsigned long seconds) {
    unsigned long hours = seconds / 3600;
    unsigned long minutes = (seconds % 3600) / 60;
    unsigned long secs = seconds % 60;

    char buffer[16];
    snprintf(buffer, sizeof(buffer), "%02lu:%02lu:%02lu", hours, minutes, secs);
    return String(buffer);
}

String DisplayManager::formatRSSI(int rssi) {
    char buffer[16];
    snprintf(buffer, sizeof(buffer), "%ddBm", rssi);
    return String(buffer);
}

bool DisplayManager::stateHasChanged() {
    if (networkManager == nullptr || mqttManager == nullptr) {
        return true;  // Always refresh if managers not set
    }

    // Get current state
    String currentIP = networkManager->getIP().toString();
    bool currentNetworkConnected = networkManager->isConnected();
    bool currentMQTTConnected = mqttManager->isConnected();
    int currentRSSI = networkManager->getRSSI();

    // Check if anything changed
    bool changed = (currentIP != lastIPAddress) ||
                   (currentNetworkConnected != lastNetworkConnected) ||
                   (currentMQTTConnected != lastMQTTConnected) ||
                   (currentRSSI != lastRSSI);

    // Update cached state
    if (changed) {
        lastIPAddress = currentIP;
        lastNetworkConnected = currentNetworkConnected;
        lastMQTTConnected = currentMQTTConnected;
        lastRSSI = currentRSSI;
    }

    return changed;
}
