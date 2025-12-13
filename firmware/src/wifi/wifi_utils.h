#pragma once
#include <Arduino.h>

/**
 * WiFi utility functions for signal strength conversion and display
 */
namespace WiFiUtils {
    /**
     * Convert RSSI (dBm) to signal strength bars (0-4)
     *
     * Signal strength ranges:
     * - 4 bars (Excellent): >= -50 dBm
     * - 3 bars (Good):      -51 to -60 dBm
     * - 2 bars (Fair):      -61 to -70 dBm
     * - 1 bar  (Poor):      <= -71 dBm
     * - 0 bars:             No signal / disconnected (RSSI = 0)
     *
     * @param rssi Signal strength in dBm (negative value)
     * @return Bar count (0 = disconnected, 1-4 = signal strength)
     */
    uint8_t rssiToBars(int rssi);

    /**
     * Convert RSSI to Unicode bar string for display
     *
     * Uses Unicode block characters:
     * - Full bar:  █ (U+2588)
     * - Empty bar: ▂ (U+2582)
     *
     * Examples:
     * - "████" for excellent signal
     * - "██▂▂" for fair signal
     * - "▂▂▂▂" for no signal
     *
     * @param rssi Signal strength in dBm
     * @return Unicode string like "████" or "██▂▂"
     */
    String rssiToBarString(int rssi);

    /**
     * Get signal quality description from RSSI
     *
     * @param rssi Signal strength in dBm
     * @return "Excellent", "Good", "Fair", "Poor", or "No Signal"
     */
    String rssiToQuality(int rssi);
}
