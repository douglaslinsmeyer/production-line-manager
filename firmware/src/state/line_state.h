#pragma once

#include <Arduino.h>

/**
 * Production Line State Manager
 *
 * Tracks local production line state (On/Off/Maintenance/Error)
 * and provides state transition logic for control button.
 *
 * State is:
 * - Initialized to UNKNOWN on boot
 * - Synchronized with MQTT commands from API
 * - Updated immediately on button press (firmware authority)
 * - Persisted to NVS for power-cycle resilience
 */

enum LineState {
    LINE_STATE_UNKNOWN = 0,     // Initial state on boot before sync
    LINE_STATE_OFF = 1,          // Production line stopped
    LINE_STATE_ON = 2,           // Production line running
    LINE_STATE_MAINTENANCE = 3,  // Maintenance mode
    LINE_STATE_ERROR = 4         // Error state (set by API only)
};

// State change callback type
typedef void (*StateChangeCallback)(LineState oldState, LineState newState);

class LineStateManager {
public:
    LineStateManager();

    /**
     * Initialize state manager
     * Loads last known state from NVS (if available)
     */
    void begin();

    /**
     * Get current production line state
     */
    LineState getState() const { return currentState; }

    /**
     * Get state as string (for logging/MQTT)
     */
    const char* getStateString() const;
    static const char* stateToString(LineState state);

    /**
     * Set state (from MQTT command or button press)
     * @param newState Target state
     * @param source "button", "mqtt", "boot", etc. for logging
     * @return true if state changed
     */
    bool setState(LineState newState, const char* source = "unknown");

    /**
     * Handle short button press (toggle logic)
     * Implements:
     *   On → Off
     *   Off → On
     *   Maintenance → On
     *   Error → On
     * @return New state after transition
     */
    LineState handleShortPress();

    /**
     * Handle long button press (5+ seconds)
     * Any state → Maintenance
     * @return New state (always MAINTENANCE)
     */
    LineState handleLongPress();

    /**
     * Set callback for state changes
     */
    void setStateChangeCallback(StateChangeCallback callback);

    /**
     * Check if state transitions are allowed
     * (Future: could block certain transitions)
     */
    bool isTransitionAllowed(LineState from, LineState to) const;

private:
    LineState currentState;
    StateChangeCallback changeCallback;

    // NVS persistence
    void saveState();
    void loadState();
};
