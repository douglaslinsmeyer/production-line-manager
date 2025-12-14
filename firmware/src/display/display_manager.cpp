#include "display_manager.h"
#include "config.h"
#include "network/connection_manager.h"
#include "mqtt/mqtt_client.h"
#include "wifi/wifi_utils.h"

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
      lastRSSI(0),
      otaInProgress(false),
      lastOTAPercent(0),
      lastOTAUpdate(0) {
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

    // Skip normal updates during OTA
    if (otaInProgress) {
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

        if (rssi == 0) {
            // WiFi disconnected or not connected yet
            display->print("WiFi: Connecting...");
        } else {
            // Connected - show WiFi icon with signal bars
            display->print("WiFi: ");

            // Get cursor position to draw icon after text
            int16_t x = display->getCursorX();
            int16_t y = display->getCursorY();

            // Draw WiFi signal icon
            uint8_t bars = WiFiUtils::rssiToBars(rssi);
            drawWiFiIcon(x, y, bars);
        }
    } else {
        display->print("Ethernet OK");
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

void DisplayManager::drawWiFiIcon(int16_t x, int16_t y, uint8_t bars) {
    const uint8_t* icon;

    // Select appropriate icon based on signal strength
    switch (bars) {
        case 0:
            icon = WIFI_ICON_0;
            break;
        case 1:
            icon = WIFI_ICON_1;
            break;
        case 2:
            icon = WIFI_ICON_2;
            break;
        case 3:
            icon = WIFI_ICON_3;
            break;
        case 4:
        default:
            icon = WIFI_ICON_4;
            break;
    }

    // Draw the bitmap (8x8 pixels)
    display->drawBitmap(x, y, icon, 8, 8, SSD1306_WHITE);
}

void DisplayManager::showOTAProgress(uint8_t percent) {
    if (!displayInitialized || display == nullptr) return;

    // Throttle updates (every 1 second)
    if (millis() - lastOTAUpdate < 1000 && percent != lastOTAPercent) {
        return;
    }

    otaInProgress = true;
    lastOTAPercent = percent;
    lastOTAUpdate = millis();

    display->clearDisplay();
    display->setTextSize(1);
    display->setTextColor(SSD1306_WHITE);

    // Title
    display->setCursor(10, 0);
    display->println("FIRMWARE UPDATE");

    // Progress bar (100px wide, 14px tall)
    int barX = 14;
    int barY = 20;
    int barWidth = 100;
    int barHeight = 14;

    display->drawRect(barX, barY, barWidth, barHeight, SSD1306_WHITE);
    int fillWidth = ((percent * (barWidth - 2)) / 100);
    display->fillRect(barX + 1, barY + 1, fillWidth, barHeight - 2, SSD1306_WHITE);

    // Percentage (large font)
    display->setTextSize(3);
    display->setCursor(38, 38);
    display->printf("%2d%%", percent);

    // Warning message
    display->setTextSize(1);
    display->setCursor(6, 56);
    display->println("DO NOT POWER OFF!");

    display->display();
}

void DisplayManager::showOTAComplete() {
    if (!displayInitialized || display == nullptr) return;

    display->clearDisplay();
    display->setTextSize(2);
    display->setTextColor(SSD1306_WHITE);

    // Success symbol
    display->setCursor(52, 10);
    display->println("OK");

    display->setTextSize(2);
    display->setCursor(16, 30);
    display->println("UPDATE");
    display->setCursor(10, 48);
    display->println("SUCCESS!");

    display->display();

    otaInProgress = false;
}

void DisplayManager::showOTAError(const char* error) {
    if (!displayInitialized || display == nullptr) return;

    display->clearDisplay();
    display->setTextSize(2);
    display->setTextColor(SSD1306_WHITE);

    // Error symbol
    display->setCursor(52, 10);
    display->println("X");

    display->setTextSize(1);
    display->setCursor(0, 30);
    display->println("UPDATE FAILED:");

    display->setCursor(0, 42);
    // Truncate error message to fit
    String errorMsg = String(error);
    if (errorMsg.length() > 21) {
        errorMsg = errorMsg.substring(0, 18) + "...";
    }
    display->println(errorMsg);

    display->display();

    otaInProgress = false;
}

void DisplayManager::showBootStep(const char* step) {
    if (!displayInitialized || display == nullptr) return;

    display->clearDisplay();
    display->setTextSize(1);
    display->setTextColor(SSD1306_WHITE);

    // Device name/version at top
    display->setCursor(0, 0);
    display->println(DEVICE_TYPE);
    display->setCursor(0, 10);
    display->println(FIRMWARE_VERSION);

    // Current step
    display->setCursor(0, 28);
    display->print("> ");
    display->println(step);

    display->display();
}

void DisplayManager::showBootCountdown(uint8_t seconds) {
    if (!displayInitialized || display == nullptr) return;

    display->clearDisplay();
    display->setTextSize(1);
    display->setTextColor(SSD1306_WHITE);

    display->setCursor(0, 0);
    display->println("Press BOOT for");
    display->setCursor(0, 10);
    display->println("  AP Mode Setup");

    // Countdown (large font)
    display->setTextSize(3);
    display->setCursor(52, 32);
    display->printf("%2d", seconds);

    display->setTextSize(1);
    display->setCursor(0, 56);
    display->println("or wait to continue");

    display->display();
}
