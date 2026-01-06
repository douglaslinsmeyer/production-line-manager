#include "tower_light.h"

TowerLightManager::TowerLightManager(DigitalOutputManager* outputMgr)
    : outputs(outputMgr),
      redMode(ProfileManager::LIGHT_OFF),
      yellowMode(ProfileManager::LIGHT_OFF),
      greenMode(ProfileManager::LIGHT_OFF),
      redState(false),
      yellowState(false),
      greenState(false),
      lastRedToggle(0),
      lastYellowToggle(0),
      lastGreenToggle(0) {
}

void TowerLightManager::begin() {
    Serial.println("[TowerLight] Initializing tower lights...");
    allLightsOff();
    Serial.println("[TowerLight] All lights off (initial state)");
}

void TowerLightManager::setFromProfile(const ProfileManager::StateOutputs& outputs) {
    // Update configuration
    redMode = outputs.redLight;
    yellowMode = outputs.yellowLight;
    greenMode = outputs.greenLight;

    // Reset blink state and timers to start fresh
    redState = false;
    yellowState = false;
    greenState = false;
    lastRedToggle = millis();
    lastYellowToggle = millis();
    lastGreenToggle = millis();

    Serial.printf("[TowerLight] Pattern updated: R=%s Y=%s G=%s\n",
                 ProfileManager::lightModeToString(redMode),
                 ProfileManager::lightModeToString(yellowMode),
                 ProfileManager::lightModeToString(greenMode));

    // Apply initial state immediately
    updateLight(CHANNEL_RED, redMode, redState, lastRedToggle);
    updateLight(CHANNEL_YELLOW, yellowMode, yellowState, lastYellowToggle);
    updateLight(CHANNEL_GREEN, greenMode, greenState, lastGreenToggle);
}

void TowerLightManager::update() {
    // Update each light independently (non-blocking)
    updateLight(CHANNEL_RED, redMode, redState, lastRedToggle);
    updateLight(CHANNEL_YELLOW, yellowMode, yellowState, lastYellowToggle);
    updateLight(CHANNEL_GREEN, greenMode, greenState, lastGreenToggle);
}

void TowerLightManager::allLightsOff() {
    outputs->setOutput(CHANNEL_RED, false);
    outputs->setOutput(CHANNEL_YELLOW, false);
    outputs->setOutput(CHANNEL_GREEN, false);

    redMode = ProfileManager::LIGHT_OFF;
    yellowMode = ProfileManager::LIGHT_OFF;
    greenMode = ProfileManager::LIGHT_OFF;
}

bool TowerLightManager::isTowerLightChannel(uint8_t channel) {
    return (channel == CHANNEL_RED || channel == CHANNEL_YELLOW || channel == CHANNEL_GREEN);
}

void TowerLightManager::setLight(uint8_t channel, bool state) {
    if (isTowerLightChannel(channel)) {
        outputs->setOutput(channel, state);
    }
}

// ========== Private Helper Methods ==========

void TowerLightManager::updateLight(uint8_t channel, ProfileManager::LightMode mode,
                                    bool& currentState, unsigned long& lastToggle) {
    unsigned long now = millis();

    switch (mode) {
        case ProfileManager::LIGHT_OFF:
            // Always off
            if (currentState != false) {
                outputs->setOutput(channel, false);
                currentState = false;
            }
            break;

        case ProfileManager::LIGHT_ON:
            // Always on
            if (currentState != true) {
                outputs->setOutput(channel, true);
                currentState = true;
            }
            break;

        case ProfileManager::LIGHT_SHORT_BLINK:
        case ProfileManager::LIGHT_LONG_BLINK:
            // Blink pattern
            uint16_t interval = getBlinkInterval(mode);

            // Check if it's time to toggle
            if (now - lastToggle >= interval) {
                currentState = !currentState;
                outputs->setOutput(channel, currentState);
                lastToggle = now;
            }
            break;
    }
}

uint16_t TowerLightManager::getBlinkInterval(ProfileManager::LightMode mode) {
    switch (mode) {
        case ProfileManager::LIGHT_SHORT_BLINK:
            return SHORT_BLINK_INTERVAL;
        case ProfileManager::LIGHT_LONG_BLINK:
            return LONG_BLINK_INTERVAL;
        default:
            return 1000;  // Fallback
    }
}
