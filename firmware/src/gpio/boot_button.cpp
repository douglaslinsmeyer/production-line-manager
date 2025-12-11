#include "boot_button.h"

BootButton::BootButton()
    : pressed(false),
      longPress(false),
      warningGiven(false),
      pressStartTime(0),
      lastDebounceTime(0),
      lastButtonState(HIGH),
      lastStableState(HIGH),
      longPressCallback(nullptr) {
}

void BootButton::begin() {
    pinMode(BUTTON_PIN, INPUT_PULLUP);
    lastButtonState = digitalRead(BUTTON_PIN);
    lastStableState = lastButtonState;

    Serial.printf("Boot button initialized on GPIO%d\n", BUTTON_PIN);
    Serial.printf("  Long press duration: %d seconds\n", LONG_PRESS_DURATION / 1000);
}

void BootButton::update() {
    // Read current button state (LOW = pressed due to pullup)
    bool currentReading = digitalRead(BUTTON_PIN);

    // Check if state changed (for debouncing)
    if (currentReading != lastButtonState) {
        lastDebounceTime = millis();
    }
    lastButtonState = currentReading;

    // If enough time has passed since last change, accept the state
    if ((millis() - lastDebounceTime) > BUTTON_DEBOUNCE_DELAY) {
        // If state has changed from last stable state
        if (currentReading != lastStableState) {
            lastStableState = currentReading;

            // Button pressed (LOW)
            if (currentReading == LOW && !pressed) {
                pressed = true;
                pressStartTime = millis();
                warningGiven = false;
                Serial.println("Boot button pressed");
            }
            // Button released (HIGH)
            else if (currentReading == HIGH && pressed) {
                pressed = false;
                uint32_t pressDuration = millis() - pressStartTime;
                Serial.printf("Boot button released after %lu ms\n", pressDuration);

                // Don't trigger long press callback on release, it's already been triggered
            }
        }
    }

    // If button is pressed, check duration
    if (pressed) {
        uint32_t currentDuration = millis() - pressStartTime;

        // 10-second warning (fast beep)
        if (currentDuration >= WARNING_DURATION && !warningGiven) {
            Serial.println("⚠ Boot button held for 10 seconds - 5 more for AP mode reset");
            warningGiven = true;
            // TODO: Trigger warning beep via DeviceIdentification when integrated
        }

        // 15-second long press (trigger AP mode)
        if (currentDuration >= LONG_PRESS_DURATION && !longPress) {
            longPress = true;
            Serial.println("🔴 LONG PRESS DETECTED - AP MODE TRIGGER");

            // Trigger callback if set
            if (longPressCallback) {
                longPressCallback(currentDuration);
            }

            // TODO: Trigger long beep + red LED via DeviceIdentification when integrated
        }
    }
}

bool BootButton::isPressed() {
    return pressed;
}

uint32_t BootButton::getPressDuration() {
    if (!pressed) {
        return 0;
    }
    return millis() - pressStartTime;
}

void BootButton::setLongPressCallback(BootButtonCallback callback) {
    longPressCallback = callback;
}
