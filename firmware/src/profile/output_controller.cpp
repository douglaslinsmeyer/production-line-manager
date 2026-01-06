#include "output_controller.h"

OutputController::OutputController(TowerLightManager* towerLight,
                                   BuzzerController* buzzer,
                                   ProfileManager* profileMgr)
    : towerLight(towerLight),
      buzzer(buzzer),
      profileManager(profileMgr) {
}

void OutputController::begin() {
    Serial.println("[OutputController] Initialized");
}

bool OutputController::applyStateOutputs(const char* stateName) {
    if (!profileManager || !profileManager->hasProfile()) {
        Serial.println("[OutputController] No profile loaded, cannot apply state outputs");
        return false;
    }

    // Get state output configuration from profile
    ProfileManager::StateOutputs outputs;
    if (!profileManager->getStateOutputs(stateName, outputs)) {
        Serial.printf("[OutputController] Error: State '%s' not found in profile\n", stateName);
        return false;
    }

    Serial.printf("[OutputController] Applying outputs for state '%s':\n", stateName);
    Serial.printf("  Red: %s, Yellow: %s, Green: %s, Buzzer: %s\n",
                 ProfileManager::lightModeToString(outputs.redLight),
                 ProfileManager::lightModeToString(outputs.yellowLight),
                 ProfileManager::lightModeToString(outputs.greenLight),
                 ProfileManager::buzzerModeToString(outputs.buzzer));

    // Apply to tower lights
    towerLight->setFromProfile(outputs);

    // Apply to buzzer
    buzzer->setMode(outputs.buzzer);

    return true;
}

void OutputController::update() {
    // Update all pattern timers
    towerLight->update();
    buzzer->update();
}

void OutputController::stopAll() {
    Serial.println("[OutputController] Emergency stop - all outputs off");
    towerLight->allLightsOff();
    buzzer->stop();
}

void OutputController::testOutputs(bool redOn, bool yellowOn, bool greenOn, bool buzzerOn) {
    Serial.println("[OutputController] Test mode - manual control");

    // Direct hardware control (bypass profile)
    towerLight->setLight(TowerLightManager::CHANNEL_RED, redOn);
    towerLight->setLight(TowerLightManager::CHANNEL_YELLOW, yellowOn);
    towerLight->setLight(TowerLightManager::CHANNEL_GREEN, greenOn);

    if (buzzerOn) {
        buzzer->setMode(ProfileManager::BUZZER_ON);
    } else {
        buzzer->setMode(ProfileManager::BUZZER_OFF);
    }
}
