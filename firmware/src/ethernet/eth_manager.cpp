#include "eth_manager.h"
#include "config.h"

// Static instance pointer for event callback
EthernetManager* EthernetManager::instance = nullptr;

EthernetManager::EthernetManager()
    : connected(false), connCallback(nullptr), lastStatusCheck(0) {
    instance = this;
}

bool EthernetManager::begin() {
    Serial.println("Initializing W5500 Ethernet...");

    // CRITICAL: Wait for ESP32-S3 boot glitches to settle
    // ESP32-S3 Datasheet: GPIO1-20 have 60µs low-level glitches during power-up
    // GPIO12-16 are used for W5500 SPI - must wait for glitches to complete
    delay(BOOT_STABILIZATION_DELAY);

    // Initialize SPI bus for W5500 FIRST (before event registration)
    SPI.begin(ETH_SPI_SCK, ETH_SPI_MISO, ETH_SPI_MOSI);

    // Register network event handler AFTER SPI is ready
    Network.onEvent(onEvent);

    // Initialize W5500 Ethernet (using ESP32 ETH library with W5500 support)
    // Let ETH.begin() handle the reset internally - do NOT manually reset beforehand
    Serial.printf("  W5500 CS: GPIO%d\n", ETH_PHY_CS);
    Serial.printf("  W5500 RST: GPIO%d\n", ETH_PHY_RST);
    Serial.printf("  SPI SCK: GPIO%d, MISO: GPIO%d, MOSI: GPIO%d\n",
                 ETH_SPI_SCK, ETH_SPI_MISO, ETH_SPI_MOSI);

    // IMPORTANT: ETH.begin() will handle the RST pin reset internally
    // Do not manually reset the W5500 beforehand as it causes DHCP issues
    bool success = ETH.begin(ETH_PHY_TYPE, ETH_PHY_ADDR, ETH_PHY_CS,
                             ETH_PHY_IRQ, ETH_PHY_RST, SPI);

    if (!success) {
        Serial.println("ERROR: Failed to initialize W5500");
        return false;
    }

    // Set hostname for DHCP
    ETH.setHostname("esp32-s3-poe-eth");

    // Configure static IP if not using DHCP
    if (!USE_DHCP) {
        IPAddress local_ip, gateway, subnet, dns;
        local_ip.fromString(STATIC_IP);
        gateway.fromString(GATEWAY);
        subnet.fromString(SUBNET);
        dns.fromString(DNS_SERVER);
        ETH.config(local_ip, gateway, subnet, dns);
        Serial.println("Static IP configured");
    } else {
        Serial.println("DHCP enabled - waiting for IP address...");
    }

    Serial.println("W5500 initialized - waiting for connection...");
    return true;
}

void EthernetManager::update() {
    // No polling needed - events are handled via callbacks
}

bool EthernetManager::isConnected() {
    return connected && ETH.linkUp();
}

IPAddress EthernetManager::getIP() {
    return ETH.localIP();
}

void EthernetManager::setConnectionCallback(ConnectionCallback callback) {
    connCallback = callback;
}

void EthernetManager::onEvent(arduino_event_id_t event, arduino_event_info_t info) {
    switch (event) {
        case ARDUINO_EVENT_ETH_START:
            Serial.println("ETH Started");
            break;

        case ARDUINO_EVENT_ETH_CONNECTED:
            Serial.println("ETH Cable Connected");
            break;

        case ARDUINO_EVENT_ETH_GOT_IP:
            Serial.print("ETH Got IP: ");
            Serial.println(ETH.localIP());
            Serial.print("  Gateway: ");
            Serial.println(ETH.gatewayIP());
            Serial.print("  Subnet: ");
            Serial.println(ETH.subnetMask());
            Serial.print("  MAC: ");
            Serial.println(ETH.macAddress());

            if (instance != nullptr) {
                instance->connected = true;
                if (instance->connCallback != nullptr) {
                    instance->connCallback(true);
                }
            }
            break;

        case ARDUINO_EVENT_ETH_DISCONNECTED:
        case ARDUINO_EVENT_ETH_LOST_IP:
            Serial.println("ETH Disconnected");

            if (instance != nullptr) {
                instance->connected = false;
                if (instance->connCallback != nullptr) {
                    instance->connCallback(false);
                }
            }
            break;

        case ARDUINO_EVENT_ETH_STOP:
            Serial.println("ETH Stopped");

            if (instance != nullptr) {
                instance->connected = false;
                if (instance->connCallback != nullptr) {
                    instance->connCallback(false);
                }
            }
            break;

        default:
            break;
    }
}
