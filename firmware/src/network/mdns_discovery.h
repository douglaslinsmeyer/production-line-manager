#pragma once

#include <Arduino.h>
#include <ESPmDNS.h>
#include <IPAddress.h>
#include <Preferences.h>

/**
 * MDNSDiscovery - MQTT Broker Auto-Discovery via mDNS
 *
 * Provides automatic discovery of MQTT brokers on the local network using
 * multicast DNS (mDNS) service discovery. Supports caching for fast reconnection.
 *
 * Usage:
 *   MDNSDiscovery discovery;
 *   discovery.begin("device-hostname");
 *
 *   MDNSDiscovery::DiscoveryConfig config;
 *   config.enabled = true;
 *   config.timeoutMs = 5000;
 *
 *   auto broker = discovery.discoverBroker(config);
 *   if (broker.valid) {
 *     // Use broker.ip and broker.port
 *   }
 */
class MDNSDiscovery {
public:
    /**
     * Configuration for mDNS discovery
     */
    struct DiscoveryConfig {
        bool enabled;                  // Enable/disable mDNS discovery
        char serviceName[32];          // Service name to discover (e.g., "_mqtt")
        char protocol[8];              // Protocol (e.g., "_tcp")
        uint32_t timeoutMs;            // Discovery timeout in milliseconds
        bool cacheResults;             // Cache discovered IPs in NVS
        uint32_t cacheExpiryMs;        // Cache validity period in milliseconds

        DiscoveryConfig()
            : enabled(true),
              timeoutMs(5000),
              cacheResults(true),
              cacheExpiryMs(3600000) {  // 1 hour default
            strncpy(serviceName, "_mqtt", sizeof(serviceName) - 1);
            serviceName[sizeof(serviceName) - 1] = '\0';
            strncpy(protocol, "_tcp", sizeof(protocol) - 1);
            protocol[sizeof(protocol) - 1] = '\0';
        }
    };

    /**
     * Discovered broker information
     */
    struct DiscoveredBroker {
        IPAddress ip;                  // Broker IP address
        uint16_t port;                 // Broker port
        char hostname[64];             // Broker hostname (if available)
        bool valid;                    // Whether discovery was successful

        DiscoveredBroker() : port(0), valid(false) {
            hostname[0] = '\0';
        }
    };

    /**
     * Constructor
     */
    MDNSDiscovery();

    /**
     * Destructor - cleanup resources
     */
    ~MDNSDiscovery();

    /**
     * Initialize mDNS responder
     *
     * @param hostname Optional hostname for this device (defaults to "esp32-device")
     * @return true if initialization successful
     */
    bool begin(const char* hostname = nullptr);

    /**
     * Discover MQTT broker on the network
     *
     * Performs mDNS service discovery to locate MQTT brokers. Returns the
     * first valid broker found, or an invalid result if none found within timeout.
     *
     * @param config Discovery configuration (service name, timeout, etc.)
     * @return DiscoveredBroker struct with results (check .valid field)
     */
    DiscoveredBroker discoverBroker(const DiscoveryConfig& config);

    /**
     * Get cached broker from NVS storage
     *
     * Retrieves previously discovered broker from cache if available and not expired.
     *
     * @param ip Output parameter for cached IP address
     * @param port Output parameter for cached port
     * @return true if valid cached broker found (not expired)
     */
    bool getCachedBroker(IPAddress& ip, uint16_t& port);

    /**
     * Cache discovered broker to NVS storage
     *
     * Stores broker IP and port to NVS with current timestamp for later retrieval.
     *
     * @param ip Broker IP address to cache
     * @param port Broker port to cache
     */
    void cacheBroker(const IPAddress& ip, uint16_t port);

    /**
     * Clear cached broker from NVS storage
     *
     * Removes cached broker information, forcing fresh discovery on next attempt.
     */
    void clearCache();

    /**
     * Check if mDNS is initialized
     *
     * @return true if mDNS.begin() was successful
     */
    bool isInitialized() const { return initialized; }

private:
    bool initialized;                  // mDNS initialization state
    Preferences cachePrefs;            // NVS handle for cache storage

    /**
     * Validate IP address (not 0.0.0.0)
     *
     * @param ip IP address to validate
     * @return true if IP is valid
     */
    bool isValidIP(const IPAddress& ip) const;

    /**
     * Check if cached data is still valid (not expired)
     *
     * @param cacheTimestamp Timestamp when data was cached (millis)
     * @param expiryMs Cache expiry period in milliseconds
     * @return true if cache is still valid
     */
    bool isCacheValid(unsigned long cacheTimestamp, uint32_t expiryMs) const;
};
