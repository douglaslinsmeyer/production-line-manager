#include "line_state.h"
#include <Preferences.h>

// NVS namespace for state persistence
static const char* NVS_NAMESPACE = "linestate";
static const char* NVS_STATE_KEY = "current";

LineStateManager::LineStateManager()
    : currentState(LINE_STATE_UNKNOWN),
      changeCallback(nullptr) {
}

void LineStateManager::begin() {
    Serial.println("Initializing Line State Manager...");

    // Load last known state from NVS
    loadState();

    Serial.printf("Initial state: %s\n", getStateString());
}

const char* LineStateManager::getStateString() const {
    return stateToString(currentState);
}

const char* LineStateManager::stateToString(LineState state) {
    switch (state) {
        case LINE_STATE_UNKNOWN:     return "UNKNOWN";
        case LINE_STATE_OFF:         return "OFF";
        case LINE_STATE_ON:          return "ON";
        case LINE_STATE_MAINTENANCE: return "MAINTENANCE";
        case LINE_STATE_ERROR:       return "ERROR";
        default:                     return "INVALID";
    }
}

bool LineStateManager::setState(LineState newState, const char* source) {
    if (currentState == newState) {
        return false;  // No change
    }

    // Check if transition is allowed
    if (!isTransitionAllowed(currentState, newState)) {
        Serial.printf("State transition blocked: %s -> %s\n",
                     stateToString(currentState),
                     stateToString(newState));
        return false;
    }

    LineState oldState = currentState;
    currentState = newState;

    Serial.printf("State changed: %s -> %s (source: %s)\n",
                 stateToString(oldState),
                 stateToString(newState),
                 source);

    // Persist to NVS
    saveState();

    // Notify callback
    if (changeCallback != nullptr) {
        changeCallback(oldState, newState);
    }

    return true;
}

LineState LineStateManager::handleShortPress() {
    LineState newState;

    switch (currentState) {
        case LINE_STATE_ON:
            newState = LINE_STATE_OFF;
            break;

        case LINE_STATE_OFF:
        case LINE_STATE_MAINTENANCE:
        case LINE_STATE_ERROR:
        case LINE_STATE_UNKNOWN:
            newState = LINE_STATE_ON;
            break;

        default:
            newState = LINE_STATE_ON;
            break;
    }

    setState(newState, "button_short");
    return newState;
}

LineState LineStateManager::handleLongPress() {
    setState(LINE_STATE_MAINTENANCE, "button_long");
    return LINE_STATE_MAINTENANCE;
}

void LineStateManager::setStateChangeCallback(StateChangeCallback callback) {
    changeCallback = callback;
}

bool LineStateManager::isTransitionAllowed(LineState from, LineState to) const {
    // Currently all transitions are allowed
    // Future: could block transitions like ERROR->ON without clearance
    return true;
}

void LineStateManager::saveState() {
    Preferences prefs;
    if (prefs.begin(NVS_NAMESPACE, false)) {  // Read-write mode
        prefs.putUChar(NVS_STATE_KEY, static_cast<uint8_t>(currentState));
        prefs.end();
    } else {
        Serial.println("Failed to save state to NVS");
    }
}

void LineStateManager::loadState() {
    Preferences prefs;
    if (prefs.begin(NVS_NAMESPACE, true)) {  // Read-only mode
        uint8_t savedState = prefs.getUChar(NVS_STATE_KEY, LINE_STATE_UNKNOWN);
        currentState = static_cast<LineState>(savedState);
        prefs.end();

        if (currentState == LINE_STATE_UNKNOWN) {
            Serial.println("No saved state found in NVS");
        } else {
            Serial.printf("Loaded state from NVS: %s\n", getStateString());
        }
    } else {
        Serial.println("Failed to load state from NVS");
        currentState = LINE_STATE_UNKNOWN;
    }
}
