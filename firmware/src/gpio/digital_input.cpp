#include "digital_input.h"
#include "config.h"

// Digital input pin array (DIN CH1-8 = GPIO4-11)
const uint8_t DigitalInputManager::DIN_PINS[8] = {
    DIN_PIN_1, DIN_PIN_2, DIN_PIN_3, DIN_PIN_4,
    DIN_PIN_5, DIN_PIN_6, DIN_PIN_7, DIN_PIN_8
};

DigitalInputManager::DigitalInputManager()
    : debounceDelay(DEBOUNCE_DELAY),
      changeCallback(nullptr),
      bootStabilized(false),
      bootTime(0) {

    for (int i = 0; i < 8; i++) {
        inputState[i] = false;
        lastReading[i] = false;
        lastDebounceTime[i] = 0;
    }
}

void DigitalInputManager::begin() {
    // Initialize all digital input pins
    // No pull-up needed - optocoupler isolated
    for (int i = 0; i < 8; i++) {
        pinMode(DIN_PINS[i], INPUT);
    }

    bootTime = millis();

    Serial.println("Digital inputs initialized (GPIO4-11)");
    Serial.println("WARNING: Waiting for boot stabilization due to ESP32-S3 power-up glitches");
}

void DigitalInputManager::update() {
    // Wait for boot stabilization period to avoid power-up glitches
    // ESP32-S3 Datasheet: GPIO1-20 have 60µs low-level glitches during power-up
    if (!bootStabilized) {
        if (millis() - bootTime < INPUT_READY_DELAY) {
            return;  // Still in stabilization period
        }
        bootStabilized = true;
        Serial.println("Digital inputs ready - boot stabilization complete");
    }

    // Read and debounce all inputs
    for (int i = 0; i < 8; i++) {
        bool reading = digitalRead(DIN_PINS[i]);

        // Check if reading changed
        if (reading != lastReading[i]) {
            lastDebounceTime[i] = millis();
        }

        // If stable for debounce delay, update state
        if ((millis() - lastDebounceTime[i]) > debounceDelay) {
            // Check if state actually changed
            if (reading != inputState[i]) {
                inputState[i] = reading;
                notifyChange(i, reading);
            }
        }

        lastReading[i] = reading;
    }
}

bool DigitalInputManager::getInput(uint8_t channel) {
    if (channel >= 8) return false;
    return inputState[channel];
}

uint8_t DigitalInputManager::getAllInputs() {
    uint8_t state = 0;
    for (int i = 0; i < 8; i++) {
        if (inputState[i]) {
            state |= (1 << i);
        }
    }
    return state;
}

void DigitalInputManager::setCallback(InputChangeCallback callback) {
    changeCallback = callback;
}

void DigitalInputManager::notifyChange(uint8_t channel, bool state) {
    Serial.printf("Input CH%d changed to %s\n",
                 channel + 1, state ? "HIGH" : "LOW");

    if (changeCallback != nullptr) {
        changeCallback(channel, state);
    }
}
