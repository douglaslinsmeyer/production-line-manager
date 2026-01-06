#pragma once

#include <Arduino.h>
#include "../profile/profile_manager.h"

/**
 * Production Line State Manager (Refactored for Signal Profiles)
 *
 * Manages production line state using dynamic state names from signal profiles.
 * Supports both legacy enum-based states and profile-based string states.
 *
 * State is:
 * - Loaded from ProfileStorage on boot (current_state key)
 * - Synchronized with MQTT commands from API
 * - Updated on button press using profile button cycles
 * - Persisted via ProfileStorage (not separate NVS namespace)
 *
 * Legacy Compatibility:
 * - Keeps LineState enum for compilation compatibility
 * - Internal logic uses String state names
 */

// Legacy enum (kept for backward compatibility)
enum LineState {
    LINE_STATE_UNKNOWN = 0,
    LINE_STATE_OFF = 1,
    LINE_STATE_ON = 2,
    LINE_STATE_MAINTENANCE = 3,
    LINE_STATE_ERROR = 4
};

// State change callback type (now uses String state names)
typedef void (*StateChangeCallback)(const char* oldState, const char* newState);

class LineStateManager {
public:
    LineStateManager();

    /**
     * Initialize state manager with profile manager reference
     * @param profileMgr Pointer to ProfileManager instance
     */
    void begin(ProfileManager* profileMgr);

    /**
     * Get current production line state as string
     */
    const char* getState() const;

    /**
     * Get state as legacy enum (for backward compatibility)
     * Maps dynamic state names to closest enum value
     */
    LineState getStateLegacy() const;

    /**
     * Set state by name (from MQTT command or manual change)
     * @param stateName Target state name
     * @param source "button", "mqtt", "boot", etc. for logging
     * @return true if state changed
     */
    bool setState(const char* stateName, const char* source = "unknown");

    /**
     * Handle short button press (uses profile's shortPressCycle)
     * @return New state name after transition
     */
    const char* handleShortPress();

    /**
     * Handle long button press (uses profile's longPressCycle)
     * @return New state name after transition
     */
    const char* handleLongPress();

    /**
     * Set callback for state changes
     */
    void setStateChangeCallback(StateChangeCallback callback);

    /**
     * Check if state name is valid (exists in profile)
     * @return true if state is valid
     */
    bool isValidState(const char* stateName) const;

    /**
     * Get ProfileManager reference
     */
    ProfileManager* getProfileManager() { return profileManager; }

    /**
     * Legacy method: Convert state name to enum
     */
    static LineState stringToEnum(const char* stateName);

    /**
     * Legacy method: Convert enum to state name
     */
    static const char* enumToString(LineState state);

private:
    String currentStateName;
    ProfileManager* profileManager;
    StateChangeCallback changeCallback;

    // Helper methods
    void saveState();
    void loadState();
};
