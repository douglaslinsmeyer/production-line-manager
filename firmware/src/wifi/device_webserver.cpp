#include "device_webserver.h"
#include "config.h"
#include "wifi_utils.h"
#include "network/connection_manager.h"
#include "ota/ota_manager.h"
#include "../profile/profile_manager.h"
#include "../profile/profile_storage.h"
#include "../state/line_state.h"
#include "../profile/output_controller.h"
#include <ArduinoJson.h>

extern DeviceConfig deviceConfig;
extern char deviceMAC[18];

DeviceWebServer::DeviceWebServer()
    : webServer(nullptr),
      connectionManager(nullptr),
      otaManager(nullptr),
      profileManager(nullptr),
      profileStorage(nullptr),
      lineState(nullptr),
      outputController(nullptr),
      running(false),
      serverPort(80),
      otaStartTime(0) {
}

DeviceWebServer::~DeviceWebServer() {
    stop();
}

bool DeviceWebServer::begin(uint16_t port) {
    serverPort = port;

    Serial.printf("Starting device web server on port %d...\n", serverPort);

    webServer = new WebServer(serverPort);

    // Register HTTP handlers
    webServer->on("/", [this]() { handleRoot(); });
    webServer->on("/config", [this]() { handleConfig(); });
    webServer->on("/wifi", [this]() { handleWiFiConfig(); });
    webServer->on("/ethernet", [this]() { handleEthernetConfig(); });
    webServer->on("/mqtt", [this]() { handleMQTTConfig(); });
    webServer->on("/device", [this]() { handleDeviceConfig(); });
    webServer->on("/save-wifi", HTTP_POST, [this]() { handleSaveWiFi(); });
    webServer->on("/save-ethernet", HTTP_POST, [this]() { handleSaveEthernet(); });
    webServer->on("/save-mqtt", HTTP_POST, [this]() { handleSaveMQTT(); });
    webServer->on("/save-device", HTTP_POST, [this]() { handleSaveDevice(); });
    webServer->on("/reboot", HTTP_POST, [this]() { handleReboot(); });
    webServer->on("/reset", HTTP_POST, [this]() { handleReset(); });
    webServer->on("/status", [this]() { handleStatus(); });

    // OTA update endpoints
    webServer->on("/update", HTTP_GET, [this]() { handleOTAPage(); });
    webServer->on("/update", HTTP_POST,
        [this]() { handleOTAUpload(); },        // Called when upload complete
        [this]() { handleOTAUploadData(); }     // Called for each chunk
    );
    webServer->on("/ota/progress", [this]() { handleOTAProgress(); });

    // Signal profile API endpoints
    webServer->on("/profile", HTTP_GET, [this]() { handleProfileView(); });
    webServer->on("/state", HTTP_GET, [this]() { handleStateView(); });
    webServer->on("/state/set", HTTP_POST, [this]() { handleStateSet(); });
    webServer->on("/override/clear", HTTP_POST, [this]() { handleOverrideClear(); });
    webServer->on("/outputs/test", HTTP_POST, [this]() { handleOutputTest(); });

    // Signal profile HTML pages
    webServer->on("/profile-config", [this]() { handleProfilePage(); });
    webServer->on("/state-control", [this]() { handleStateControlPage(); });
    webServer->on("/output-test", [this]() { handleOutputTestPage(); });

    webServer->onNotFound([this]() { handleNotFound(); });

    webServer->begin();
    running = true;

    Serial.printf("✓ Device web server started\n");
    Serial.printf("  Access configuration at: http://<device-ip>:%d\n", serverPort);

    return true;
}

void DeviceWebServer::stop() {
    if (webServer) {
        webServer->stop();
        delete webServer;
        webServer = nullptr;
        running = false;
        Serial.println("Device web server stopped");
    }
}

void DeviceWebServer::update() {
    if (webServer && running) {
        webServer->handleClient();
    }
}

void DeviceWebServer::setConnectionManager(ConnectionManager* manager) {
    connectionManager = manager;
}

void DeviceWebServer::setOTAManager(OTAManager* manager) {
    otaManager = manager;
}

void DeviceWebServer::setProfileComponents(ProfileManager* profileMgr, ProfileStorage* profileStore,
                                           LineStateManager* stateMgr, OutputController* outputCtrl) {
    profileManager = profileMgr;
    profileStorage = profileStore;
    lineState = stateMgr;
    outputController = outputCtrl;
}

// HTTP Handlers

void DeviceWebServer::handleRoot() {
    webServer->send(200, "text/html", generateHomePage());
}

void DeviceWebServer::handleConfig() {
    webServer->send(200, "text/html", generateConfigPage());
}

void DeviceWebServer::handleWiFiConfig() {
    webServer->send(200, "text/html", generateWiFiPage());
}

void DeviceWebServer::handleEthernetConfig() {
    webServer->send(200, "text/html", generateEthernetPage());
}

void DeviceWebServer::handleMQTTConfig() {
    webServer->send(200, "text/html", generateMQTTPage());
}

void DeviceWebServer::handleDeviceConfig() {
    webServer->send(200, "text/html", generateDevicePage());
}

void DeviceWebServer::handleSaveWiFi() {
    const DeviceConfig::Settings& currentSettings = deviceConfig.getSettings();

    if (webServer->hasArg("ssid") && webServer->hasArg("password")) {
        String ssid = webServer->arg("ssid");
        String password = webServer->arg("password");
        bool enabled = webServer->hasArg("enabled") && webServer->arg("enabled") == "on";

        if (deviceConfig.setWiFiCredentials(ssid.c_str(), password.c_str())) {
            deviceConfig.enableWiFi(enabled);
            deviceConfig.save();

            webServer->send(200, "application/json",
                           "{\"success\":true,\"message\":\"WiFi configuration saved. Reboot to apply.\"}");
        } else {
            webServer->send(400, "application/json",
                           "{\"success\":false,\"message\":\"Invalid WiFi configuration\"}");
        }
    } else {
        webServer->send(400, "application/json",
                       "{\"success\":false,\"message\":\"Missing required fields\"}");
    }
}

void DeviceWebServer::handleSaveEthernet() {
    if (webServer->hasArg("use_dhcp")) {
        bool useDHCP = webServer->arg("use_dhcp") == "true";

        if (useDHCP) {
            deviceConfig.setNetworkMode(true);
        } else if (webServer->hasArg("static_ip") && webServer->hasArg("gateway") &&
                   webServer->hasArg("subnet") && webServer->hasArg("dns")) {
            deviceConfig.setStaticIP(
                webServer->arg("static_ip").c_str(),
                webServer->arg("gateway").c_str(),
                webServer->arg("subnet").c_str(),
                webServer->arg("dns").c_str()
            );
        } else {
            webServer->send(400, "application/json",
                           "{\"success\":false,\"message\":\"Missing static IP configuration\"}");
            return;
        }

        deviceConfig.save();
        webServer->send(200, "application/json",
                       "{\"success\":true,\"message\":\"Ethernet configuration saved. Reboot to apply.\"}");
    } else {
        webServer->send(400, "application/json",
                       "{\"success\":false,\"message\":\"Missing required fields\"}");
    }
}

void DeviceWebServer::handleSaveMQTT() {
    if (webServer->hasArg("broker") && webServer->hasArg("port")) {
        String broker = webServer->arg("broker");
        uint16_t port = webServer->arg("port").toInt();
        String user = webServer->hasArg("user") ? webServer->arg("user") : "";
        String password = webServer->hasArg("password") ? webServer->arg("password") : "";
        bool enabled = webServer->hasArg("enabled") && webServer->arg("enabled") == "on";

        deviceConfig.enableMQTT(enabled);
        deviceConfig.setMQTTBroker(broker.c_str(), port);
        if (user.length() > 0) {
            deviceConfig.setMQTTAuth(user.c_str(), password.c_str());
        }
        deviceConfig.save();

        webServer->send(200, "application/json",
                       "{\"success\":true,\"message\":\"MQTT configuration saved. Reboot to apply.\"}");
    } else {
        webServer->send(400, "application/json",
                       "{\"success\":false,\"message\":\"Missing required fields\"}");
    }
}

void DeviceWebServer::handleSaveDevice() {
    if (webServer->hasArg("device_id")) {
        String deviceID = webServer->arg("device_id");

        deviceConfig.setDeviceID(deviceID.c_str());
        deviceConfig.save();

        webServer->send(200, "application/json",
                       "{\"success\":true,\"message\":\"Device configuration saved.\"}");
    } else {
        webServer->send(400, "application/json",
                       "{\"success\":false,\"message\":\"Missing required fields\"}");
    }
}

