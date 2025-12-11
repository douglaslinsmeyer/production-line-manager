#include "captive_portal.h"

const IPAddress CaptivePortal::AP_IP(192, 168, 4, 1);

CaptivePortal::CaptivePortal()
    : webServer(nullptr),
      dnsServer(nullptr),
      credentialsSet(false),
      savedCallback(nullptr) {
}

CaptivePortal::~CaptivePortal() {
    stop();
}

bool CaptivePortal::begin(const char* mac) {
    deviceMAC = String(mac);

    Serial.println("Starting captive portal...");

    // Create DNS server (redirects all requests to AP IP)
    dnsServer = new DNSServer();
    dnsServer->start(DNS_PORT, "*", AP_IP);
    Serial.printf("  DNS server started on port %d\n", DNS_PORT);

    // Create web server
    webServer = new WebServer(WEB_PORT);

    // Register HTTP handlers
    webServer->on("/", [this]() { handleRoot(); });
    webServer->on("/scan", [this]() { handleScan(); });
    webServer->on("/save", [this]() { handleSave(); });
    webServer->onNotFound([this]() { handleNotFound(); });

    webServer->begin();
    Serial.printf("  Web server started on port %d\n", WEB_PORT);
    Serial.printf("  Access portal at: http://%s\n", AP_IP.toString().c_str());

    return true;
}

void CaptivePortal::stop() {
    if (webServer) {
        webServer->stop();
        delete webServer;
        webServer = nullptr;
    }

    if (dnsServer) {
        dnsServer->stop();
        delete dnsServer;
        dnsServer = nullptr;
    }

    Serial.println("Captive portal stopped");
}

void CaptivePortal::update() {
    if (dnsServer) {
        dnsServer->processNextRequest();
    }

    if (webServer) {
        webServer->handleClient();
    }
}

void CaptivePortal::getCredentials(String& ssid, String& password) {
    ssid = submittedSSID;
    password = submittedPassword;
}

void CaptivePortal::setCredentialsSavedCallback(CredentialsSavedCallback callback) {
    savedCallback = callback;
}

void CaptivePortal::handleRoot() {
    Serial.println("HTTP: GET /");
    webServer->send(200, "text/html", generateSetupPage());
}

void CaptivePortal::handleScan() {
    Serial.println("HTTP: GET /scan (WiFi scan)");

    // Scan for networks
    int numNetworks = WiFi.scanNetworks();

    // Build JSON response
    String json = "[";
    for (int i = 0; i < numNetworks; i++) {
        if (i > 0) json += ",";

        json += "{";
        json += "\"ssid\":\"" + WiFi.SSID(i) + "\",";
        json += "\"rssi\":" + String(WiFi.RSSI(i)) + ",";
        json += "\"encryption\":\"" + getEncryptionType(WiFi.encryptionType(i)) + "\"";
        json += "}";
    }
    json += "]";

    webServer->send(200, "application/json", json);
}

void CaptivePortal::handleSave() {
    Serial.println("HTTP: POST /save");

    // Get SSID and password from POST parameters
    if (webServer->hasArg("ssid") && webServer->hasArg("password")) {
        submittedSSID = webServer->arg("ssid");
        submittedPassword = webServer->arg("password");

        // Validate SSID
        if (submittedSSID.length() == 0 || submittedSSID.length() > 32) {
            webServer->send(400, "application/json",
                           "{\"success\":false,\"message\":\"SSID must be 1-32 characters\"}");
            return;
        }

        // Validate password (0 for open, or 8-63 for WPA2)
        if (submittedPassword.length() != 0 &&
            (submittedPassword.length() < 8 || submittedPassword.length() > 63)) {
            webServer->send(400, "application/json",
                           "{\"success\":false,\"message\":\"Password must be 0 (open) or 8-63 characters\"}");
            return;
        }

        credentialsSet = true;

        Serial.printf("Credentials received: SSID='%s', Password='%s'\n",
                     submittedSSID.c_str(),
                     submittedPassword.length() > 0 ? "****" : "(none)");

        // Trigger callback
        if (savedCallback) {
            savedCallback(submittedSSID.c_str(), submittedPassword.c_str());
        }

        // Send success response
        webServer->send(200, "application/json",
                       "{\"success\":true,\"message\":\"Configuration saved. Device will reboot in 3 seconds...\"}");
    } else {
        webServer->send(400, "application/json",
                       "{\"success\":false,\"message\":\"Missing SSID or password\"}");
    }
}

