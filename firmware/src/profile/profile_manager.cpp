#include "profile_manager.h"

ProfileManager::ProfileManager(ProfileStorage* storage)
    : storage(storage), profileLoaded(false), profileVersion(0) {
}

bool ProfileManager::loadProfile() {
    String profileJson;

    if (!storage->loadProfile(profileJson)) {
        Serial.println("[ProfileManager] No profile found in storage");
        profileLoaded = false;
        return false;
    }

    // Parse JSON
    DeserializationError error = deserializeJson(profileDoc, profileJson);

    if (error) {
        Serial.printf("[ProfileManager] JSON parse error: %s\n", error.c_str());
        profileLoaded = false;
        return false;
    }

    // Parse and validate profile structure
    if (!parseProfile()) {
        profileLoaded = false;
        return false;
    }

    profileLoaded = true;
    Serial.printf("[ProfileManager] Profile loaded: %s v%d with %d states\n",
                 profileName.c_str(), profileVersion, getStateCount());

    return true;
}

bool ProfileManager::updateProfile(const String& profileJson, const String& profileId, uint32_t version) {
    // Save to storage first
    if (!storage->saveProfile(profileJson, profileId, version)) {
        Serial.println("[ProfileManager] Failed to save profile to storage");
        return false;
    }

    // Parse JSON
    DeserializationError error = deserializeJson(profileDoc, profileJson);

    if (error) {
        Serial.printf("[ProfileManager] JSON parse error: %s\n", error.c_str());
        return false;
    }

    // Parse and validate
    if (!parseProfile()) {
        return false;
    }

    profileLoaded = true;
    Serial.printf("[ProfileManager] Profile updated: %s v%d\n", profileName.c_str(), version);

    return true;
}

bool ProfileManager::parseProfile() {
    // Extract metadata
    if (!profileDoc["id"].isNull()) {
        profileId = profileDoc["id"].as<String>();
    } else {
        profileId = storage->getProfileId();
    }

    if (!profileDoc["version"].isNull()) {
        profileVersion = profileDoc["version"].as<uint32_t>();
    } else {
        profileVersion = storage->getProfileVersion();
    }

    if (!profileDoc["name"].isNull()) {
        profileName = profileDoc["name"].as<String>();
    }

    if (!profileDoc["defaultState"].isNull()) {
        defaultState = profileDoc["defaultState"].as<String>();
    } else {
        Serial.println("[ProfileManager] Error: No defaultState in profile");
        return false;
    }

    // Validate required fields
    if (profileDoc["states"].isNull() || !profileDoc["states"].is<JsonArray>()) {
        Serial.println("[ProfileManager] Error: Missing or invalid 'states' array");
        return false;
    }

    if (profileDoc["buttonBehavior"].isNull() || !profileDoc["buttonBehavior"].is<JsonObject>()) {
        Serial.println("[ProfileManager] Error: Missing or invalid 'buttonBehavior' object");
        return false;
    }

    // Validate that default state exists
    if (!isValidState(defaultState.c_str())) {
        Serial.printf("[ProfileManager] Error: Default state '%s' not found in states array\n",
                     defaultState.c_str());
        return false;
    }

    return true;
}

bool ProfileManager::hasProfile() {
    return profileLoaded;
}

bool ProfileManager::isValidState(const char* stateName) {
    if (!profileLoaded) {
        return false;
    }

    JsonArrayConst states = profileDoc["states"];
    for (JsonObjectConst state : states) {
        if (strcmp(state["name"].as<const char*>(), stateName) == 0) {
            return true;
        }
    }

    return false;
}

const char* ProfileManager::getDefaultState() {
    if (!profileLoaded) {
        return "";
    }
    return defaultState.c_str();
}

bool ProfileManager::getStateOutputs(const char* stateName, StateOutputs& outputs) {
    if (!profileLoaded) {
        return false;
    }

    JsonArrayConst states = profileDoc["states"];
    for (JsonObjectConst state : states) {
        if (strcmp(state["name"].as<const char*>(), stateName) == 0) {
            // Found the state, parse outputs
            JsonObjectConst stateOutputs = state["outputs"];

            outputs.redLight = parseLightMode(stateOutputs["redLight"]);
            outputs.yellowLight = parseLightMode(stateOutputs["yellowLight"]);
            outputs.greenLight = parseLightMode(stateOutputs["greenLight"]);
            outputs.buzzer = parseBuzzerMode(stateOutputs["buzzer"]);

            return true;
        }
    }

    return false;
}