void DeviceWebServer::handleReboot() {
    webServer->send(200, "application/json",
                   "{\"success\":true,\"message\":\"Device rebooting in 3 seconds...\"}");

    Serial.println("Reboot requested via web interface");
    delay(3000);
    ESP.restart();
}

void DeviceWebServer::handleReset() {
    webServer->send(200, "application/json",
                   "{\"success\":true,\"message\":\"Configuration reset to defaults. Device will reboot.\"}");

    Serial.println("Factory reset requested via web interface");
    deviceConfig.resetToDefaults();
    delay(3000);
    ESP.restart();
}

void DeviceWebServer::handleStatus() {
    const DeviceConfig::Settings& settings = deviceConfig.getSettings();

    String json = "{";
    json += "\"device_id\":\"" + String(deviceMAC) + "\",";
    json += "\"uptime\":" + String(millis() / 1000) + ",";
    json += "\"free_heap\":" + String(ESP.getFreeHeap()) + ",";
    json += "\"connection_mode\":\"" + String(settings.connectionMode == MODE_WIFI ? "wifi" : "ethernet") + "\"";

    // Add WiFi signal info if in WiFi mode and connectionManager is available
    if (settings.connectionMode == MODE_WIFI && connectionManager) {
        int rssi = connectionManager->getRSSI();
        json += ",\"wifi_enabled\":" + String(settings.wifiEnabled ? "true" : "false");
        json += ",\"wifi_rssi\":" + String(rssi);
        json += ",\"wifi_signal_bars\":" + String(WiFiUtils::rssiToBars(rssi));
        json += ",\"wifi_signal_quality\":\"" + WiFiUtils::rssiToQuality(rssi) + "\"";
    } else {
        json += ",\"wifi_enabled\":" + String(settings.wifiEnabled ? "true" : "false");
    }

    json += "}";

    webServer->send(200, "application/json", json);
}

void DeviceWebServer::handleNotFound() {
    webServer->send(404, "text/plain", "404 Not Found");
}

// HTML Page Generators

