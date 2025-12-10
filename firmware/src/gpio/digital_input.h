#pragma once

#include <Arduino.h>

// Callback function type for input change events
typedef void (*InputChangeCallback)(uint8_t channel, bool state);

// Digital Input Manager with debouncing
class DigitalInputManager {
public:
    DigitalInputManager();

    // Initialize digital input pins
    void begin();

    // Update inputs (call in loop for debouncing)
    void update();

    // Get individual input state (channel 0-7)
    bool getInput(uint8_t channel);

    // Get all inputs as bitmask (bit 0 = CH1, bit 7 = CH8)
    uint8_t getAllInputs();

    // Set callback for input change events
    void setCallback(InputChangeCallback callback);

private:
    static const uint8_t DIN_PINS[8];

    bool inputState[8];
    bool lastReading[8];
    unsigned long lastDebounceTime[8];
    unsigned long debounceDelay;
    bool bootStabilized;
    unsigned long bootTime;

    InputChangeCallback changeCallback;

    void notifyChange(uint8_t channel, bool state);
};
