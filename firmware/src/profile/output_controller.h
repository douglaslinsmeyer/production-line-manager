#pragma once

#include <Arduino.h>
#include "profile_manager.h"
#include "../gpio/tower_light.h"
#include "../buzzer/buzzer_controller.h"

/**
 * Output Controller
 *
 * Coordinates all device outputs (tower lights and buzzer) based on
 * signal profile state configurations.
 *
 * Responsibilities:
 * - Look up state output configuration from ProfileManager
 * - Apply configuration to TowerLightManager and BuzzerController
 * - Coordinate update() calls for non-blocking patterns
 *
 * Usage:
 * 1. Initialize with references to all output components
 * 2. Call applyStateOutputs() when state changes
 * 3. Call update() in main loop for blink/chirp patterns
 */
class OutputController {
public:
    OutputController(TowerLightManager* towerLight,
                    BuzzerController* buzzer,
                    ProfileManager* profileMgr);

    /**
     * Initialize output controller
     */
    void begin();

    /**
     * Apply outputs for a specific state
     * Looks up state configuration from profile and applies to hardware
     * @param stateName State name to apply
     * @return true if state found and applied successfully
     */
    bool applyStateOutputs(const char* stateName);

    /**
     * Update all output patterns (must be called in main loop)
     * Handles blink patterns and buzzer chirp
     */
    void update();

    /**
     * Emergency stop all outputs
     */
    void stopAll();

    /**
     * Test mode: Manually control outputs (bypass profile)
     * @param redOn Red light on/off
     * @param yellowOn Yellow light on/off
     * @param greenOn Green light on/off
     * @param buzzerOn Buzzer on/off
     */
    void testOutputs(bool redOn, bool yellowOn, bool greenOn, bool buzzerOn);

private:
    TowerLightManager* towerLight;
    BuzzerController* buzzer;
    ProfileManager* profileManager;
};