String DeviceWebServer::getCSS() {
    return R"rawliteral(
<style>
    * { margin: 0; padding: 0; box-sizing: border-box; }
    body {
        font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Arial, sans-serif;
        background: #f5f5f5;
        color: #333;
    }
    .header {
        background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
        color: white;
        padding: 20px;
        box-shadow: 0 2px 10px rgba(0,0,0,0.1);
    }
    .header h1 { font-size: 24px; margin-bottom: 5px; }
    .header .subtitle { font-size: 14px; opacity: 0.9; }
    .container { max-width: 800px; margin: 20px auto; padding: 0 20px; }
    .nav {
        background: white;
        border-radius: 8px;
        padding: 15px;
        margin-bottom: 20px;
        box-shadow: 0 2px 8px rgba(0,0,0,0.1);
        display: flex;
        gap: 10px;
        flex-wrap: wrap;
    }
    .nav a {
        padding: 10px 20px;
        background: #667eea;
        color: white;
        text-decoration: none;
        border-radius: 6px;
        transition: background 0.3s;
        font-size: 14px;
    }
    .nav a:hover { background: #5568d3; }
    .nav a.active { background: #764ba2; }
    .card {
        background: white;
        border-radius: 8px;
        padding: 25px;
        box-shadow: 0 2px 8px rgba(0,0,0,0.1);
        margin-bottom: 20px;
    }
    .card h2 {
        font-size: 20px;
        margin-bottom: 20px;
        color: #667eea;
        border-bottom: 2px solid #f0f0f0;
        padding-bottom: 10px;
    }
    .form-group { margin-bottom: 20px; }
    .form-group label {
        display: block;
        margin-bottom: 8px;
        font-weight: 500;
        color: #555;
    }
    .form-group input[type="text"],
    .form-group input[type="password"],
    .form-group input[type="number"] {
        width: 100%;
        padding: 12px;
        border: 2px solid #ddd;
        border-radius: 6px;
        font-size: 14px;
        transition: border-color 0.3s;
    }
    .form-group input:focus {
        outline: none;
        border-color: #667eea;
    }
    .form-group input[type="checkbox"] {
        width: 20px;
        height: 20px;
        margin-right: 10px;
        cursor: pointer;
    }
    .checkbox-label {
        display: flex;
        align-items: center;
        cursor: pointer;
    }
    .btn {
        padding: 12px 24px;
        background: #667eea;
        color: white;
        border: none;
        border-radius: 6px;
        font-size: 14px;
        font-weight: 600;
        cursor: pointer;
        transition: background 0.3s;
        margin-right: 10px;
    }
    .btn:hover { background: #5568d3; }
    .btn-success { background: #48bb78; }
    .btn-success:hover { background: #38a169; }
    .btn-danger { background: #f56565; }
    .btn-danger:hover { background: #e53e3e; }
    .btn-secondary { background: #718096; }
    .btn-secondary:hover { background: #4a5568; }
    .info-box {
        background: #ebf8ff;
        border-left: 4px solid #4299e1;
        padding: 15px;
        border-radius: 4px;
        margin-bottom: 20px;
        font-size: 14px;
    }
    .warning-box {
        background: #fffaf0;
        border-left: 4px solid #ed8936;
        padding: 15px;
        border-radius: 4px;
        margin-bottom: 20px;
        font-size: 14px;
    }
    .status-badge {
        display: inline-block;
        padding: 4px 12px;
        border-radius: 12px;
        font-size: 12px;
        font-weight: 600;
        margin-left: 10px;
    }
    .status-badge.online { background: #c6f6d5; color: #22543d; }
    .status-badge.offline { background: #fed7d7; color: #742a2a; }
    .message {
        padding: 12px;
        border-radius: 6px;
        margin-bottom: 15px;
        display: none;
    }
    .message.success { background: #c6f6d5; color: #22543d; display: block; }
    .message.error { background: #fed7d7; color: #742a2a; display: block; }
    table {
        width: 100%;
        border-collapse: collapse;
        margin-top: 10px;
    }
    table th, table td {
        padding: 12px;
        text-align: left;
        border-bottom: 1px solid #eee;
    }
    table th {
        background: #f7fafc;
        font-weight: 600;
        color: #4a5568;
    }
    .signal-bars {
        font-size: 18px;
        font-weight: bold;
        margin-right: 10px;
        letter-spacing: 2px;
    }
    .signal-bars.signal-good { color: #48bb78; }
    .signal-bars.signal-fair { color: #ed8936; }
    .signal-bars.signal-poor { color: #f56565; }
    .signal-info {
        font-size: 14px;
        color: #718096;
    }
</style>
)rawliteral";
}

String DeviceWebServer::getHTMLHeader(const char* title) {
    String html = "<!DOCTYPE html><html><head>";
    html += "<meta charset='UTF-8'>";
    html += "<meta name='viewport' content='width=device-width, initial-scale=1.0'>";
    html += "<title>" + String(title) + " - ESP32 Configuration</title>";
    html += getCSS();
    html += "</head><body>";
    html += "<div class='header'>";
    html += "<h1>ESP32-S3 Device Configuration</h1>";
    html += "<div class='subtitle'>MAC: " + String(deviceMAC) + " | Firmware: " + String(FIRMWARE_VERSION) + "</div>";
    html += "</div>";
    return html;
}

String DeviceWebServer::getHTMLFooter() {
    return "</body></html>";
}

String DeviceWebServer::getNavigation() {
    return R"rawliteral(
<div class='nav'>
    <a href='/'>Home</a>
    <a href='/wifi'>WiFi</a>
    <a href='/ethernet'>Ethernet</a>
    <a href='/mqtt'>MQTT</a>
    <a href='/device'>Device</a>
    <a href='/profile-config'>Profile</a>
    <a href='/state-control'>State</a>
    <a href='/output-test'>Test</a>
</div>
)rawliteral";
}

String DeviceWebServer::generateHomePage() {
    const DeviceConfig::Settings& settings = deviceConfig.getSettings();

    String html = getHTMLHeader("Home");
    html += "<div class='container'>";
    html += getNavigation();

    html += "<div class='card'>";
    html += "<h2>Device Overview</h2>";
    html += "<table>";
    html += "<tr><th>Property</th><th>Value</th></tr>";
    html += "<tr><td>Device ID</td><td>" + String(settings.deviceID) + "</td></tr>";
    html += "<tr><td>MAC Address</td><td>" + String(deviceMAC) + "</td></tr>";
    html += "<tr><td>Device Type</td><td>" + String(DEVICE_TYPE) + "</td></tr>";
    html += "<tr><td>Firmware Version</td><td>" + String(FIRMWARE_VERSION) + "</td></tr>";
    html += "<tr><td>Uptime</td><td>" + String(millis() / 1000) + " seconds</td></tr>";
    html += "<tr><td>Free Heap</td><td>" + String(ESP.getFreeHeap()) + " bytes</td></tr>";
    html += "</table>";
    html += "</div>";

    html += "<div class='card'>";
    html += "<h2>Network Status</h2>";
    html += "<table>";
    html += "<tr><th>Setting</th><th>Value</th></tr>";
    html += "<tr><td>Connection Mode</td><td>" + String(settings.connectionMode == MODE_WIFI ? "WiFi" : "Ethernet") + "</td></tr>";

    if (settings.connectionMode == MODE_WIFI) {
        html += "<tr><td>WiFi SSID</td><td>" + String(settings.wifiSSID) + "</td></tr>";
        html += "<tr><td>WiFi Enabled</td><td>" + String(settings.wifiEnabled ? "Yes" : "No") + "</td></tr>";

        // Add signal strength if WiFi is connected
        if (connectionManager && connectionManager->isConnected()) {
            int rssi = connectionManager->getRSSI();
            String bars = WiFiUtils::rssiToBarString(rssi);
            String quality = WiFiUtils::rssiToQuality(rssi);

            // Determine color class based on signal strength
            String colorClass = "signal-good";
            if (rssi < -70) colorClass = "signal-poor";
            else if (rssi < -60) colorClass = "signal-fair";

            html += "<tr><td>WiFi Signal</td><td>";
            html += "<span class='signal-bars " + colorClass + "'>" + bars + "</span>";
            html += "<span class='signal-info'>" + String(rssi) + " dBm (" + quality + ")</span>";
            html += "</td></tr>";
        }
    } else {
        html += "<tr><td>Network Mode</td><td>" + String(settings.useDHCP ? "DHCP" : "Static IP") + "</td></tr>";
        if (!settings.useDHCP) {
            html += "<tr><td>Static IP</td><td>" + String(settings.staticIP) + "</td></tr>";
            html += "<tr><td>Gateway</td><td>" + String(settings.gateway) + "</td></tr>";
        }
    }

    html += "</table>";
    html += "</div>";

    html += "<div class='card'>";
    html += "<h2>Quick Actions</h2>";
    html += "<button class='btn btn-secondary' onclick='location.href=\"/config\"'>Full Configuration</button>";
    html += "<button class='btn btn-danger' onclick='if(confirm(\"Reboot device?\")) rebootDevice()'>Reboot Device</button>";
    html += "</div>";

    html += "</div>";

    html += "<script>";
    html += "function rebootDevice() {";
    html += "  fetch('/reboot', {method: 'POST'})";
    html += "    .then(r => r.json())";
    html += "    .then(d => alert(d.message));";
    html += "}";
    html += "</script>";

    html += getHTMLFooter();
    return html;
}

String DeviceWebServer::generateConfigPage() {
    String html = getHTMLHeader("Configuration");
    html += "<div class='container'>";
    html += getNavigation();

    html += "<div class='card'>";
    html += "<h2>Configuration Menu</h2>";
    html += "<p>Select a category to configure:</p>";
    html += "<div style='margin-top: 20px;'>";
    html += "<a href='/wifi'><button class='btn' style='width: 100%; margin-bottom: 10px;'>WiFi Configuration</button></a>";
    html += "<a href='/ethernet'><button class='btn' style='width: 100%; margin-bottom: 10px;'>Ethernet Configuration</button></a>";
    html += "<a href='/mqtt'><button class='btn' style='width: 100%; margin-bottom: 10px;'>MQTT Configuration</button></a>";
    html += "<a href='/device'><button class='btn' style='width: 100%; margin-bottom: 10px;'>Device Information</button></a>";
    html += "</div>";
    html += "</div>";

    html += "<div class='card'>";
    html += "<h2>System Actions</h2>";
    html += "<button class='btn btn-secondary' onclick='if(confirm(\"Reboot device?\")) rebootDevice()'>Reboot Device</button>";
    html += "<button class='btn btn-danger' onclick='if(confirm(\"Reset to factory defaults?\")) resetDevice()'>Factory Reset</button>";
    html += "</div>";

    html += "</div>";

    html += "<script>";
    html += "function rebootDevice() {";
    html += "  fetch('/reboot', {method: 'POST'}).then(r => r.json()).then(d => alert(d.message));";
    html += "}";
    html += "function resetDevice() {";
    html += "  fetch('/reset', {method: 'POST'}).then(r => r.json()).then(d => alert(d.message));";
    html += "}";
    html += "</script>";

    html += getHTMLFooter();
    return html;
}

String DeviceWebServer::generateWiFiPage() {
    const DeviceConfig::Settings& settings = deviceConfig.getSettings();

    String html = getHTMLHeader("WiFi Configuration");
    html += "<div class='container'>";
    html += getNavigation();

    html += "<div class='info-box'>";
    html += "Configure WiFi settings. Device will need to reboot to apply changes.";
    html += "</div>";

    html += "<div id='message' class='message'></div>";

    html += "<div class='card'>";
    html += "<h2>WiFi Configuration</h2>";
    html += "<form id='wifiForm' onsubmit='saveWiFi(event)'>";

    html += "<div class='form-group'>";
    html += "<label class='checkbox-label'>";
    html += "<input type='checkbox' name='enabled' " + String(settings.wifiEnabled ? "checked" : "") + ">";
    html += " Enable WiFi";
    html += "</label>";
    html += "</div>";

    html += "<div class='form-group'>";
    html += "<label>Network SSID:</label>";
    html += "<input type='text' name='ssid' value='" + String(settings.wifiSSID) + "' maxlength='32' required>";
    html += "</div>";

    html += "<div class='form-group'>";
    html += "<label>Password:</label>";
    html += "<input type='password' name='password' placeholder='Leave empty to keep current' maxlength='63'>";
    html += "<small style='color: #888;'>Min 8 characters for WPA2, or empty for open networks</small>";
    html += "</div>";

    html += "<button type='submit' class='btn btn-success'>Save WiFi Configuration</button>";
    html += "<button type='button' class='btn btn-secondary' onclick='location.href=\"/\"'>Cancel</button>";
    html += "</form>";
    html += "</div>";

    html += "<div class='card'>";
    html += "<h2>Current Status</h2>";
    html += "<p><strong>Connection Mode:</strong> " + String(settings.connectionMode == MODE_WIFI ? "WiFi" : "Ethernet") + "</p>";
    html += "<p><strong>WiFi Enabled:</strong> " + String(settings.wifiEnabled ? "Yes" : "No") + "</p>";
    if (strlen(settings.wifiSSID) > 0) {
        html += "<p><strong>Configured SSID:</strong> " + String(settings.wifiSSID) + "</p>";
    }
    html += "</div>";

    html += "</div>";

    html += "<script>";
    html += "function saveWiFi(e) {";
    html += "  e.preventDefault();";
    html += "  const form = e.target;";
    html += "  const data = new URLSearchParams(new FormData(form));";
    html += "  fetch('/save-wifi', {method: 'POST', body: data})";
    html += "    .then(r => r.json())";
    html += "    .then(d => {";
    html += "      const msg = document.getElementById('message');";
    html += "      msg.textContent = d.message;";
    html += "      msg.className = 'message ' + (d.success ? 'success' : 'error');";
    html += "      if (d.success) setTimeout(() => location.href='/', 2000);";
    html += "    });";
    html += "}";
    html += "</script>";

    html += getHTMLFooter();
    return html;
}

String DeviceWebServer::generateEthernetPage() {
    const DeviceConfig::Settings& settings = deviceConfig.getSettings();

    String html = getHTMLHeader("Ethernet Configuration");
    html += "<div class='container'>";
    html += getNavigation();

    html += "<div class='info-box'>";
    html += "Configure Ethernet network settings. Device will need to reboot to apply changes.";
    html += "</div>";

    html += "<div id='message' class='message'></div>";

    html += "<div class='card'>";
    html += "<h2>Ethernet Configuration</h2>";
    html += "<form id='ethForm' onsubmit='saveEthernet(event)'>";

    html += "<div class='form-group'>";
    html += "<label class='checkbox-label'>";
    html += "<input type='checkbox' name='use_dhcp' id='useDHCP' " + String(settings.useDHCP ? "checked" : "") + " onchange='toggleStaticIP()'>";
    html += " Use DHCP (automatic IP)";
    html += "</label>";
    html += "</div>";

    html += "<div id='staticIPFields' style='display: " + String(settings.useDHCP ? "none" : "block") + ";'>";

    html += "<div class='form-group'>";
    html += "<label>Static IP Address:</label>";
    html += "<input type='text' name='static_ip' value='" + String(settings.staticIP) + "' placeholder='192.168.1.100'>";
    html += "</div>";

    html += "<div class='form-group'>";
    html += "<label>Gateway:</label>";
    html += "<input type='text' name='gateway' value='" + String(settings.gateway) + "' placeholder='192.168.1.1'>";
    html += "</div>";

    html += "<div class='form-group'>";
    html += "<label>Subnet Mask:</label>";
    html += "<input type='text' name='subnet' value='" + String(settings.subnet) + "' placeholder='255.255.255.0'>";
    html += "</div>";

    html += "<div class='form-group'>";
    html += "<label>DNS Server:</label>";
    html += "<input type='text' name='dns' value='" + String(settings.dnsServer) + "' placeholder='8.8.8.8'>";
    html += "</div>";

    html += "</div>";

    html += "<button type='submit' class='btn btn-success'>Save Ethernet Configuration</button>";
    html += "<button type='button' class='btn btn-secondary' onclick='location.href=\"/\"'>Cancel</button>";
    html += "</form>";
    html += "</div>";

    html += "</div>";

    html += "<script>";
    html += "function toggleStaticIP() {";
    html += "  const checked = document.getElementById('useDHCP').checked;";
    html += "  document.getElementById('staticIPFields').style.display = checked ? 'none' : 'block';";
    html += "}";
    html += "function saveEthernet(e) {";
    html += "  e.preventDefault();";
    html += "  const form = e.target;";
    html += "  const formData = new FormData(form);";
    html += "  const data = new URLSearchParams();";
    html += "  data.append('use_dhcp', formData.get('use_dhcp') ? 'true' : 'false');";
    html += "  if (formData.get('use_dhcp') !== 'on') {";
    html += "    data.append('static_ip', formData.get('static_ip'));";
    html += "    data.append('gateway', formData.get('gateway'));";
    html += "    data.append('subnet', formData.get('subnet'));";
    html += "    data.append('dns', formData.get('dns'));";
    html += "  }";
    html += "  fetch('/save-ethernet', {method: 'POST', body: data})";
    html += "    .then(r => r.json())";
    html += "    .then(d => {";
    html += "      const msg = document.getElementById('message');";
    html += "      msg.textContent = d.message;";
    html += "      msg.className = 'message ' + (d.success ? 'success' : 'error');";
    html += "    });";
    html += "}";
    html += "</script>";

    html += getHTMLFooter();
    return html;
}

String DeviceWebServer::generateMQTTPage() {
    const DeviceConfig::Settings& settings = deviceConfig.getSettings();

    String html = getHTMLHeader("MQTT Configuration");
    html += "<div class='container'>";
    html += getNavigation();

    html += "<div class='info-box'>";
    html += "Configure MQTT broker connection. Device will need to reboot to apply changes.";
    html += "</div>";

    html += "<div id='message' class='message'></div>";

    html += "<div class='card'>";
    html += "<h2>MQTT Broker Configuration</h2>";
    html += "<form id='mqttForm' onsubmit='saveMQTT(event)'>";

    html += "<div class='form-group'>";
    html += "<label>";
    html += "<input type='checkbox' name='enabled' " + String(settings.mqttEnabled ? "checked" : "") + "> Enable MQTT";
    html += "</label>";
    html += "<small style='display:block;margin-top:5px;color:#666;'>Turn MQTT functionality on or off</small>";
    html += "</div>";

    html += "<div class='form-group'>";
    html += "<label>Broker Address:</label>";
    html += "<input type='text' name='broker' value='" + String(settings.mqttBroker) + "' required placeholder='10.221.21.100'>";
    html += "</div>";

    html += "<div class='form-group'>";
    html += "<label>Port:</label>";
    html += "<input type='number' name='port' value='" + String(settings.mqttPort) + "' required placeholder='1883'>";
    html += "</div>";

    html += "<div class='form-group'>";
    html += "<label>Username (optional):</label>";
    html += "<input type='text' name='user' value='" + String(settings.mqttUser) + "' placeholder='Leave empty if not required'>";
    html += "</div>";

    html += "<div class='form-group'>";
    html += "<label>Password (optional):</label>";
    html += "<input type='password' name='password' placeholder='Leave empty to keep current or if not required'>";
    html += "</div>";

    html += "<button type='submit' class='btn btn-success'>Save MQTT Configuration</button>";
    html += "<button type='button' class='btn btn-secondary' onclick='location.href=\"/\"'>Cancel</button>";
    html += "</form>";
    html += "</div>";

    html += "</div>";

    html += "<script>";
    html += "function saveMQTT(e) {";
    html += "  e.preventDefault();";
    html += "  const data = new URLSearchParams(new FormData(e.target));";
    html += "  fetch('/save-mqtt', {method: 'POST', body: data})";
    html += "    .then(r => r.json())";
    html += "    .then(d => {";
    html += "      const msg = document.getElementById('message');";
    html += "      msg.textContent = d.message;";
    html += "      msg.className = 'message ' + (d.success ? 'success' : 'error');";
    html += "    });";
    html += "}";
    html += "</script>";

    html += getHTMLFooter();
    return html;
}

String DeviceWebServer::generateDevicePage() {
    const DeviceConfig::Settings& settings = deviceConfig.getSettings();

    String html = getHTMLHeader("Device Information");
    html += "<div class='container'>";
    html += getNavigation();

    html += "<div id='message' class='message'></div>";

    html += "<div class='card'>";
    html += "<h2>Device Information</h2>";
    html += "<form id='deviceForm' onsubmit='saveDevice(event)'>";

    html += "<div class='form-group'>";
    html += "<label>Device ID:</label>";
    html += "<input type='text' name='device_id' value='" + String(settings.deviceID) + "' required maxlength='32'>";
    html += "</div>";

    html += "<button type='submit' class='btn btn-success'>Save Device Configuration</button>";
    html += "<button type='button' class='btn btn-secondary' onclick='location.href=\"/\"'>Cancel</button>";
    html += "</form>";
    html += "</div>";

    html += "<div class='card'>";
    html += "<h2>Hardware Information</h2>";
    html += "<table>";
    html += "<tr><th>Property</th><th>Value</th></tr>";
    html += "<tr><td>MAC Address</td><td>" + String(deviceMAC) + "</td></tr>";
    html += "<tr><td>Chip Model</td><td>" + String(ESP.getChipModel()) + "</td></tr>";
    html += "<tr><td>CPU Frequency</td><td>" + String(ESP.getCpuFreqMHz()) + " MHz</td></tr>";
    html += "<tr><td>Flash Size</td><td>" + String(ESP.getFlashChipSize()) + " bytes</td></tr>";
    html += "<tr><td>PSRAM Size</td><td>" + String(ESP.getPsramSize()) + " bytes</td></tr>";
    html += "</table>";
    html += "</div>";

    html += "</div>";

    html += "<script>";
    html += "function saveDevice(e) {";
    html += "  e.preventDefault();";
    html += "  const data = new URLSearchParams(new FormData(e.target));";
    html += "  fetch('/save-device', {method: 'POST', body: data})";
    html += "    .then(r => r.json())";
    html += "    .then(d => {";
    html += "      const msg = document.getElementById('message');";
    html += "      msg.textContent = d.message;";
    html += "      msg.className = 'message ' + (d.success ? 'success' : 'error');";
    html += "    });";
    html += "}";
    html += "</script>";

    html += getHTMLFooter();
    return html;
}

// =============================================================================
// OTA Update Handlers
// =============================================================================

void DeviceWebServer::handleOTAPage() {
    webServer->send(200, "text/html", generateOTAPage());
}

void DeviceWebServer::handleOTAUploadData() {
    if (!otaManager) {
        webServer->send(500, "application/json",
            "{\"success\":false,\"message\":\"OTA manager not initialized\"}");
        return;
    }

    HTTPUpload& upload = webServer->upload();
    static unsigned long lastChunkTime = 0;

    if (upload.status == UPLOAD_FILE_START) {
        Serial.printf("OTA: Starting upload: %s\n", upload.filename.c_str());
        otaStartTime = millis();
        lastChunkTime = millis();

        // Get content length from HTTP header
        size_t contentLength = webServer->header("Content-Length").toInt();
        if (contentLength == 0) {
            Serial.println("OTA: ERROR - No content length in request");
            return;
        }

        // Get MD5 from form data if provided
        String md5 = webServer->hasArg("md5") ? webServer->arg("md5") : "";

        if (!otaManager->startUpdate(contentLength,
                                     md5.length() == 32 ? md5.c_str() : nullptr)) {
            Serial.printf("OTA: Start failed: %s\n", otaManager->getErrorString());
            return;
        }
    }
    else if (upload.status == UPLOAD_FILE_WRITE) {
        // Check chunk timeout
        if (millis() - lastChunkTime > 30000) {  // 30s between chunks
            Serial.println("OTA: Chunk timeout - aborting");
            otaManager->abortUpdate();
            return;
        }

        // Check total timeout
        if (millis() - otaStartTime > OTA_TIMEOUT_MS) {
            Serial.println("OTA: Total timeout - aborting");
            otaManager->abortUpdate();
            return;
        }

        lastChunkTime = millis();

        // Write chunk
        if (!otaManager->writeChunk(upload.buf, upload.currentSize)) {
            Serial.printf("OTA: Write failed at %u bytes\n",
                         otaManager->getBytesWritten());
            otaManager->abortUpdate();
            return;
        }
    }
    else if (upload.status == UPLOAD_FILE_END) {
        if (otaManager->finishUpdate()) {
            Serial.println("OTA: Update successful!");
        } else {
            Serial.printf("OTA: Finish failed: %s\n",
                         otaManager->getErrorString());
        }
    }
    else if (upload.status == UPLOAD_FILE_ABORTED) {
        Serial.println("OTA: Upload aborted by client");
        otaManager->abortUpdate();
    }
}

void DeviceWebServer::handleOTAUpload() {
    if (!otaManager) {
        webServer->send(500, "application/json",
            "{\"success\":false,\"message\":\"OTA manager not initialized\"}");
        return;
    }

    // Called after upload completes
    OTAManager::OTAState state = otaManager->getState();

    if (state == OTAManager::OTA_SUCCESS) {
        webServer->send(200, "application/json",
            "{\"success\":true,\"message\":\"Update successful! Rebooting in 5 seconds...\"}");

        delay(5000);
        ESP.restart();
    } else {
        String json = "{\"success\":false,\"message\":\"";
        json += otaManager->getErrorString();
        json += "\"}";
        webServer->send(500, "application/json", json);
    }
}

void DeviceWebServer::handleOTAProgress() {
    if (!otaManager) {
        webServer->send(500, "application/json",
            "{\"error\":\"OTA not initialized\"}");
        return;
    }

    OTAManager::OTAState state = otaManager->getState();
    uint8_t progress = otaManager->getProgressPercent();
    size_t bytesWritten = otaManager->getBytesWritten();
    size_t totalSize = otaManager->getTotalSize();

    // Calculate upload speed
    unsigned long elapsed = (millis() - otaStartTime) / 1000;  // seconds
    float speedKBps = elapsed > 0 ? (bytesWritten / 1024.0) / elapsed : 0;

    // Calculate ETA
    size_t remaining = totalSize - bytesWritten;
    unsigned long etaSeconds = speedKBps > 0 ? (remaining / 1024) / speedKBps : 0;

    // Build JSON response
    String json = "{";
    json += "\"state\":\"" + otaManager->getStateString() + "\",";
    json += "\"progress\":" + String(progress) + ",";
    json += "\"bytesWritten\":" + String(bytesWritten) + ",";
    json += "\"totalSize\":" + String(totalSize) + ",";
    json += "\"speedKBps\":" + String(speedKBps, 2) + ",";
    json += "\"etaSeconds\":" + String(etaSeconds);

    if (state == OTAManager::OTA_ERROR) {
        json += ",\"error\":\"" + String(otaManager->getErrorString()) + "\"";
    }

    json += "}";

    webServer->send(200, "application/json", json);
}

String DeviceWebServer::generateOTAPage() {
    uint32_t bootCount = otaManager ? otaManager->getBootCount() : 0;

    String html = getHTMLHeader("Firmware Update");
    html += "<div class='container'>";
    html += getNavigation();

    // Show warning if boot count is high
    if (bootCount >= OTA_BOOT_COUNT_THRESHOLD) {
        html += "<div style='background:#fed7d7; border-left:4px solid #f56565; padding:15px; margin-bottom:20px; border-radius:4px'>";
        html += "⚠️ <strong>WARNING:</strong> Multiple boot failures detected!<br>";
        html += "Boot count: " + String(bootCount) + " (threshold: " + String(OTA_BOOT_COUNT_THRESHOLD) + ")<br>";
        html += "Recent firmware may be unstable. Consider rolling back.";
        html += "</div>";
    }

    // Warning box
    html += "<div style='background:#fef3c7; border-left:4px solid #f59e0b; padding:15px; margin-bottom:20px; border-radius:4px'>";
    html += "⚠️ <strong>WARNING:</strong> Do not disconnect power during update!<br>";
    html += "Current firmware: " + String(FIRMWARE_VERSION) + "<br>";
    html += "Update will take approximately 15-30 seconds.";
    html += "</div>";

    // System info card
    html += "<div class='card'>";
    html += "<h2>Current Firmware</h2>";
    html += "<table>";
    html += "<tr><td><strong>Version</strong></td><td>" + String(FIRMWARE_VERSION) + "</td></tr>";
    html += "<tr><td><strong>Device Type</strong></td><td>" + String(DEVICE_TYPE) + "</td></tr>";
    html += "<tr><td><strong>Boot Count</strong></td><td>" + String(bootCount) + "</td></tr>";
    html += "<tr><td><strong>Uptime</strong></td><td>" + String(millis() / 1000) + " seconds</td></tr>";
    html += "</table>";
    html += "</div>";

    // Upload form
    html += "<div class='card'>";
    html += "<h2>Upload New Firmware</h2>";
    html += "<form id='uploadForm' onsubmit='return false;'>";

    html += "<div style='margin-bottom:15px'>";
    html += "<label style='display:block; margin-bottom:5px; font-weight:600'>Firmware File (.bin):</label>";
    html += "<input type='file' id='firmwareFile' accept='.bin' required style='width:100%; padding:8px; border:1px solid #e2e8f0; border-radius:4px'>";
    html += "</div>";

    html += "<div style='margin-bottom:15px'>";
    html += "<label style='display:block; margin-bottom:5px; font-weight:600'>MD5 Checksum (optional but recommended):</label>";
    html += "<input type='text' id='md5sum' name='md5' placeholder='32-character hexadecimal hash' ";
    html += "pattern='[a-fA-F0-9]{32}' maxlength='32' style='width:100%; padding:8px; border:1px solid #e2e8f0; border-radius:4px'>";
    html += "<small style='color:#718096; display:block; margin-top:5px'>";
    html += "Verifies firmware integrity. Generate with: <code>md5sum firmware.bin</code>";
    html += "</small>";
    html += "</div>";

    html += "<button type='button' class='btn btn-success' onclick='uploadFirmware()' style='width:100%'>Start Update</button>";
    html += "</form>";
    html += "</div>";

    // Progress card (hidden initially)
    html += "<div id='progressCard' class='card' style='display:none'>";
    html += "<h2>Update Progress</h2>";

    // Progress bar
    html += "<div style='background:#eee; height:40px; border-radius:8px; overflow:hidden; margin-bottom:15px'>";
    html += "<div id='progressBar' style='background:linear-gradient(90deg, #667eea, #764ba2); height:100%; width:0%; transition:width 0.5s ease'></div>";
    html += "</div>";

    // Percentage
    html += "<div id='progressText' style='text-align:center; font-size:32px; font-weight:bold; color:#667eea; margin-bottom:20px'>0%</div>";

    // Statistics
    html += "<div style='display:grid; grid-template-columns:1fr 1fr; gap:15px; margin-bottom:20px'>";
    html += "<div style='text-align:center; padding:15px; background:#f7fafc; border-radius:6px'>";
    html += "<div style='font-size:12px; color:#718096; margin-bottom:5px'>Upload Speed</div>";
    html += "<div id='speed' style='font-size:20px; font-weight:bold; color:#667eea'>-- KB/s</div>";
    html += "</div>";
    html += "<div style='text-align:center; padding:15px; background:#f7fafc; border-radius:6px'>";
    html += "<div style='font-size:12px; color:#718096; margin-bottom:5px'>Time Remaining</div>";
    html += "<div id='eta' style='font-size:20px; font-weight:bold; color:#667eea'>--:--</div>";
    html += "</div>";
    html += "</div>";

    // Bytes progress
    html += "<div style='text-align:center; font-size:14px; color:#718096; margin-bottom:15px'>";
    html += "<span id='bytesWritten'>0</span> / <span id='totalBytes'>0</span> bytes";
    html += "</div>";

    // Status message
    html += "<div id='statusText' style='text-align:center; padding:15px; border-radius:6px; background:#ebf8ff; color:#2c5282'></div>";
    html += "</div>";

    // JavaScript
    html += "<script>";
    html += "let progressInterval = null;";

    html += "function uploadFirmware() {";
    html += "  const fileInput = document.getElementById('firmwareFile');";
    html += "  const md5Input = document.getElementById('md5sum');";
    html += "  if (!fileInput.files.length) { alert('Please select a firmware file'); return; }";

    html += "  document.getElementById('uploadForm').style.display = 'none';";
    html += "  document.getElementById('progressCard').style.display = 'block';";
    html += "  document.getElementById('statusText').innerText = 'Starting upload...';";

    html += "  const formData = new FormData();";
    html += "  formData.append('firmware', fileInput.files[0]);";
    html += "  if (md5Input.value.length === 32) {";
    html += "    formData.append('md5', md5Input.value.toLowerCase());";
    html += "  }";

    html += "  progressInterval = setInterval(updateProgress, 500);";

    html += "  fetch('/update', { method: 'POST', body: formData })";
    html += "    .then(r => r.json())";
    html += "    .then(d => {";
    html += "      clearInterval(progressInterval);";
    html += "      const statusEl = document.getElementById('statusText');";
    html += "      if (d.success) {";
    html += "        statusEl.style.background = '#c6f6d5';";
    html += "        statusEl.style.color = '#22543d';";
    html += "        statusEl.innerHTML = '✓ ' + d.message;";
    html += "      } else {";
    html += "        statusEl.style.background = '#fed7d7';";
    html += "        statusEl.style.color = '#742a2a';";
    html += "        statusEl.innerHTML = '✗ Error: ' + d.message;";
    html += "      }";
    html += "    })";
    html += "    .catch(e => {";
    html += "      clearInterval(progressInterval);";
    html += "      const statusEl = document.getElementById('statusText');";
    html += "      statusEl.style.background = '#fed7d7';";
    html += "      statusEl.style.color = '#742a2a';";
    html += "      statusEl.innerHTML = '✗ Upload failed: ' + e;";
    html += "    });";
    html += "}";

    html += "function updateProgress() {";
    html += "  fetch('/ota/progress')";
    html += "    .then(r => r.json())";
    html += "    .then(d => {";
    html += "      document.getElementById('progressBar').style.width = d.progress + '%';";
    html += "      document.getElementById('progressText').textContent = d.progress + '%';";
    html += "      document.getElementById('speed').textContent = d.speedKBps.toFixed(1) + ' KB/s';";
    html += "      const mins = Math.floor(d.etaSeconds / 60);";
    html += "      const secs = d.etaSeconds % 60;";
    html += "      document.getElementById('eta').textContent = mins + ':' + (secs < 10 ? '0' : '') + secs;";
    html += "      document.getElementById('bytesWritten').textContent = (d.bytesWritten / 1024).toFixed(1) + ' KB';";
    html += "      document.getElementById('totalBytes').textContent = (d.totalSize / 1024).toFixed(1) + ' KB';";
    html += "      if (d.state === 'in_progress') {";
    html += "        document.getElementById('statusText').innerText = 'Uploading firmware...';";
    html += "      } else if (d.state === 'success') {";
    html += "        document.getElementById('statusText').innerText = 'Finalizing update...';";
    html += "      } else if (d.state === 'error') {";
    html += "        clearInterval(progressInterval);";
    html += "        const statusEl = document.getElementById('statusText');";
    html += "        statusEl.style.background = '#fed7d7';";
    html += "        statusEl.style.color = '#742a2a';";
    html += "        statusEl.innerText = '✗ Error: ' + d.error;";
    html += "      }";
    html += "    })";
    html += "    .catch(e => console.error('Progress fetch error:', e));";
    html += "}";
    html += "</script>";

    html += "</div>";
    html += getHTMLFooter();
    return html;
}

// ============================================================================
// Signal Profile API Handlers
// ============================================================================

void DeviceWebServer::handleProfileView() {
    if (!profileManager || !profileManager->hasProfile()) {
        webServer->send(404, "application/json", "{\"error\":\"No profile loaded\"}");
        return;
    }

    String profileJson;
    if (profileStorage->loadProfile(profileJson)) {
        webServer->send(200, "application/json", profileJson);
    } else {
        webServer->send(500, "application/json", "{\"error\":\"Failed to load profile\"}");
    }
}

void DeviceWebServer::handleStateView() {
    String json = "{";
    json += "\"currentState\":\"" + String(lineState->getState()) + "\",";
    json += "\"isOverridden\":" + String(profileStorage->getOverrideFlag() ? "true" : "false") + ",";
    json += "\"profileId\":\"" + profileStorage->getProfileId() + "\",";
    json += "\"profileVersion\":" + String(profileStorage->getProfileVersion());

    // Add available states if profile loaded
    if (profileManager && profileManager->hasProfile()) {
        json += ",\"availableStates\":[";

        String profileJson;
        profileStorage->loadProfile(profileJson);
        JsonDocument doc;
        deserializeJson(doc, profileJson);

        JsonArrayConst states = doc["states"];
        bool first = true;
        for (JsonObjectConst state : states) {
            if (!first) json += ",";
            json += "\"" + String(state["name"].as<const char*>()) + "\"";
            first = false;
        }
        json += "]";
    }

    json += "}";
    webServer->send(200, "application/json", json);
}

void DeviceWebServer::handleStateSet() {
    if (!webServer->hasArg("state")) {
        webServer->send(400, "application/json",
                       "{\"success\":false,\"error\":\"Missing 'state' parameter\"}");
        return;
    }

    String stateName = webServer->arg("state");

    if (!lineState || !profileManager->isValidState(stateName.c_str())) {
        webServer->send(400, "application/json",
                       "{\"success\":false,\"error\":\"Invalid state name\"}");
        return;
    }

    // Set state
    lineState->setState(stateName.c_str(), "webui");
    profileStorage->setCurrentState(stateName.c_str());
    profileStorage->setOverrideFlag(true);

    // Apply outputs
    if (outputController) {
        outputController->applyStateOutputs(stateName.c_str());
    }

    webServer->send(200, "application/json",
                   "{\"success\":true,\"state\":\"" + stateName + "\"}");
}

void DeviceWebServer::handleOverrideClear() {
    profileStorage->setOverrideFlag(false);

    // Return to default state
    if (profileManager && profileManager->hasProfile()) {
        const char* defaultState = profileManager->getDefaultState();
        lineState->setState(defaultState, "webui_clear");
        profileStorage->setCurrentState(defaultState);

        if (outputController) {
            outputController->applyStateOutputs(defaultState);
        }

        webServer->send(200, "application/json",
                       "{\"success\":true,\"state\":\"" + String(defaultState) + "\"}");
    } else {
        webServer->send(200, "application/json",
                       "{\"success\":true,\"state\":\"unknown\"}");
    }
}

void DeviceWebServer::handleOutputTest() {
    // Parse JSON body
    JsonDocument doc;
    DeserializationError error = deserializeJson(doc, webServer->arg("plain"));

    if (error) {
        webServer->send(400, "application/json",
                       "{\"success\":false,\"error\":\"Invalid JSON\"}");
        return;
    }

    bool red = doc["red"] | false;
    bool yellow = doc["yellow"] | false;
    bool green = doc["green"] | false;
    // Support both old "buzzer" field and new separate buzzer fields
    bool primaryBuzzer = doc["primaryBuzzer"] | doc["buzzer"] | false;
    bool towerBuzzer = doc["towerBuzzer"] | doc["buzzer"] | false;

    if (outputController) {
        outputController->testOutputs(red, yellow, green, primaryBuzzer, towerBuzzer);
        webServer->send(200, "application/json", "{\"success\":true}");
    } else {
        webServer->send(500, "application/json",
                       "{\"success\":false,\"error\":\"Output controller not initialized\"}");
    }
}

// ============================================================================
// Signal Profile HTML Page Handlers
// ============================================================================

void DeviceWebServer::handleProfilePage() {
    webServer->send(200, "text/html", generateProfilePage());
}

void DeviceWebServer::handleStateControlPage() {
    webServer->send(200, "text/html", generateStateControlPage());
}

void DeviceWebServer::handleOutputTestPage() {
    webServer->send(200, "text/html", generateOutputTestPage());
}

// ============================================================================
// Signal Profile HTML Page Generators
// ============================================================================

String DeviceWebServer::generateProfilePage() {
    String html = getHTMLHeader("Signal Profile");
    html += "<div class='container'>";
    html += getNavigation();

    if (!profileManager || !profileManager->hasProfile()) {
        html += "<div class='warning-box'>";
        html += "No signal profile loaded. Device is using default behavior.";
        html += "</div>";
        html += "</div>";
        html += getHTMLFooter();
        return html;
    }

    // Profile info card
    html += "<div class='card'>";
    html += "<h2>Profile Information</h2>";
    html += "<table>";
    html += "<tr><th>Property</th><th>Value</th></tr>";
    html += "<tr><td>Profile Name</td><td>" + profileManager->getProfileName() + "</td></tr>";
    html += "<tr><td>Profile ID</td><td style='font-family:monospace;font-size:12px;'>" + profileManager->getProfileId() + "</td></tr>";
    html += "<tr><td>Version</td><td>" + String(profileManager->getProfileVersion()) + "</td></tr>";
    html += "<tr><td>Default State</td><td><strong>" + String(profileManager->getDefaultState()) + "</strong></td></tr>";
    html += "<tr><td>Total States</td><td>" + String(profileManager->getStateCount()) + "</td></tr>";
    html += "</table>";
    html += "</div>";

    // States configuration card
    html += "<div class='card'>";
    html += "<h2>States & Output Configuration</h2>";
    html += "<div id='statesContainer'></div>";
    html += "</div>";

    // Button behavior card
    html += "<div class='card'>";
    html += "<h2>Button Behavior</h2>";
    html += "<div id='buttonBehavior'></div>";
    html += "</div>";

    // Add CSS for cycle list
    html += "<style>";
    html += ".cycle-list { background:#f7fafc; padding:15px; border-radius:6px; }";
    html += ".cycle-item { padding:8px; margin-bottom:5px; background:white; ";
    html += "  border-left:3px solid #667eea; border-radius:4px; }";
    html += "</style>";

    // JavaScript to load and render profile
    html += "<script>";
    html += "fetch('/profile').then(r => r.json()).then(renderProfile);";
    html += "function renderProfile(profile) {";
    html += "  renderStates(profile.states);";
    html += "  renderButtonBehavior(profile.buttonBehavior);";
    html += "}";

    // Render states function
    html += "function renderStates(states) {";
    html += "  let html = '<table><thead><tr>';";
    html += "  html += '<th>State</th><th>Red</th><th>Yellow</th><th>Green</th><th>Primary Buzzer</th><th>Tower Buzzer</th>';";
    html += "  html += '</tr></thead><tbody>';";
    html += "  states.forEach(s => {";
    html += "    html += '<tr>';";
    html += "    html += '<td><strong>' + s.name + '</strong></td>';";
    html += "    html += '<td>' + formatOutput(s.outputs.redLight, 'red') + '</td>';";
    html += "    html += '<td>' + formatOutput(s.outputs.yellowLight, 'yellow') + '</td>';";
    html += "    html += '<td>' + formatOutput(s.outputs.greenLight, 'green') + '</td>';";
    html += "    html += '<td>' + formatBuzzer(s.outputs.primaryBuzzer || s.outputs.buzzer || 'off') + '</td>';";
    html += "    html += '<td>' + formatBuzzer(s.outputs.towerBuzzer || s.outputs.buzzer || 'off') + '</td>';";
    html += "    html += '</tr>';";
    html += "  });";
    html += "  html += '</tbody></table>';";
    html += "  document.getElementById('statesContainer').innerHTML = html;";
    html += "}";

    // Format output helper
    html += "function formatOutput(mode, color) {";
    html += "  const colors = {red:'#f56565', yellow:'#ed8936', green:'#48bb78'};";
    html += "  const icons = {off:'○', on:'●', shortBlink:'◐', longBlink:'◑'};";
    html += "  const bg = mode === 'off' ? '#eee' : colors[color];";
    html += "  const text = mode === 'off' ? '#999' : 'white';";
    html += "  let span = '<span style=\"display:inline-block;padding:4px 12px;border-radius:4px;";
    html += "    background:' + bg + ';color:' + text + ';font-size:12px;font-weight:600;\">';";
    html += "  return span + icons[mode] + ' ' + mode + '</span>';";
    html += "}";

    // Format buzzer helper
    html += "function formatBuzzer(mode) {";
    html += "  const icons = {off:'🔇', on:'🔊', chirp:'🔔'};";
    html += "  const colors = {off:'#eee', on:'#f56565', chirp:'#ed8936'};";
    html += "  const text = {off:'#999', on:'white', chirp:'white'};";
    html += "  let span = '<span style=\"display:inline-block;padding:4px 12px;border-radius:4px;";
    html += "    background:' + colors[mode] + ';color:' + text[mode] + ';font-size:12px;font-weight:600;\">';";
    html += "  return span + icons[mode] + ' ' + mode + '</span>';";
    html += "}";

    // Render button behavior
    html += "function renderButtonBehavior(behavior) {";
    html += "  let html = '<div style=\"display:grid;grid-template-columns:1fr 1fr;gap:20px;\">';";
    html += "  html += '<div>';";
    html += "  html += '<h3 style=\"color:#667eea;margin-bottom:10px;\">Short Press Cycle</h3>';";
    html += "  html += '<div class=\"cycle-list\">';";
    html += "  behavior.shortPressCycle.forEach((s, i) => {";
    html += "    html += '<div class=\"cycle-item\">' + (i+1) + '. ' + s + '</div>';";
    html += "  });";
    html += "  html += '</div></div>';";
    html += "  html += '<div>';";
    html += "  html += '<h3 style=\"color:#667eea;margin-bottom:10px;\">Long Press Cycle</h3>';";
    html += "  html += '<div class=\"cycle-list\">';";
    html += "  behavior.longPressCycle.forEach((s, i) => {";
    html += "    html += '<div class=\"cycle-item\">' + (i+1) + '. ' + s + '</div>';";
    html += "  });";
    html += "  html += '</div></div>';";
    html += "  html += '</div>';";
    html += "  document.getElementById('buttonBehavior').innerHTML = html;";
    html += "}";
    html += "</script>";

    html += "</div>";
    html += getHTMLFooter();
    return html;
}

String DeviceWebServer::generateStateControlPage() {
    String html = getHTMLHeader("State Control");
    html += "<div class='container'>";
    html += getNavigation();

    if (!profileManager || !profileManager->hasProfile()) {
        html += "<div class='warning-box'>";
        html += "No signal profile loaded. State control unavailable.";
        html += "</div>";
        html += "</div>";
        html += getHTMLFooter();
        return html;
    }

    // Current state card
    html += "<div class='card'>";
    html += "<h2>Current State</h2>";
    html += "<div id='currentStateDisplay' style='text-align:center;padding:30px;'>";
    html += "<div id='stateValue' style='font-size:36px;font-weight:bold;color:#667eea;'>Loading...</div>";
    html += "<div id='overrideIndicator' style='margin-top:15px;'></div>";
    html += "</div>";
    html += "</div>";

    // State selector card
    html += "<div class='card'>";
    html += "<h2>Manual State Control</h2>";
    html += "<div class='info-box'>";
    html += "Manually setting state creates an override. The device will remain in this state ";
    html += "until changed via button press, backend command, or override is cleared.";
    html += "</div>";
    html += "<div id='message' class='message'></div>";
    html += "<form id='stateForm' onsubmit='setState(event)'>";
    html += "<div class='form-group'>";
    html += "<label>Select State:</label>";
    html += "<select id='stateSelector' name='state' style='width:100%;padding:12px;border:2px solid #ddd;";
    html += "  border-radius:6px;font-size:14px;'>";
    html += "<option value=''>-- Loading states --</option>";
    html += "</select>";
    html += "</div>";
    html += "<button type='submit' class='btn btn-success'>Set State</button>";
    html += "<button type='button' class='btn btn-danger' onclick='clearOverride()' id='clearBtn' disabled>";
    html += "Clear Override & Return to Default</button>";
    html += "</form>";
    html += "</div>";

    // JavaScript
    html += "<script>";
    html += "let currentStateData = null;";

    // Load current state
    html += "function loadState() {";
    html += "  fetch('/state').then(r => r.json()).then(data => {";
    html += "    currentStateData = data;";
    html += "    document.getElementById('stateValue').textContent = data.currentState;";
    html += "    ";
    html += "    if (data.isOverridden) {";
    html += "      document.getElementById('overrideIndicator').innerHTML = ";
    html += "        '<span class=\"status-badge\" style=\"background:#fed7d7;color:#742a2a;\">";
    html += "        ⚠ OVERRIDE ACTIVE</span>';";
    html += "      document.getElementById('clearBtn').disabled = false;";
    html += "    } else {";
    html += "      document.getElementById('overrideIndicator').innerHTML = ";
    html += "        '<span class=\"status-badge\" style=\"background:#c6f6d5;color:#22543d;\">";
    html += "        ✓ Normal Operation</span>';";
    html += "      document.getElementById('clearBtn').disabled = true;";
    html += "    }";
    html += "    ";
    html += "    if (data.availableStates) {";
    html += "      const select = document.getElementById('stateSelector');";
    html += "      select.innerHTML = '';";
    html += "      data.availableStates.forEach(s => {";
    html += "        const opt = document.createElement('option');";
    html += "        opt.value = s;";
    html += "        opt.textContent = s;";
    html += "        if (s === data.currentState) opt.selected = true;";
    html += "        select.appendChild(opt);";
    html += "      });";
    html += "    }";
    html += "  });";
    html += "}";

    // Set state
    html += "function setState(e) {";
    html += "  e.preventDefault();";
    html += "  const state = document.getElementById('stateSelector').value;";
    html += "  const data = new URLSearchParams({state: state});";
    html += "  fetch('/state/set', {method: 'POST', body: data})";
    html += "    .then(r => r.json())";
    html += "    .then(d => {";
    html += "      const msg = document.getElementById('message');";
    html += "      msg.textContent = d.success ? 'State changed to: ' + d.state : 'Error: ' + d.error;";
    html += "      msg.className = 'message ' + (d.success ? 'success' : 'error');";
    html += "      if (d.success) setTimeout(loadState, 500);";
    html += "    });";
    html += "}";

    // Clear override
    html += "function clearOverride() {";
    html += "  if (!confirm('Clear override and return to default state?')) return;";
    html += "  fetch('/override/clear', {method: 'POST'})";
    html += "    .then(r => r.json())";
    html += "    .then(d => {";
    html += "      const msg = document.getElementById('message');";
    html += "      msg.textContent = 'Override cleared. Returned to state: ' + d.state;";
    html += "      msg.className = 'message success';";
    html += "      setTimeout(loadState, 500);";
    html += "    });";
    html += "}";

    html += "loadState();";
    html += "setInterval(loadState, 5000);";  // Refresh every 5 seconds
    html += "</script>";

    html += "</div>";
    html += getHTMLFooter();
    return html;
}

String DeviceWebServer::generateOutputTestPage() {
    String html = getHTMLHeader("Output Testing");
    html += "<div class='container'>";
    html += getNavigation();

    html += "<div class='warning-box'>";
    html += "⚠ <strong>WARNING:</strong> Output testing directly controls hardware outputs, ";
    html += "bypassing the signal profile. Use for hardware verification only.";
    html += "</div>";

    // Tower lights card
    html += "<div class='card'>";
    html += "<h2>Tower Lights</h2>";
    html += "<div style='display:grid;grid-template-columns:repeat(3,1fr);gap:15px;margin-bottom:20px;'>";

    // Red light
    html += "<div style='text-align:center;'>";
    html += "<div style='width:80px;height:80px;border-radius:50%;background:#f56565;";
    html += "  margin:0 auto 10px;box-shadow:0 4px 10px rgba(245,101,101,0.3);opacity:0.3;' id='redPreview'></div>";
    html += "<label style='display:block;margin-bottom:8px;font-weight:600;'>Red Light</label>";
    html += "<button class='btn' id='redBtn' onclick='toggleLight(\"red\")' ";
    html += "  style='width:100%;background:#f56565;'>OFF</button>";
    html += "</div>";

    // Yellow light
    html += "<div style='text-align:center;'>";
    html += "<div style='width:80px;height:80px;border-radius:50%;background:#ed8936;";
    html += "  margin:0 auto 10px;opacity:0.3;box-shadow:0 4px 10px rgba(237,137,54,0.3);' id='yellowPreview'></div>";
    html += "<label style='display:block;margin-bottom:8px;font-weight:600;'>Yellow Light</label>";
    html += "<button class='btn' id='yellowBtn' onclick='toggleLight(\"yellow\")' ";
    html += "  style='width:100%;background:#ed8936;'>OFF</button>";
    html += "</div>";

    // Green light
    html += "<div style='text-align:center;'>";
    html += "<div style='width:80px;height:80px;border-radius:50%;background:#48bb78;";
    html += "  margin:0 auto 10px;opacity:0.3;box-shadow:0 4px 10px rgba(72,187,120,0.3);' id='greenPreview'></div>";
    html += "<label style='display:block;margin-bottom:8px;font-weight:600;'>Green Light</label>";
    html += "<button class='btn' id='greenBtn' onclick='toggleLight(\"green\")' ";
    html += "  style='width:100%;background:#48bb78;'>OFF</button>";
    html += "</div>";

    html += "</div>";
    html += "</div>";

    // Buzzer card
    html += "<div class='card'>";
    html += "<h2>Buzzers</h2>";
    html += "<div style='display:grid;grid-template-columns:1fr 1fr;gap:20px;padding:20px;'>";

    // Primary Buzzer
    html += "<div style='text-align:center;'>";
    html += "<div style='font-size:48px;margin-bottom:10px;' id='primaryBuzzerPreview'>🔇</div>";
    html += "<label style='display:block;margin-bottom:8px;font-weight:600;'>Primary Buzzer (GPIO46)</label>";
    html += "<button class='btn' id='primaryBuzzerBtn' onclick='togglePrimaryBuzzer()' ";
    html += "  style='width:100%;background:#718096;'>OFF</button>";
    html += "</div>";

    // Tower Buzzer
    html += "<div style='text-align:center;'>";
    html += "<div style='font-size:48px;margin-bottom:10px;' id='towerBuzzerPreview'>🔇</div>";
    html += "<label style='display:block;margin-bottom:8px;font-weight:600;'>Tower Buzzer (DO4)</label>";
    html += "<button class='btn' id='towerBuzzerBtn' onclick='toggleTowerBuzzer()' ";
    html += "  style='width:100%;background:#718096;'>OFF</button>";
    html += "</div>";

    html += "</div>";
    html += "</div>";

    // All off button
    html += "<div class='card'>";
    html += "<h2>Quick Actions</h2>";
    html += "<button class='btn btn-danger' onclick='allOff()' style='width:100%;'>Turn All Outputs OFF</button>";
    html += "</div>";

    // JavaScript
    html += "<script>";
    html += "const state = {red:false, yellow:false, green:false, primaryBuzzer:false, towerBuzzer:false};";

    html += "function toggleLight(color) {";
    html += "  state[color] = !state[color];";
    html += "  updateOutputs();";
    html += "  updateUI();";
    html += "}";

    html += "function togglePrimaryBuzzer() {";
    html += "  state.primaryBuzzer = !state.primaryBuzzer;";
    html += "  updateOutputs();";
    html += "  updateUI();";
    html += "}";

    html += "function toggleTowerBuzzer() {";
    html += "  state.towerBuzzer = !state.towerBuzzer;";
    html += "  updateOutputs();";
    html += "  updateUI();";
    html += "}";

    html += "function allOff() {";
    html += "  state.red = state.yellow = state.green = state.primaryBuzzer = state.towerBuzzer = false;";
    html += "  updateOutputs();";
    html += "  updateUI();";
    html += "}";

    html += "function updateOutputs() {";
    html += "  fetch('/outputs/test', {";
    html += "    method: 'POST',";
    html += "    headers: {'Content-Type': 'application/json'},";
    html += "    body: JSON.stringify(state)";
    html += "  });";
    html += "}";

    html += "function updateUI() {";
    html += "  document.getElementById('redBtn').textContent = state.red ? 'ON' : 'OFF';";
    html += "  document.getElementById('yellowBtn').textContent = state.yellow ? 'ON' : 'OFF';";
    html += "  document.getElementById('greenBtn').textContent = state.green ? 'ON' : 'OFF';";
    html += "  document.getElementById('primaryBuzzerBtn').textContent = state.primaryBuzzer ? 'ON' : 'OFF';";
    html += "  document.getElementById('towerBuzzerBtn').textContent = state.towerBuzzer ? 'ON' : 'OFF';";
    html += "  ";
    html += "  document.getElementById('redPreview').style.opacity = state.red ? '1' : '0.3';";
    html += "  document.getElementById('yellowPreview').style.opacity = state.yellow ? '1' : '0.3';";
    html += "  document.getElementById('greenPreview').style.opacity = state.green ? '1' : '0.3';";
    html += "  document.getElementById('primaryBuzzerPreview').textContent = state.primaryBuzzer ? '🔊' : '🔇';";
    html += "  document.getElementById('towerBuzzerPreview').textContent = state.towerBuzzer ? '🔊' : '🔇';";
    html += "}";
    html += "</script>";

    html += "</div>";
    html += getHTMLFooter();
    return html;
}
