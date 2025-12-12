#include "digital_output.h"

// TCA9554PWR Register definitions
#define TCA9554_INPUT_REG    0x00
#define TCA9554_OUTPUT_REG   0x01
#define TCA9554_POLARITY_REG 0x02
#define TCA9554_CONFIG_REG   0x03

DigitalOutputManager::DigitalOutputManager() : outputState(0xFF) {
    // Initialize to 0xFF (all outputs OFF due to inverted logic)
}

bool DigitalOutputManager::begin() {
    // Initialize I2C on GPIO41 (SCL) and GPIO42 (SDA)
    // Note: These are JTAG pins (MTDI/MTMS) - hardware JTAG will not be available
    Wire.begin(I2C_SDA_PIN, I2C_SCL_PIN);
    delay(10);

    // Test I2C communication
    Wire.beginTransmission(TCA9554_ADDRESS);
    uint8_t error = Wire.endTransmission();

    if (error != 0) {
        Serial.printf("TCA9554PWR not found at address 0x%02X (I2C error: %d)\n",
                     TCA9554_ADDRESS, error);
        return false;
    }

    // Configure all pins as outputs (0 = output, 1 = input)
    if (!writeRegister(TCA9554_CONFIG_REG, 0x00)) {
        Serial.println("Failed to configure TCA9554PWR pins as outputs");
        return false;
    }

    // Set all outputs OFF initially (0xFF due to inverted logic)
    // Note: Darlington sinking transistors require HIGH=OFF, LOW=ON
    if (!writeRegister(TCA9554_OUTPUT_REG, 0xFF)) {
        Serial.println("Failed to set initial output state");
        return false;
    }

    outputState = 0xFF;
    Serial.println("TCA9554PWR initialized - all outputs OFF");
    return true;
}

bool DigitalOutputManager::setOutput(uint8_t channel, bool state) {
    if (channel >= 8) {
        Serial.printf("Invalid channel %d (must be 0-7)\n", channel);
        return false;
    }

    // IMPORTANT: TCA9554PWR with Darlington sinking transistors uses INVERTED logic:
    // - state=true (want LED ON) → write 0 (LOW) → transistor conducts → LED ON
    // - state=false (want LED OFF) → write 1 (HIGH) → transistor off → LED OFF
    bool invertedState = !state;

    if (invertedState) {
        outputState |= (1 << channel);
    } else {
        outputState &= ~(1 << channel);
    }

    return writeRegister(TCA9554_OUTPUT_REG, outputState);
}

bool DigitalOutputManager::setAllOutputs(uint8_t state) {
    outputState = state;
    return writeRegister(TCA9554_OUTPUT_REG, outputState);
}

bool DigitalOutputManager::toggleOutput(uint8_t channel) {
    if (channel >= 8) {
        Serial.printf("Invalid channel %d (must be 0-7)\n", channel);
        return false;
    }

    outputState ^= (1 << channel);
    return writeRegister(TCA9554_OUTPUT_REG, outputState);
}

uint8_t DigitalOutputManager::getAllOutputs() {
    return outputState;
}

bool DigitalOutputManager::getOutput(uint8_t channel) {
    if (channel >= 8) return false;
    return (outputState & (1 << channel)) != 0;
}

bool DigitalOutputManager::writeRegister(uint8_t reg, uint8_t data) {
    Wire.beginTransmission(TCA9554_ADDRESS);
    Wire.write(reg);
    Wire.write(data);
    uint8_t error = Wire.endTransmission();

    if (error != 0) {
        Serial.printf("I2C write error: %d (reg=0x%02X, data=0x%02X)\n", error, reg, data);
        return false;
    }

    return true;
}

uint8_t DigitalOutputManager::readRegister(uint8_t reg) {
    Wire.beginTransmission(TCA9554_ADDRESS);
    Wire.write(reg);
    Wire.endTransmission();

    Wire.requestFrom((int)TCA9554_ADDRESS, (int)1);
    if (Wire.available()) {
        return Wire.read();
    }

    return 0xFF;  // Error value
}
