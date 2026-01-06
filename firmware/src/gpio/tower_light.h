#pragma once

#include <Arduino.h>
#include "../profile/profile_manager.h"
#include "digital_output.h"

/**
 * Tower Light Controller (Refactored for Signal Profiles)
 *
 * Controls stack light with dynamic patterns from signal profiles:
 * - D01 (Channel 0): Red Light
 * - D02 (Channel 1): Yellow Light
 * - D03 (Channel 2): Green Light
 *
 * Supports 4 light modes:
 * - OFF: Light is off
 * - ON: Light is steady on
 * - SHORT_BLINK: 500ms on, 500ms off
 * - LONG_BLINK: 1500ms on, 1500ms off
 *
 * Each light has an independent blink timer for smooth operation.
 * update() must be called in main loop for blink patterns to work.
 */
class TowerLightManager {
public:
    // Tower light channels
    static const uint8_t CHANNEL_RED = 0;
    static const uint8_t CHANNEL_YELLOW = 1;
    static const uint8_t CHANNEL_GREEN = 2;

    // Blink intervals (milliseconds)
    static const uint16_t SHORT_BLINK_INTERVAL = 500;
    static const uint16_t LONG_BLINK_INTERVAL = 1500;

    TowerLightManager(DigitalOutputManager* outputMgr);

    /**
     * Initialize tower lights
     * Sets all lights off initially
     */
    void begin();

    /**
     * Set tower light pattern from profile state outputs
     * @param outputs State output configuration from ProfileManager
     */
    void setFromProfile(const ProfileManager::StateOutputs& outputs);

    /**
     * Update blink patterns (must be called in main loop)
     * Handles non-blocking blink timing for all lights
     */
    void update();

    /**
     * Set all lights off (bypass profile)
     */
    void allLightsOff();

    /**
     * Check if a channel is reserved for tower lights
     * @param channel Output channel (0-7)
     * @return true if channel is tower light (0-2)
     */
    static bool isTowerLightChannel(uint8_t channel);

    /**
     * Manual control (for testing)
     */
    void setLight(uint8_t channel, bool state);

private:
    DigitalOutputManager* outputs;

    // Current configuration
    ProfileManager::LightMode redMode;
    ProfileManager::LightMode yellowMode;
    ProfileManager::LightMode greenMode;

    // Blink state tracking (for non-blocking operation)
    bool redState;
    bool yellowState;
    bool greenState;

    unsigned long lastRedToggle;
    unsigned long lastYellowToggle;
    unsigned long lastGreenToggle;

    // Helper: Update a single light with blink pattern
    void updateLight(uint8_t channel, ProfileManager::LightMode mode,
                    bool& currentState, unsigned long& lastToggle);

    // Helper: Get blink interval for mode
    uint16_t getBlinkInterval(ProfileManager::LightMode mode);
};
