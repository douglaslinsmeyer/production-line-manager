#include "tower_light.h"
#include "config.h"

TowerLightManager::TowerLightManager(DigitalOutputManager* outputMgr)
    : outputs(outputMgr),
      currentState(LINE_STATE_UNKNOWN) {
}

void TowerLightManager::begin() {
    Serial.println("Tower Light Manager initialized (D01=Red, D02=Yellow, D03=Green)");

    // Start with all lights off
    allLightsOff();
}

void TowerLightManager::setStatePattern(LineState state) {
    if (currentState == state) {
        return;  // No change needed
    }

    Serial.printf("Tower Light pattern changed: %s -> %s\n",
                 LineStateManager::stateToString(currentState),
                 LineStateManager::stateToString(state));

    currentState = state;

    // Turn off all lights first
    allLightsOff();

    // Set appropriate light based on state
    switch (state) {
        case LINE_STATE_ON:
            // Green light only
            outputs->setOutput(TOWER_LIGHT_GREEN_CHANNEL, true);
            Serial.println("Tower Light: GREEN (Production Running)");
            break;

        case LINE_STATE_OFF:
            // Red light only
            outputs->setOutput(TOWER_LIGHT_RED_CHANNEL, true);
            Serial.println("Tower Light: RED (Production Stopped)");
            break;

        case LINE_STATE_MAINTENANCE:
            // Yellow light only
            outputs->setOutput(TOWER_LIGHT_YELLOW_CHANNEL, true);
            Serial.println("Tower Light: YELLOW (Maintenance Mode)");
            break;

        case LINE_STATE_ERROR:
            // Red light only
            outputs->setOutput(TOWER_LIGHT_RED_CHANNEL, true);
            Serial.println("Tower Light: RED (Error State)");
            break;

        case LINE_STATE_UNKNOWN:
        default:
            // All lights off
            Serial.println("Tower Light: OFF (Unknown/Boot State)");
            break;
    }
}

bool TowerLightManager::isTowerLightChannel(uint8_t channel) {
    return (channel == TOWER_LIGHT_RED_CHANNEL ||
            channel == TOWER_LIGHT_YELLOW_CHANNEL ||
            channel == TOWER_LIGHT_GREEN_CHANNEL);
}

void TowerLightManager::allLightsOff() {
    if (outputs != nullptr) {
        outputs->setOutput(TOWER_LIGHT_RED_CHANNEL, false);
        outputs->setOutput(TOWER_LIGHT_YELLOW_CHANNEL, false);
        outputs->setOutput(TOWER_LIGHT_GREEN_CHANNEL, false);
    }
}