const char* ProfileManager::getNextStateInCycle(const char* currentState, bool longPress) {
    if (!profileLoaded) {
        return currentState;
    }

    JsonObjectConst buttonBehavior = profileDoc["buttonBehavior"];
    JsonArrayConst cycle;

    if (longPress) {
        cycle = buttonBehavior["longPressCycle"];
    } else {
        cycle = buttonBehavior["shortPressCycle"];
    }

    if (cycle.size() == 0) {
        Serial.printf("[ProfileManager] Warning: Empty %s press cycle\n",
                     longPress ? "long" : "short");
        return currentState;
    }

    // Find current state in cycle
    int currentIndex = findCurrentIndexInCycle(currentState, cycle);

    if (currentIndex == -1) {
        // Current state not in cycle, return first state in cycle
        Serial.printf("[ProfileManager] Current state '%s' not in cycle, returning first: %s\n",
                     currentState, cycle[0].as<const char*>());
        return cycle[0].as<const char*>();
    }

    // Return next state (wrap to 0 if at end)
    int nextIndex = (currentIndex + 1) % cycle.size();
    return cycle[nextIndex].as<const char*>();
}

String ProfileManager::getProfileId() {
    return profileId;
}

uint32_t ProfileManager::getProfileVersion() {
    return profileVersion;
}

String ProfileManager::getProfileName() {
    return profileName;
}

int ProfileManager::getStateCount() {
    if (!profileLoaded) {
        return 0;
    }
    return profileDoc["states"].as<JsonArray>().size();
}

void ProfileManager::clearProfile() {
    profileDoc.clear();
    profileLoaded = false;
    profileId = "";
    profileVersion = 0;
    profileName = "";
    defaultState = "";

    Serial.println("[ProfileManager] Profile cleared from memory");
}

// ========== Helper Methods ==========

ProfileManager::LightMode ProfileManager::parseLightMode(const char* mode) {
    if (strcmp(mode, "on") == 0) {
        return LIGHT_ON;
    } else if (strcmp(mode, "shortBlink") == 0) {
        return LIGHT_SHORT_BLINK;
    } else if (strcmp(mode, "longBlink") == 0) {
        return LIGHT_LONG_BLINK;
    } else {
        return LIGHT_OFF;
    }
}

ProfileManager::BuzzerMode ProfileManager::parseBuzzerMode(const char* mode) {
    if (strcmp(mode, "on") == 0) {
        return BUZZER_ON;
    } else if (strcmp(mode, "chirp") == 0) {
        return BUZZER_CHIRP;
    } else {
        return BUZZER_OFF;
    }
}

int ProfileManager::findStateIndex(const char* stateName) {
    if (!profileLoaded) {
        return -1;
    }

    JsonArrayConst states = profileDoc["states"];
    int index = 0;
    for (JsonObjectConst state : states) {
        if (strcmp(state["name"].as<const char*>(), stateName) == 0) {
            return index;
        }
        index++;
    }

    return -1;
}

int ProfileManager::findCurrentIndexInCycle(const char* currentState, JsonArrayConst cycle) {
    for (size_t i = 0; i < cycle.size(); i++) {
        if (strcmp(cycle[i].as<const char*>(), currentState) == 0) {
            return i;
        }
    }
    return -1;
}

const char* ProfileManager::lightModeToString(LightMode mode) {
    switch (mode) {
        case LIGHT_OFF: return "off";
        case LIGHT_ON: return "on";
        case LIGHT_SHORT_BLINK: return "shortBlink";
        case LIGHT_LONG_BLINK: return "longBlink";
        default: return "unknown";
    }
}

const char* ProfileManager::buzzerModeToString(BuzzerMode mode) {
    switch (mode) {
        case BUZZER_OFF: return "off";
        case BUZZER_ON: return "on";
        case BUZZER_CHIRP: return "chirp";
        default: return "unknown";
    }
}
