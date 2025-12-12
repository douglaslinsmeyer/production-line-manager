#include "status_led.h"
#include "config.h"

StatusLEDController::StatusLEDController(DigitalOutputManager* outputMgr)
    : outputs(outputMgr),
      currentStatus(STATUS_NO_NETWORK),
      ledState(false),
      lastToggle(0),
      currentPhase(PHASE_FIRST_ON),
      phaseStartTime(0) {
}

void StatusLEDController::begin() {
    Serial.println("Status LED initialized on DO4 (TCA9554PWR CH3)");

    // Ensure LED is off initially
    setLED(false);
}

void StatusLEDController::update() {
    switch (currentStatus) {
        case STATUS_CONNECTED:
            // Solid on - no update needed, set once in setConnectionStatus()
            break;

        case STATUS_NO_MQTT:
            // Complex double blink pattern
            updateDoubleBlink();
            break;

        case STATUS_NO_NETWORK:
            // Single blink with asymmetric timing
            if (ledState) {
                // LED is currently ON, check if we should turn it OFF
                if (millis() - lastToggle >= PATTERN_SINGLE_BLINK_ON) {
                    setLED(false);
                    ledState = false;
                    lastToggle = millis();
                }
            } else {
                // LED is currently OFF, check if we should turn it ON
                if (millis() - lastToggle >= PATTERN_SINGLE_BLINK_OFF) {
                    setLED(true);
                    ledState = true;
                    lastToggle = millis();
                }
            }
            break;

        case STATUS_AP_MODE:
            // Simple symmetric slow blink
            if (millis() - lastToggle >= PATTERN_SLOW_BLINK_PERIOD) {
                ledState = !ledState;
                setLED(ledState);
                lastToggle = millis();
            }
            break;
    }
}

void StatusLEDController::setConnectionStatus(ConnectionStatus status) {
    if (currentStatus == status) {
        return;  // No change
    }

    // Log state transition
    const char* oldStatusStr;
    const char* newStatusStr;

    switch (currentStatus) {
        case STATUS_CONNECTED:   oldStatusStr = "CONNECTED"; break;
        case STATUS_NO_MQTT:     oldStatusStr = "NO_MQTT"; break;
        case STATUS_NO_NETWORK:  oldStatusStr = "NO_NETWORK"; break;
        case STATUS_AP_MODE:     oldStatusStr = "AP_MODE"; break;
        default:                 oldStatusStr = "UNKNOWN"; break;
    }

    switch (status) {
        case STATUS_CONNECTED:   newStatusStr = "CONNECTED"; break;
        case STATUS_NO_MQTT:     newStatusStr = "NO_MQTT"; break;
        case STATUS_NO_NETWORK:  newStatusStr = "NO_NETWORK"; break;
        case STATUS_AP_MODE:     newStatusStr = "AP_MODE"; break;
        default:                 newStatusStr = "UNKNOWN"; break;
    }

    Serial.printf("Status LED pattern changed: %s -> %s\n", oldStatusStr, newStatusStr);

    // Update state
    currentStatus = status;

    // Reset timing variables for clean transition
    lastToggle = millis();
    currentPhase = PHASE_FIRST_ON;
    phaseStartTime = millis();

    // Set initial LED state based on new pattern
    switch (status) {
        case STATUS_CONNECTED:
            setLED(true);
            ledState = true;
            Serial.println("Status LED: Solid ON (connected)");
            break;

        case STATUS_NO_MQTT:
            setLED(true);  // Start with first blink ON
            ledState = true;
            Serial.println("Status LED: Double blink (network only, no MQTT)");
            break;

        case STATUS_NO_NETWORK:
            setLED(true);  // Start with LED ON
            ledState = true;
            Serial.println("Status LED: Single blink (no network)");
            break;

        case STATUS_AP_MODE:
            setLED(true);  // Start with LED ON
            ledState = true;
            Serial.println("Status LED: Slow blink (AP mode)");
            break;

        default:
            setLED(false);
            ledState = false;
            break;
    }
}

void StatusLEDController::updateDoubleBlink() {
    unsigned long now = millis();
    unsigned long elapsed = now - phaseStartTime;

    switch (currentPhase) {
        case PHASE_FIRST_ON:
            // First blink ON for 150ms
            if (elapsed >= PATTERN_DOUBLE_BLINK_ON) {
                setLED(false);
                ledState = false;
                currentPhase = PHASE_FIRST_OFF;
                phaseStartTime = now;
            }
            break;

        case PHASE_FIRST_OFF:
            // First blink OFF for 150ms
            if (elapsed >= PATTERN_DOUBLE_BLINK_OFF) {
                setLED(true);
                ledState = true;
                currentPhase = PHASE_SECOND_ON;
                phaseStartTime = now;
            }
            break;

        case PHASE_SECOND_ON:
            // Second blink ON for 150ms
            if (elapsed >= PATTERN_DOUBLE_BLINK_ON) {
                setLED(false);
                ledState = false;
                currentPhase = PHASE_PAUSE;
                phaseStartTime = now;
            }
            break;

        case PHASE_PAUSE:
            // Long pause for 650ms
            if (elapsed >= PATTERN_DOUBLE_BLINK_PAUSE) {
                setLED(true);
                ledState = true;
                currentPhase = PHASE_FIRST_ON;
                phaseStartTime = now;
            }
            break;
    }
}

void StatusLEDController::setLED(bool on) {
    if (outputs != nullptr) {
        outputs->setOutput(STATUS_LED_CHANNEL, on);
    }
}
