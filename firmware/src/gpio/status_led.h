#pragma once

#include <Arduino.h>
#include "digital_output.h"

/**
 * Connection Status Enum
 *
 * Represents the four network/MQTT connection states
 */
enum ConnectionStatus {
    STATUS_CONNECTED,      // Network + MQTT connected (solid ON)
    STATUS_NO_MQTT,        // Network connected, no MQTT (double blink)
    STATUS_NO_NETWORK,     // No network connectivity (single blink)
    STATUS_AP_MODE         // Access Point mode (slow blink)
};

/**
 * Status LED Controller
 *
 * Controls DO4 (TCA9554PWR channel 3) to provide visual feedback
 * of network and MQTT connection status.
 *
 * LED Patterns:
 * - CONNECTED: Solid ON (network + MQTT connected)
 * - NO_MQTT: Double blink (150ms on, 150ms off, 150ms on, 650ms pause)
 * - NO_NETWORK: Single blink (500ms on, 1000ms off)
 * - AP_MODE: Slow blink (1000ms on, 1000ms off)
 *
 * Uses non-blocking pattern system similar to ButtonLED.
 */
class StatusLEDController {
public:
    StatusLEDController(DigitalOutputManager* outputMgr);

    /**
     * Initialize status LED
     * Ensures DO4 is configured and LED starts in off state
     */
    void begin();

    /**
     * Update LED state based on pattern (call in main loop)
     * Non-blocking pattern generation
     */
    void update();

    /**
     * Set connection status and update LED pattern
     * @param status Current network/MQTT connection status
     */
    void setConnectionStatus(ConnectionStatus status);

private:
    /**
     * Blink Phase Enum (for double blink pattern)
     *
     * Tracks which phase of the double blink we're in
     */
    enum BlinkPhase {
        PHASE_FIRST_ON,    // First blink ON (150ms)
        PHASE_FIRST_OFF,   // First blink OFF (150ms)
        PHASE_SECOND_ON,   // Second blink ON (150ms)
        PHASE_PAUSE        // Long pause (650ms)
    };

    DigitalOutputManager* outputs;

    ConnectionStatus currentStatus;
    bool ledState;                   // Current physical LED state (on/off)
    unsigned long lastToggle;        // Last pattern toggle time (for simple patterns)

    // Double blink pattern state
    BlinkPhase currentPhase;
    unsigned long phaseStartTime;

    // Pattern timing constants (channel defined in config.h)
    static const uint16_t PATTERN_DOUBLE_BLINK_ON = 150;     // 150ms for each blink ON
    static const uint16_t PATTERN_DOUBLE_BLINK_OFF = 150;    // 150ms for each blink OFF
    static const uint16_t PATTERN_DOUBLE_BLINK_PAUSE = 650;  // 650ms pause between cycles
    static const uint16_t PATTERN_SINGLE_BLINK_ON = 500;     // 500ms single blink ON
    static const uint16_t PATTERN_SINGLE_BLINK_OFF = 1000;   // 1000ms single blink OFF
    static const uint16_t PATTERN_SLOW_BLINK_PERIOD = 1000;  // 1000ms slow blink period

    /**
     * Set LED physical state
     * @param on true = LED on, false = LED off
     */
    void setLED(bool on);

    /**
     * Update double blink pattern (complex phase-based state machine)
     * Handles the four phases of the double blink pattern
     */
    void updateDoubleBlink();
};
