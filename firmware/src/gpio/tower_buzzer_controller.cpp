#include "tower_buzzer_controller.h"

TowerBuzzerController::TowerBuzzerController(DigitalOutputManager* outputMgr, uint8_t channel)
    : outputs(outputMgr),
      buzzerChannel(channel),
      currentMode(ProfileManager::BUZZER_OFF),
      chirpStep(CHIRP_BEEP_1),
      chirpStepStart(0) {
}

void TowerBuzzerController::begin() {
    setBuzzer(false);
    Serial.printf("[TowerBuzzerController] Initialized on DO%d (channel %d)\n", buzzerChannel + 1, buzzerChannel);
}

void TowerBuzzerController::setMode(ProfileManager::BuzzerMode mode) {
    if (currentMode == mode) {
        return;  // No change
    }

    currentMode = mode;

    Serial.printf("[TowerBuzzerController] Mode changed to: %s\n",
                 ProfileManager::buzzerModeToString(mode));

    switch (mode) {
        case ProfileManager::BUZZER_OFF:
            setBuzzer(false);
            break;

        case ProfileManager::BUZZER_ON:
            setBuzzer(true);
            break;

        case ProfileManager::BUZZER_CHIRP:
            // Reset chirp pattern to beginning
            chirpStep = CHIRP_BEEP_1;
            chirpStepStart = millis();
            setBuzzer(true);  // Start first beep immediately
            break;
    }
}

void TowerBuzzerController::update() {
    // Only chirp mode requires updates
    if (currentMode == ProfileManager::BUZZER_CHIRP) {
        updateChirpPattern();
    }
}

void TowerBuzzerController::stop() {
    currentMode = ProfileManager::BUZZER_OFF;
    setBuzzer(false);
}

// ========== Private Helper Methods ==========

void TowerBuzzerController::setBuzzer(bool state) {
    outputs->setOutput(buzzerChannel, state);
}

void TowerBuzzerController::updateChirpPattern() {
    unsigned long now = millis();
    unsigned long elapsed = now - chirpStepStart;

    switch (chirpStep) {
        case CHIRP_BEEP_1:
            // First beep (150ms)
            if (elapsed >= CHIRP_BEEP_DURATION) {
                setBuzzer(false);
                chirpStep = CHIRP_PAUSE_1;
                chirpStepStart = now;
            }
            break;

        case CHIRP_PAUSE_1:
            // First pause (150ms)
            if (elapsed >= CHIRP_PAUSE_DURATION) {
                setBuzzer(true);
                chirpStep = CHIRP_BEEP_2;
                chirpStepStart = now;
            }
            break;

        case CHIRP_BEEP_2:
            // Second beep (150ms)
            if (elapsed >= CHIRP_BEEP_DURATION) {
                setBuzzer(false);
                chirpStep = CHIRP_PAUSE_2;
                chirpStepStart = now;
            }
            break;

        case CHIRP_PAUSE_2:
            // Second pause (150ms)
            if (elapsed >= CHIRP_PAUSE_DURATION) {
                setBuzzer(true);
                chirpStep = CHIRP_BEEP_3;
                chirpStepStart = now;
            }
            break;

        case CHIRP_BEEP_3:
            // Third beep (150ms)
            if (elapsed >= CHIRP_BEEP_DURATION) {
                setBuzzer(false);
                chirpStep = CHIRP_DELAY;
                chirpStepStart = now;
            }
            break;

        case CHIRP_DELAY:
            // Long pause (3000ms)
            if (elapsed >= CHIRP_DELAY_DURATION) {
                setBuzzer(true);
                chirpStep = CHIRP_BEEP_1;
                chirpStepStart = now;
            }
            break;
    }
}
