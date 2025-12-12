#include "button_led.h"
#include "config.h"

ButtonLED::ButtonLED(DigitalOutputManager* outputMgr)
    : outputs(outputMgr),
      currentState(LINE_STATE_UNKNOWN),
      ledState(false),
      lastToggle(0),
      currentPeriod(0) {
}

void ButtonLED::begin() {
    Serial.println("Button LED initialized on EXIO5 (TCA9554PWR CH4)");

    // Ensure LED is off initially
    setLED(false);
}

void ButtonLED::update() {
    // Non-blinking states (solid or off)
    if (currentState == LINE_STATE_ON) {
        // Solid on
        if (!ledState) {
            setLED(true);
            ledState = true;
        }
        return;
    }

    if (currentState == LINE_STATE_OFF ||
        currentState == LINE_STATE_UNKNOWN) {
        // Off
        if (ledState) {
            setLED(false);
            ledState = false;
        }
        return;
    }

    // Blinking states (MAINTENANCE, ERROR)
    if (currentPeriod == 0) {
        return;  // No pattern set
    }

    unsigned long now = millis();
    if (now - lastToggle >= currentPeriod) {
        lastToggle = now;
        ledState = !ledState;
        setLED(ledState);
    }
}

void ButtonLED::setStatePattern(LineState state) {
    if (currentState == state) {
        return;  // No change
    }

    Serial.printf("Button LED pattern changed: %s -> %s\n",
                 LineStateManager::stateToString(currentState),
                 LineStateManager::stateToString(state));

    currentState = state;
    lastToggle = millis();
    ledState = false;

    // Set pattern period based on state
    switch (state) {
        case LINE_STATE_ON:
            currentPeriod = 0;  // Solid on (handled in update())
            setLED(true);
            Serial.println("Button LED: Solid ON");
            break;

        case LINE_STATE_OFF:
        case LINE_STATE_UNKNOWN:
            currentPeriod = 0;  // Off
            setLED(false);
            Serial.println("Button LED: OFF");
            break;

        case LINE_STATE_MAINTENANCE:
            currentPeriod = PATTERN_MAINTENANCE_PERIOD;
            Serial.println("Button LED: Blinking (maintenance)");
            break;

        case LINE_STATE_ERROR:
            currentPeriod = PATTERN_ERROR_PERIOD;
            Serial.println("Button LED: Fast blinking (error)");
            break;

        default:
            currentPeriod = 0;
            setLED(false);
            break;
    }
}

void ButtonLED::setLED(bool on) {
    if (outputs != nullptr) {
        outputs->setOutput(BUTTON_LED_CHANNEL, on);
    }
}
