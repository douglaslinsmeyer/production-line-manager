#include "line_state.h"

LineStateManager::LineStateManager()
    : currentStateName("Unknown"),
      profileManager(nullptr),
      changeCallback(nullptr) {
}

void LineStateManager::begin(ProfileManager* profileMgr) {
    Serial.println("[LineStateManager] Initializing...");

    profileManager = profileMgr;

    // Load state from ProfileStorage
    loadState();

    // Validate loaded state against profile
    if (profileManager && profileManager->hasProfile()) {
        if (!profileManager->isValidState(currentStateName.c_str())) {
            Serial.printf("[LineStateManager] Warning: Loaded state '%s' not in profile, using default\n",
                         currentStateName.c_str());
            currentStateName = profileManager->getDefaultState();
            saveState();
        }
    }

    Serial.printf("[LineStateManager] Initial state: %s\n", currentStateName.c_str());
}

const char* LineStateManager::getState() const {
    return currentStateName.c_str();
}

LineState LineStateManager::getStateLegacy() const {
    return stringToEnum(currentStateName.c_str());
}

bool LineStateManager::setState(const char* stateName, const char* source) {
    // Check if new state is different
    if (currentStateName.equals(stateName)) {
        return false;  // No change
    }

    // Validate state if profile is loaded
    if (profileManager && profileManager->hasProfile()) {
        if (!profileManager->isValidState(stateName)) {
            Serial.printf("[LineStateManager] Error: Invalid state '%s' for current profile\n", stateName);
            return false;
        }
    }

    String oldStateName = currentStateName;
    currentStateName = stateName;

    Serial.printf("[LineStateManager] State changed: %s -> %s (source: %s)\n",
                 oldStateName.c_str(),
                 currentStateName.c_str(),
                 source);

    // Persist to storage
    saveState();

    // Notify callback
    if (changeCallback != nullptr) {
        changeCallback(oldStateName.c_str(), currentStateName.c_str());
    }

    return true;
}

const char* LineStateManager::handleShortPress() {
    if (!profileManager || !profileManager->hasProfile()) {
        // Legacy behavior if no profile loaded
        Serial.println("[LineStateManager] No profile loaded, using legacy short press behavior");

        if (currentStateName.equals("On")) {
            setState("Off", "button_short");
        } else {
            setState("On", "button_short");
        }

        return currentStateName.c_str();
    }

    // Use profile's short press cycle
    const char* nextState = profileManager->getNextStateInCycle(currentStateName.c_str(), false);

    Serial.printf("[LineStateManager] Short press: %s -> %s\n",
                 currentStateName.c_str(), nextState);

    setState(nextState, "button_short");
    return currentStateName.c_str();
}

const char* LineStateManager::handleLongPress() {
    if (!profileManager || !profileManager->hasProfile()) {
        // Legacy behavior if no profile loaded
        Serial.println("[LineStateManager] No profile loaded, using legacy long press behavior");
        setState("Maintenance", "button_long");
        return currentStateName.c_str();
    }

    // Use profile's long press cycle
    const char* nextState = profileManager->getNextStateInCycle(currentStateName.c_str(), true);

    Serial.printf("[LineStateManager] Long press: %s -> %s\n",
                 currentStateName.c_str(), nextState);

    setState(nextState, "button_long");
    return currentStateName.c_str();
}

void LineStateManager::setStateChangeCallback(StateChangeCallback callback) {
    changeCallback = callback;
}

bool LineStateManager::isValidState(const char* stateName) const {
    if (!profileManager || !profileManager->hasProfile()) {
        // Without profile, accept legacy states
        return (strcmp(stateName, "On") == 0 ||
                strcmp(stateName, "Off") == 0 ||
                strcmp(stateName, "Maintenance") == 0 ||
                strcmp(stateName, "Error") == 0 ||
                strcmp(stateName, "Unknown") == 0);
    }

    return profileManager->isValidState(stateName);
}

void LineStateManager::saveState() {
    // State is now saved via ProfileStorage (current_state key)
    // This method is kept for API compatibility but delegates to ProfileStorage
    if (profileManager && profileManager->hasProfile()) {
        // ProfileStorage is accessed via ProfileManager's storage reference
        // We'll update this when integrating - for now just log
        Serial.printf("[LineStateManager] State should be saved: %s\n", currentStateName.c_str());
    }
}

void LineStateManager::loadState() {
    // State is loaded from ProfileStorage (current_state key)
    // For now, use legacy NVS loading as fallback
    Preferences prefs;
    if (prefs.begin("linestate", true)) {  // Read-only
        uint8_t savedState = prefs.getUChar("current", LINE_STATE_UNKNOWN);
        LineState legacyState = static_cast<LineState>(savedState);
        currentStateName = enumToString(legacyState);
        prefs.end();

        if (legacyState != LINE_STATE_UNKNOWN) {
            Serial.printf("[LineStateManager] Loaded legacy state from NVS: %s\n", currentStateName.c_str());
        }
    } else {
        currentStateName = "Unknown";
    }
}

// ========== Legacy Conversion Methods ==========

LineState LineStateManager::stringToEnum(const char* stateName) {
    if (strcmp(stateName, "On") == 0 || strcmp(stateName, "ON") == 0) {
        return LINE_STATE_ON;
    } else if (strcmp(stateName, "Off") == 0 || strcmp(stateName, "OFF") == 0) {
        return LINE_STATE_OFF;
    } else if (strcmp(stateName, "Maintenance") == 0 || strcmp(stateName, "MAINTENANCE") == 0) {
        return LINE_STATE_MAINTENANCE;
    } else if (strcmp(stateName, "Error") == 0 || strcmp(stateName, "ERROR") == 0) {
        return LINE_STATE_ERROR;
    } else {
        return LINE_STATE_UNKNOWN;
    }
}

const char* LineStateManager::enumToString(LineState state) {
    switch (state) {
        case LINE_STATE_ON:          return "On";
        case LINE_STATE_OFF:         return "Off";
        case LINE_STATE_MAINTENANCE: return "Maintenance";
        case LINE_STATE_ERROR:       return "Error";
        case LINE_STATE_UNKNOWN:
        default:                     return "Unknown";
    }
}
