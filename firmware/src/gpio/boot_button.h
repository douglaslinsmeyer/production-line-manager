#pragma once

#include <Arduino.h>
#include "config.h"

// Callback type for long press event
typedef void (*BootButtonCallback)(uint32_t pressDuration);

/**
 * Boot Button Handler
 *
 * Monitors the BOOT button (GPIO0) on ESP32-S3 for long press detection.
 * Provides visual/audio feedback and triggers AP mode reset after 15 seconds.
 */
class BootButton {
public:
    BootButton();

    /**
     * Initialize the boot button (GPIO0)
     * Sets up INPUT_PULLUP mode (button is LOW when pressed)
     */
    void begin();

    /**
     * Update button state (call in main loop)
     * Handles debouncing, press duration tracking, and callback triggering
     */
    void update();

    /**
     * Check if button is currently pressed
     * @return true if button is pressed (LOW), false otherwise
     */
    bool isPressed();

    /**
     * Get current press duration in milliseconds
     * @return Duration button has been held, 0 if not pressed
     */
    uint32_t getPressDuration();

    /**
     * Set callback for long press event (15 seconds)
     * Callback is triggered once when LONG_PRESS_DURATION is reached
     * @param callback Function to call on long press
     */
    void setLongPressCallback(BootButtonCallback callback);

    /**
     * Check if long press was detected this cycle
     * @return true if long press flag is set
     */
    bool longPressDetected() const { return longPress; }

    /**
     * Reset long press flag after handling
     */
    void resetLongPress() { longPress = false; }

private:
    bool pressed;                   // Current pressed state
    bool longPress;                 // Long press detected flag
    bool warningGiven;              // 10-second warning flag
    unsigned long pressStartTime;   // Time when press started
    unsigned long lastDebounceTime; // Last time button state changed
    bool lastButtonState;           // Last raw button state
    bool lastStableState;           // Last debounced stable state

    BootButtonCallback longPressCallback;

    static const uint8_t BUTTON_PIN = BOOT_BUTTON_PIN;
    static const uint32_t BUTTON_DEBOUNCE_DELAY = 50;  // 50ms debounce
    static const uint32_t LONG_PRESS_DURATION = BOOT_BUTTON_LONG_PRESS;  // 15 seconds
    static const uint32_t WARNING_DURATION = BOOT_BUTTON_WARNING_TIME;    // 10 seconds
};
