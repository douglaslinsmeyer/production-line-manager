#pragma once

#include <Arduino.h>
#include <Wire.h>

// TCA9554PWR I2C GPIO Expander for Digital Outputs
class DigitalOutputManager {
public:
    DigitalOutputManager();

    // Initialize the TCA9554PWR GPIO expander
    bool begin();

    // Control individual output (channel 0-7)
    bool setOutput(uint8_t channel, bool state);

    // Control all outputs at once (bitmask 0x00-0xFF)
    bool setAllOutputs(uint8_t state);

    // Toggle individual output
    bool toggleOutput(uint8_t channel);

    // Get current state of all outputs
    uint8_t getAllOutputs();

    // Get individual output state
    bool getOutput(uint8_t channel);

private:
    uint8_t outputState;  // Current state of all outputs

    // I2C register operations
    bool writeRegister(uint8_t reg, uint8_t data);
    uint8_t readRegister(uint8_t reg);
};
