#pragma once

#include <Arduino.h>
#include "state/line_state.h"
#include "digital_output.h"

/**
 * Tower Light Controller
 *
 * Controls stack light (D01=Red, D02=Yellow, D03=Green) to provide
 * visual indication of production line state.
 *
 * Light Patterns (automatic, not controllable via MQTT):
 * - ON state: Green light only (D03)
 * - OFF state: Red light only (D01)
 * - MAINTENANCE state: Yellow light only (D02)
 * - ERROR state: Red light only (D01)
 * - UNKNOWN state: All lights off (boot state)
 *
 * Tower light channels (0-2) are reserved and cannot be controlled
 * via MQTT set_output commands.
 */
class TowerLightManager {
public:
    TowerLightManager(DigitalOutputManager* outputMgr);

    /**
     * Initialize tower lights
     * Sets all lights off initially
     */
    void begin();

    /**
     * Update tower lights based on production line state
     * @param state Current production line state
     */
    void setStatePattern(LineState state);

    /**
     * Check if a channel is reserved for tower lights
     * @param channel Output channel (0-7)
     * @return true if channel is tower light (0-2)
     */
    static bool isTowerLightChannel(uint8_t channel);

private:
    DigitalOutputManager* outputs;
    LineState currentState;

    // Helper to set all lights off
    void allLightsOff();
};
