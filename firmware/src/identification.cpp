#include "identification.h"

DeviceIdentification::DeviceIdentification()
    : flashing(false),
      flashEndTime(0),
      currentPattern(LED_PATTERN_OFF),
      ledState(false),
      lastPatternToggle(0),
      currentPatternPeriod(0) {
}

void DeviceIdentification::begin() {
    // Configure RGB LED PWM channels
    ledcAttach(GPIO_RGB_LED, PWM_FREQ, PWM_RESOLUTION);

    // Configure buzzer PWM
    ledcAttach(GPIO_BUZZER, 1000, PWM_RESOLUTION);  // 1kHz for buzzer

    // Turn off initially
    setRGB(0, 0, 0);
    setBuzzer(false);

    Serial.println("Device identification (LED + Buzzer) initialized");
}

void DeviceIdentification::update() {
    // Handle pattern-based LED control
    if (currentPattern == LED_PATTERN_OFF) {
        // Ensure LED is off
        if (ledState) {
            setRGB(0, 0, 0);
            ledState = false;
        }
        return;
    }

    // Check if it's time to toggle
    unsigned long now = millis();
    if (now - lastPatternToggle >= currentPatternPeriod) {
        lastPatternToggle = now;
        ledState = !ledState;

        // Set LED based on pattern and state
        if (ledState) {
            switch (currentPattern) {
                case LED_PATTERN_IDENTIFY:
                    setRGB(0, 255, 0);  // Bright green
                    setBuzzer(true);
                    break;

                case LED_PATTERN_AP_MODE:
                    setRGB(0, 128, 0);  // Dimmed green (50%)
                    // No buzzer for AP mode
                    break;

                default:
                    break;
            }
        } else {
            // Off state for all patterns
            setRGB(0, 0, 0);
            setBuzzer(false);
        }
    }
}

void DeviceIdentification::setLEDPattern(LEDPattern pattern) {
    // Only change if pattern is different
    if (currentPattern == pattern) {
        return;
    }

    Serial.printf("LED Pattern changed: %d -> %d\n", currentPattern, pattern);

    currentPattern = pattern;
    lastPatternToggle = millis();
    ledState = false;

    // Set pattern period
    switch (pattern) {
        case LED_PATTERN_IDENTIFY:
            currentPatternPeriod = PATTERN_IDENTIFY_PERIOD;
            Serial.println("LED: Fast blink (identify mode)");
            break;

        case LED_PATTERN_AP_MODE:
            currentPatternPeriod = PATTERN_AP_MODE_PERIOD;
            Serial.println("LED: Slow blink (AP mode)");
            break;

        case LED_PATTERN_OFF:
        default:
            currentPatternPeriod = 0;
            setRGB(0, 0, 0);
            setBuzzer(false);
            Serial.println("LED: Off");
            break;
    }
}

void DeviceIdentification::flashIdentify(uint16_t durationSeconds) {
    Serial.printf("Flashing device for identification (%d seconds)...\n", durationSeconds);

    flashing = true;
    flashEndTime = millis() + (durationSeconds * 1000);

    while (millis() < flashEndTime && flashing) {
        // Green flash pattern + buzzer beep
        setRGB(0, 255, 0);  // Bright green
        setBuzzer(true);
        delay(200);

        setRGB(0, 0, 0);  // Off
        setBuzzer(false);
        delay(200);

        // Feed watchdog during long flash sequence
        yield();
    }

    // Turn off
    setRGB(0, 0, 0);
    setBuzzer(false);
    flashing = false;

    Serial.println("Flash identification complete");
}

bool DeviceIdentification::isFlashing() {
    return flashing && (millis() < flashEndTime);
}

void DeviceIdentification::stopFlashing() {
    flashing = false;
    setRGB(0, 0, 0);
    setBuzzer(false);
    Serial.println("Flash identification stopped");
}

void DeviceIdentification::setRGB(uint8_t red, uint8_t green, uint8_t blue) {
    // Note: GPIO38 might control a single RGB LED or separate R/G/B pins
    // Adjust based on actual hardware configuration
    // For now, using green channel only since we know it's a single LED
    ledcWrite(GPIO_RGB_LED, green);
}

void DeviceIdentification::setBuzzer(bool on) {
    if (on) {
        ledcWrite(GPIO_BUZZER, 128);  // 50% duty cycle
    } else {
        ledcWrite(GPIO_BUZZER, 0);    // Off
    }
}
