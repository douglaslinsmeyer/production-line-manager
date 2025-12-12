#include "device_webserver.h"
#include "config.h"

extern DeviceConfig deviceConfig;
extern char deviceMAC[18];

DeviceWebServer::DeviceWebServer()
    : webServer(nullptr),
      running(false),
      serverPort(80) {
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
    json += "\"connection_mode\":\"" + String(settings.connectionMode == MODE_WIFI ? "wifi" : "ethernet") + "\",";
    json += "\"wifi_enabled\":" + String(settings.wifiEnabled ? "true" : "false");
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
