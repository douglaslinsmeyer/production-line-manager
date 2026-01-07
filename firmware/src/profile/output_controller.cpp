#include "output_controller.h"

OutputController::OutputController(TowerLightManager* towerLight,
                                   BuzzerController* primaryBuzzer,
                                   TowerBuzzerController* towerBuzzer,
                                   ProfileManager* profileMgr)
    : towerLight(towerLight),
      primaryBuzzer(primaryBuzzer),
      towerBuzzer(towerBuzzer),
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
    Serial.printf("  Red: %s, Yellow: %s, Green: %s\n",
                 ProfileManager::lightModeToString(outputs.redLight),
                 ProfileManager::lightModeToString(outputs.yellowLight),
                 ProfileManager::lightModeToString(outputs.greenLight));
    Serial.printf("  Primary Buzzer: %s, Tower Buzzer: %s\n",
                 ProfileManager::buzzerModeToString(outputs.primaryBuzzer),
                 ProfileManager::buzzerModeToString(outputs.towerBuzzer));

    // Apply to tower lights
    towerLight->setFromProfile(outputs);

    // Apply to both buzzers
    primaryBuzzer->setMode(outputs.primaryBuzzer);
    towerBuzzer->setMode(outputs.towerBuzzer);

    return true;
}

void OutputController::update() {
    // Update all pattern timers
    towerLight->update();
    primaryBuzzer->update();
    towerBuzzer->update();
}

void OutputController::stopAll() {
    Serial.println("[OutputController] Emergency stop - all outputs off");
    towerLight->allLightsOff();
    primaryBuzzer->stop();
    towerBuzzer->stop();
}

void OutputController::testOutputs(bool redOn, bool yellowOn, bool greenOn, bool primaryBuzzerOn, bool towerBuzzerOn) {
    Serial.println("[OutputController] Test mode - manual control");
    Serial.printf("  Red: %s, Yellow: %s, Green: %s\n",
                  redOn ? "ON" : "OFF",
                  yellowOn ? "ON" : "OFF",
                  greenOn ? "ON" : "OFF");
    Serial.printf("  Primary Buzzer: %s, Tower Buzzer: %s\n",
                  primaryBuzzerOn ? "ON" : "OFF",
                  towerBuzzerOn ? "ON" : "OFF");

    // Create manual state outputs using mode system (not direct GPIO)
    // This allows the update loop to maintain the test state instead of overwriting it
    ProfileManager::StateOutputs outputs;
    outputs.redLight = redOn ? ProfileManager::LIGHT_ON : ProfileManager::LIGHT_OFF;
    outputs.yellowLight = yellowOn ? ProfileManager::LIGHT_ON : ProfileManager::LIGHT_OFF;
    outputs.greenLight = greenOn ? ProfileManager::LIGHT_ON : ProfileManager::LIGHT_OFF;
    outputs.primaryBuzzer = primaryBuzzerOn ? ProfileManager::BUZZER_ON : ProfileManager::BUZZER_OFF;
    outputs.towerBuzzer = towerBuzzerOn ? ProfileManager::BUZZER_ON : ProfileManager::BUZZER_OFF;

    // Apply using the normal mode system - update loop will maintain this state
    towerLight->setFromProfile(outputs);
    primaryBuzzer->setMode(outputs.primaryBuzzer);
    towerBuzzer->setMode(outputs.towerBuzzer);
}
