#pragma once

#include <Arduino.h>
#include "config.h"

// Callback types
typedef void (*ControlButtonShortPressCallback)();
typedef void (*ControlButtonLongPressCallback)();

/**
 * Control Button Handler
 *
 * Monitors the production line control button (DIN1/GPIO4) for:
 * - Short press: Toggle production line state (< 5 seconds)
 * - Long press: Set to Maintenance mode (>= 5 seconds)
 *
 * Uses same debounce pattern as BootButton but with different timing.
 */
class ControlButton {
public:
    ControlButton();

    /**
     * Initialize the control button (DIN1/GPIO4)
     * Note: Digital inputs already configured by DigitalInputManager
     * This just sets up state tracking
     */
    void begin();

    /**
     * Handle button state change from DigitalInputManager
     * Call this from the digital input callback when DIN1 changes
     * @param pressed true if button pressed (HIGH), false if released
     */
    void handleButtonChange(bool pressed);

    /**
     * Update press duration tracking (call in main loop)
     * Checks for long press threshold
     */
    void update();

    /**
     * Set callback for short press (button released before 5s)
     */
    void setShortPressCallback(ControlButtonShortPressCallback callback);

    /**
     * Set callback for long press (button held >= 5s)
     */
    void setLongPressCallback(ControlButtonLongPressCallback callback);

    /**
     * Check if button is currently pressed
     */
    bool isPressed() const { return pressed; }

    /**
     * Get current press duration in milliseconds
     */
    uint32_t getPressDuration() const;

private:
    bool pressed;                    // Current pressed state
    bool longPressTriggered;         // Long press already fired
    unsigned long pressStartTime;    // Time when press started

    ControlButtonShortPressCallback shortPressCallback;
    ControlButtonLongPressCallback longPressCallback;

    static const uint32_t LONG_PRESS_DURATION = 5000;  // 5 seconds
};