void CaptivePortal::handleNotFound() {
    // Redirect all other requests to root (captive portal behavior)
    Serial.printf("HTTP: GET %s (redirect to /)\n", webServer->uri().c_str());
    webServer->sendHeader("Location", "/", true);
    webServer->send(302, "text/plain", "");
}

String CaptivePortal::generateSetupPage() {
    String html = R"rawliteral(
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>ESP32 WiFi Setup</title>
    <style>
        * { margin: 0; padding: 0; box-sizing: border-box; }
        body {
            font-family: Arial, sans-serif;
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            min-height: 100vh;
            display: flex;
            align-items: center;
            justify-content: center;
            padding: 20px;
        }
        .container {
            background: white;
            border-radius: 12px;
            box-shadow: 0 20px 60px rgba(0,0,0,0.3);
            max-width: 500px;
            width: 100%;
            padding: 30px;
        }
        h1 {
            color: #333;
            margin-bottom: 10px;
            font-size: 24px;
        }
        .device-info {
            background: #f5f5f5;
            padding: 12px;
            border-radius: 6px;
            margin-bottom: 20px;
            font-size: 14px;
            color: #666;
        }
        .form-group {
            margin-bottom: 20px;
        }
        label {
            display: block;
            margin-bottom: 8px;
            color: #555;
            font-weight: 500;
        }
        input, select {
            width: 100%;
            padding: 12px;
            border: 2px solid #ddd;
            border-radius: 6px;
            font-size: 16px;
            transition: border-color 0.3s;
        }
        input:focus, select:focus {
            outline: none;
            border-color: #667eea;
        }
        .btn {
            width: 100%;
            padding: 14px;
            background: #667eea;
            color: white;
            border: none;
            border-radius: 6px;
            font-size: 16px;
            font-weight: 600;
            cursor: pointer;
            transition: background 0.3s;
        }
        .btn:hover { background: #5568d3; }
        .btn:disabled {
            background: #ccc;
            cursor: not-allowed;
        }
        .btn-scan {
            background: #48bb78;
            margin-bottom: 15px;
        }
        .btn-scan:hover { background: #38a169; }
        .networks {
            max-height: 200px;
            overflow-y: auto;
            border: 2px solid #ddd;
            border-radius: 6px;
            margin-bottom: 15px;
        }
        .network-item {
            padding: 12px;
            border-bottom: 1px solid #eee;
            cursor: pointer;
            transition: background 0.2s;
        }
        .network-item:hover { background: #f5f5f5; }
        .network-item:last-child { border-bottom: none; }
        .network-ssid {
            font-weight: 500;
            color: #333;
        }
        .network-rssi {
            font-size: 12px;
            color: #888;
            margin-left: 10px;
        }
        .network-lock {
            float: right;
            color: #666;
        }
        .message {
            padding: 12px;
            border-radius: 6px;
            margin-bottom: 15px;
            display: none;
        }
        .message.success {
            background: #c6f6d5;
            color: #22543d;
            display: block;
        }
        .message.error {
            background: #fed7d7;
            color: #742a2a;
            display: block;
        }
        .loading {
            text-align: center;
            color: #666;
            padding: 20px;
        }
    </style>
</head>
<body>
    <div class="container">
        <h1>🛜 WiFi Configuration</h1>
        <div class="device-info">
            <strong>Device:</strong> ESP32-S3-POE-8DI8DO<br>
            <strong>MAC:</strong> )rawliteral";

    html += deviceMAC;

    html += R"rawliteral(
        </div>

        <div id="message" class="message"></div>

        <button class="btn btn-scan" onclick="scanNetworks()">Scan WiFi Networks</button>

        <div id="networks" class="networks" style="display:none;"></div>

        <form onsubmit="saveConfig(event)">
            <div class="form-group">
                <label>Network SSID:</label>
                <input type="text" id="ssid" name="ssid" required maxlength="32"
                       placeholder="Enter WiFi network name">
            </div>

            <div class="form-group">
                <label>Password:</label>
                <input type="password" id="password" name="password"
                       placeholder="Leave empty for open networks" maxlength="63">
            </div>

            <button type="submit" class="btn" id="saveBtn">Connect to WiFi</button>
        </form>
    </div>

    <script>
        function showMessage(text, type) {
            const msg = document.getElementById('message');
            msg.textContent = text;
            msg.className = 'message ' + type;
        }

        function scanNetworks() {
            const networksDiv = document.getElementById('networks');
            networksDiv.innerHTML = '<div class="loading">Scanning...</div>';
            networksDiv.style.display = 'block';

            fetch('/scan')
                .then(response => response.json())
                .then(networks => {
                    if (networks.length === 0) {
                        networksDiv.innerHTML = '<div class="loading">No networks found</div>';
                        return;
                    }

                    let html = '';
                    networks.forEach(net => {
                        const lock = net.encryption !== 'Open' ? '🔒' : '';
                        const rssiText = net.rssi + ' dBm';
                        html += `<div class="network-item" onclick="selectNetwork('${net.ssid}')">
                            <span class="network-lock">${lock}</span>
                            <span class="network-ssid">${net.ssid}</span>
                            <span class="network-rssi">${rssiText}</span>
                        </div>`;
                    });
                    networksDiv.innerHTML = html;
                })
                .catch(err => {
                    networksDiv.innerHTML = '<div class="loading">Scan failed</div>';
                    console.error('Scan error:', err);
                });
        }

        function selectNetwork(ssid) {
            document.getElementById('ssid').value = ssid;
            document.getElementById('password').focus();
        }

        function saveConfig(event) {
            event.preventDefault();

            const ssid = document.getElementById('ssid').value;
            const password = document.getElementById('password').value;
            const saveBtn = document.getElementById('saveBtn');

            saveBtn.disabled = true;
            saveBtn.textContent = 'Saving...';

            fetch('/save', {
                method: 'POST',
                headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
                body: 'ssid=' + encodeURIComponent(ssid) + '&password=' + encodeURIComponent(password)
            })
            .then(response => response.json())
            .then(data => {
                if (data.success) {
                    showMessage(data.message, 'success');
                    setTimeout(() => {
                        document.body.innerHTML = '<div class="container"><h1>✓ Configuration Saved</h1><p>Device is rebooting and connecting to WiFi...</p><p style="margin-top: 20px;">After reboot, access the device configuration at:<br><strong>http://&lt;device-ip&gt;</strong></p></div>';
                    }, 1000);
                } else {
                    showMessage(data.message, 'error');
                    saveBtn.disabled = false;
                    saveBtn.textContent = 'Connect to WiFi';
                }
            })
            .catch(err => {
                showMessage('Connection error. Please try again.', 'error');
                saveBtn.disabled = false;
                saveBtn.textContent = 'Connect to WiFi';
                console.error('Save error:', err);
            });
        }

        // Auto-scan on page load
        window.addEventListener('load', () => {
            setTimeout(scanNetworks, 500);
        });
    </script>
</body>
</html>
)rawliteral";

    return html;
}

String CaptivePortal::getEncryptionType(wifi_auth_mode_t encryption) {
    switch (encryption) {
        case WIFI_AUTH_OPEN:
            return "Open";
        case WIFI_AUTH_WEP:
            return "WEP";
        case WIFI_AUTH_WPA_PSK:
            return "WPA";
        case WIFI_AUTH_WPA2_PSK:
            return "WPA2";
        case WIFI_AUTH_WPA_WPA2_PSK:
            return "WPA/WPA2";
        case WIFI_AUTH_WPA2_ENTERPRISE:
            return "WPA2-Enterprise";
        case WIFI_AUTH_WPA3_PSK:
            return "WPA3";
        case WIFI_AUTH_WPA2_WPA3_PSK:
            return "WPA2/WPA3";
        default:
            return "Unknown";
    }
}
