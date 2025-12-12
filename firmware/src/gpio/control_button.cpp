#include "control_button.h"

ControlButton::ControlButton()
    : pressed(false),
      longPressTriggered(false),
      pressStartTime(0),
      shortPressCallback(nullptr),
      longPressCallback(nullptr) {
}

void ControlButton::begin() {
    // Digital input already initialized by DigitalInputManager
    // This is just for state setup
    Serial.println("Control button initialized on DIN1 (GPIO4)");
    Serial.printf("  Short press: < %lu seconds (toggle state)\n", LONG_PRESS_DURATION / 1000);
    Serial.printf("  Long press: >= %lu seconds (maintenance mode)\n", LONG_PRESS_DURATION / 1000);
}

void ControlButton::handleButtonChange(bool newPressed) {
    if (newPressed && !pressed) {
        // Button just pressed
        pressed = true;
        pressStartTime = millis();
        longPressTriggered = false;
        Serial.println("Control button pressed");
    }
    else if (!newPressed && pressed) {
        // Button just released
        pressed = false;
        uint32_t pressDuration = millis() - pressStartTime;

        Serial.printf("Control button released after %lu ms\n", pressDuration);

        // If released before long press threshold, it's a short press
        if (!longPressTriggered && pressDuration < LONG_PRESS_DURATION) {
            Serial.println("Short press detected");
            if (shortPressCallback != nullptr) {
                shortPressCallback();
            }
        }
        // Long press already triggered in update()
    }
}

void ControlButton::update() {
    // Check if button is held long enough for long press
    if (pressed && !longPressTriggered) {
        uint32_t currentDuration = millis() - pressStartTime;

        if (currentDuration >= LONG_PRESS_DURATION) {
            longPressTriggered = true;
            Serial.println("Long press detected (5 seconds) - Maintenance mode");

            if (longPressCallback != nullptr) {
                longPressCallback();
            }
        }
    }
}

uint32_t ControlButton::getPressDuration() const {
    if (!pressed) {
        return 0;
    }
    return millis() - pressStartTime;
}

void ControlButton::setShortPressCallback(ControlButtonShortPressCallback callback) {
    shortPressCallback = callback;
}

void ControlButton::setLongPressCallback(ControlButtonLongPressCallback callback) {
    longPressCallback = callback;
}
