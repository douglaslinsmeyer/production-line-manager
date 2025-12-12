#pragma once

#include <Arduino.h>
#include "state/line_state.h"
#include "digital_output.h"

/**
 * Button LED Controller
 *
 * Controls EXIO5 (TCA9554PWR channel 4) to provide visual feedback
 * of production line state on the control button.
 *
 * LED Patterns:
 * - ON state: Solid on
 * - OFF state: Off
 * - MAINTENANCE state: Blinking (500ms on/off)
 * - ERROR state: Fast blinking (200ms on/off)
 * - UNKNOWN state: Off
 *
 * Uses non-blocking pattern system similar to DeviceIdentification.
 */
class ButtonLED {
public:
    ButtonLED(DigitalOutputManager* outputMgr);

    /**
     * Initialize button LED
     * Ensures EXIO5 is configured as output
     */
    void begin();

    /**
     * Update LED state based on pattern (call in main loop)
     * Non-blocking pattern generation
     */
    void update();

    /**
     * Set LED pattern based on production line state
     * @param state Current production line state
     */
    void setStatePattern(LineState state);

    /**
     * Force LED on/off (for testing or override)
     */
    void setLED(bool on);

private:
    DigitalOutputManager* outputs;

    LineState currentState;
    bool ledState;                   // Current physical LED state (on/off)
    unsigned long lastToggle;        // Last pattern toggle time
    uint16_t currentPeriod;          // Current pattern half-period (ms)

    // Pattern timing constants (LED channel defined in config.h)
    static const uint16_t PATTERN_MAINTENANCE_PERIOD = 500;  // 500ms blink
    static const uint16_t PATTERN_ERROR_PERIOD = 200;        // 200ms fast blink
};
