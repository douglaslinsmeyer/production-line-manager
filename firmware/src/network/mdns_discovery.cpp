#include "mdns_discovery.h"

// NVS namespace for mDNS cache
#define MDNS_CACHE_NAMESPACE "mdns_cache"
#define MDNS_CACHE_KEY_IP "broker_ip"
#define MDNS_CACHE_KEY_PORT "broker_port"
#define MDNS_CACHE_KEY_TIME "cache_time"

MDNSDiscovery::MDNSDiscovery() : initialized(false) {
}

MDNSDiscovery::~MDNSDiscovery() {
    cachePrefs.end();
}

bool MDNSDiscovery::begin(const char* hostname) {
    // Use provided hostname or default
    const char* mdnsHostname = (hostname && strlen(hostname) > 0) ? hostname : "esp32-device";

    Serial.printf("Initializing mDNS with hostname: %s\n", mdnsHostname);

    // Initialize mDNS responder
    if (!MDNS.begin(mdnsHostname)) {
        Serial.println("ERROR: mDNS initialization failed");
        initialized = false;
        return false;
    }

    Serial.printf("✓ mDNS initialized: %s.local\n", mdnsHostname);
    initialized = true;

    // Open NVS for cache storage
    if (!cachePrefs.begin(MDNS_CACHE_NAMESPACE, false)) {
        Serial.println("WARNING: Failed to open mDNS cache storage");
    }

    return true;
}

MDNSDiscovery::DiscoveredBroker MDNSDiscovery::discoverBroker(const DiscoveryConfig& config) {
    DiscoveredBroker result;

    if (!config.enabled) {
        Serial.println("mDNS discovery disabled by config");
        return result;
    }

    if (!initialized) {
        Serial.println("ERROR: mDNS not initialized, call begin() first");
        return result;
    }

    Serial.printf("Discovering MQTT brokers via mDNS (%s.%s)...\n",
                 config.serviceName, config.protocol);
    Serial.printf("  Timeout: %u ms\n", config.timeoutMs);

    // Extract service name without leading underscore for queryService
    const char* serviceName = config.serviceName;
    if (serviceName[0] == '_') {
        serviceName++;  // Skip leading underscore
    }

    // Extract protocol without leading underscore
    const char* protocol = config.protocol;
    if (protocol[0] == '_') {
        protocol++;  // Skip leading underscore
    }

    // Perform mDNS query
    unsigned long startTime = millis();
    int numServices = MDNS.queryService(serviceName, protocol);

    unsigned long elapsed = millis() - startTime;
    Serial.printf("mDNS query completed in %lu ms\n", elapsed);

    if (numServices == 0) {
        Serial.println("✗ No MQTT brokers found via mDNS");
        return result;
    }

    Serial.printf("✓ Found %d MQTT broker(s)\n", numServices);

    // Use the first broker found
    for (int i = 0; i < numServices; i++) {
        IPAddress ip = MDNS.address(i);
        uint16_t port = MDNS.port(i);
        String hostname = MDNS.hostname(i);

        Serial.printf("  [%d] %s.local (%s:%d)\n",
                     i, hostname.c_str(), ip.toString().c_str(), port);

        // Validate IP address
        if (!isValidIP(ip)) {
            Serial.printf("  → Skipping invalid IP: %s\n", ip.toString().c_str());
            continue;
        }

        // Take the first valid broker
        result.ip = ip;
        result.port = port;
        strncpy(result.hostname, hostname.c_str(), sizeof(result.hostname) - 1);
        result.hostname[sizeof(result.hostname) - 1] = '\0';
        result.valid = true;

        Serial.printf("✓ Selected broker: %s (%s:%d)\n",
                     result.hostname, result.ip.toString().c_str(), result.port);
        break;
    }

    if (!result.valid) {
        Serial.println("✗ No valid broker found (all had invalid IPs)");
    }

    return result;
}

bool MDNSDiscovery::getCachedBroker(IPAddress& ip, uint16_t& port) {
    // Check if cache storage is available
    if (!cachePrefs.isKey(MDNS_CACHE_KEY_IP) ||
        !cachePrefs.isKey(MDNS_CACHE_KEY_PORT) ||
        !cachePrefs.isKey(MDNS_CACHE_KEY_TIME)) {
        Serial.println("No cached broker found");
        return false;
    }

    // Read cached data
    String cachedIP = cachePrefs.getString(MDNS_CACHE_KEY_IP, "");
    port = cachePrefs.getUShort(MDNS_CACHE_KEY_PORT, 0);
    unsigned long cacheTime = cachePrefs.getULong(MDNS_CACHE_KEY_TIME, 0);

    if (cachedIP.length() == 0 || port == 0) {
        Serial.println("Invalid cached broker data");
        return false;
    }

    // Parse IP address
    if (!ip.fromString(cachedIP)) {
        Serial.printf("Failed to parse cached IP: %s\n", cachedIP.c_str());
        return false;
    }

    // Validate IP
    if (!isValidIP(ip)) {
        Serial.printf("Cached IP is invalid: %s\n", cachedIP.c_str());
        return false;
    }

    // Note: We can't reliably check cache expiry using millis() because:
    // 1. millis() resets on reboot
    // 2. We'd need to store wall-clock time (RTC) for cross-reboot validation
    // For now, we'll trust the cache and let the connection attempt validate it
    // Future enhancement: Use NTP time for proper expiry checking

    Serial.printf("✓ Cached broker found: %s:%d (cached at millis=%lu)\n",
                 cachedIP.c_str(), port, cacheTime);

    return true;
}

void MDNSDiscovery::cacheBroker(const IPAddress& ip, uint16_t port) {
    if (!isValidIP(ip) || port == 0) {
        Serial.println("ERROR: Cannot cache invalid broker");
        return;
    }

    // Store to NVS
    cachePrefs.putString(MDNS_CACHE_KEY_IP, ip.toString());
    cachePrefs.putUShort(MDNS_CACHE_KEY_PORT, port);
    cachePrefs.putULong(MDNS_CACHE_KEY_TIME, millis());

    Serial.printf("✓ Broker cached: %s:%d\n", ip.toString().c_str(), port);
}

void MDNSDiscovery::clearCache() {
    cachePrefs.remove(MDNS_CACHE_KEY_IP);
    cachePrefs.remove(MDNS_CACHE_KEY_PORT);
    cachePrefs.remove(MDNS_CACHE_KEY_TIME);

    Serial.println("✓ Broker cache cleared");
}

bool MDNSDiscovery::isValidIP(const IPAddress& ip) const {
    // Check for 0.0.0.0 (invalid)
    return ip != IPAddress(0, 0, 0, 0);
}

bool MDNSDiscovery::isCacheValid(unsigned long cacheTimestamp, uint32_t expiryMs) const {
    // Note: This doesn't work across reboots since millis() resets
    // For now, we accept the cached value and let connection validation handle it
    // Future: Use RTC/NTP time for proper cross-reboot expiry checking
    unsigned long currentTime = millis();

    // Handle millis() overflow (every ~49 days)
    if (currentTime < cacheTimestamp) {
        // millis() has overflowed, invalidate cache
        return false;
    }

    unsigned long age = currentTime - cacheTimestamp;
    return age < expiryMs;
}
