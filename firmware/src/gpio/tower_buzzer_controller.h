#pragma once

#include <Arduino.h>
#include "../profile/profile_manager.h"
#include "digital_output.h"

/**
 * Tower Buzzer Controller (for Signal Profiles)
 *
 * Controls tower buzzer on DO4 via TCA9554PWR I2C GPIO expander.
 * Supports same patterns as primary buzzer controller:
 * - OFF: Silent
 * - ON: Continuous tone
 * - CHIRP: beep (150ms), pause (150ms), beep, pause, beep, long pause (3000ms), repeat
 *
 * Uses non-blocking state machine for chirp pattern.
 * update() must be called in main loop for chirp pattern to work.
 */
class TowerBuzzerController {
public:
    // Chirp pattern timing constants (same as BuzzerController)
    static const uint16_t CHIRP_BEEP_DURATION = 150;   // ms
    static const uint16_t CHIRP_PAUSE_DURATION = 150;  // ms
    static const uint16_t CHIRP_DELAY_DURATION = 3000; // ms between chirp sequences

    TowerBuzzerController(DigitalOutputManager* outputMgr, uint8_t channel);

    /**
     * Initialize tower buzzer
     * Sets buzzer off initially
     */
    void begin();

    /**
     * Set buzzer mode
     * @param mode Buzzer mode (OFF, ON, CHIRP)
     */
    void setMode(ProfileManager::BuzzerMode mode);

    /**
     * Update buzzer pattern (must be called in main loop)
     * Handles non-blocking chirp pattern state machine
     */
    void update();

    /**
     * Get current mode
     */
    ProfileManager::BuzzerMode getMode() const { return currentMode; }

    /**
     * Emergency stop buzzer (sets to OFF immediately)
     */
    void stop();

private:
    DigitalOutputManager* outputs;
    uint8_t buzzerChannel;
    ProfileManager::BuzzerMode currentMode;

    // Chirp pattern state machine
    enum ChirpStep {
        CHIRP_BEEP_1 = 0,   // First beep
        CHIRP_PAUSE_1 = 1,  // First pause
        CHIRP_BEEP_2 = 2,   // Second beep
        CHIRP_PAUSE_2 = 3,  // Second pause
        CHIRP_BEEP_3 = 4,   // Third beep
        CHIRP_DELAY = 5     // Long pause before repeat
    };

    ChirpStep chirpStep;
    unsigned long chirpStepStart;

    // Helper methods
    void setBuzzer(bool state);
    void updateChirpPattern();
};
