#include "identification.h"

DeviceIdentification::DeviceIdentification() : flashing(false), flashEndTime(0) {
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
