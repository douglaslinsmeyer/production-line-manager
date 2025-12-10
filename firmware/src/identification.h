#pragma once

#include <Arduino.h>

// Device Identification via LED and Buzzer
class DeviceIdentification {
public:
    DeviceIdentification();

    // Initialize RGB LED and buzzer
    void begin();

    // Flash device for physical identification
    // Blinks green LED and beeps buzzer for specified duration
    void flashIdentify(uint16_t durationSeconds = 10);

    // Check if currently flashing
    bool isFlashing();

    // Stop flashing immediately
    void stopFlashing();

    // Set RGB LED color
    void setRGB(uint8_t red, uint8_t green, uint8_t blue);

    // Set buzzer state
    void setBuzzer(bool on);

private:
    bool flashing;
    unsigned long flashEndTime;

    // LED PWM channels
    static const uint8_t PWM_CHANNEL_R = 0;
    static const uint8_t PWM_CHANNEL_G = 1;
    static const uint8_t PWM_CHANNEL_B = 2;
    static const uint8_t PWM_CHANNEL_BUZZER = 3;

    static const uint16_t PWM_FREQ = 5000;
    static const uint8_t PWM_RESOLUTION = 8;
};
