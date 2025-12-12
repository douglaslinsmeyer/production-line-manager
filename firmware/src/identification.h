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

    // LED Pattern Management
    enum LEDPattern {
        LED_PATTERN_OFF,       // LED completely off
        LED_PATTERN_IDENTIFY,  // Fast blink for identification (200ms)
        LED_PATTERN_AP_MODE    // Slow blink for AP mode (500ms)
    };

    // Update LED state (call in main loop for non-blocking patterns)
    void update();

    // Set current LED pattern
    void setLEDPattern(LEDPattern pattern);

    // Get current pattern
    LEDPattern getCurrentPattern() const { return currentPattern; }

private:
    bool flashing;
    unsigned long flashEndTime;

    // Pattern state machine
    LEDPattern currentPattern;
    bool ledState;                  // Current on/off state
    unsigned long lastPatternToggle;// Last time pattern toggled
    uint16_t currentPatternPeriod;  // Current pattern half-period (ms)

    // Pattern timing constants
    static const uint16_t PATTERN_IDENTIFY_PERIOD = 200;  // 200ms on/off
    static const uint16_t PATTERN_AP_MODE_PERIOD = 500;   // 500ms on/off

    // LED PWM channels
    static const uint8_t PWM_CHANNEL_R = 0;
    static const uint8_t PWM_CHANNEL_G = 1;
    static const uint8_t PWM_CHANNEL_B = 2;
    static const uint8_t PWM_CHANNEL_BUZZER = 3;

    static const uint16_t PWM_FREQ = 5000;
    static const uint8_t PWM_RESOLUTION = 8;
};
