#include "wifi_utils.h"

namespace WiFiUtils {
    uint8_t rssiToBars(int rssi) {
        // No signal or disconnected
        if (rssi == 0) {
            return 0;
        }

        // Apply signal strength thresholds
        if (rssi >= -50) {
            return 4;  // Excellent
        }
        if (rssi >= -60) {
            return 3;  // Good
        }
        if (rssi >= -70) {
            return 2;  // Fair
        }

        return 1;  // Poor
    }

    String rssiToBarString(int rssi) {
        uint8_t bars = rssiToBars(rssi);

        // No signal - show all empty bars
        if (bars == 0) {
            return "▂▂▂▂";
        }

        // Build bar string with appropriate number of filled bars
        String result = "";
        for (uint8_t i = 0; i < 4; i++) {
            result += (i < bars) ? "█" : "▂";
        }
        return result;
    }

    String rssiToQuality(int rssi) {
        // Map bars to quality description
        uint8_t bars = rssiToBars(rssi);

        switch (bars) {
            case 4:
                return "Excellent";
            case 3:
                return "Good";
            case 2:
                return "Fair";
            case 1:
                return "Poor";
            default:
                return "No Signal";
        }
    }
}
