#pragma once

#include <Arduino.h>
#include <ArduinoJson.h>
#include "profile_storage.h"

/**
 * Signal Profile Manager
 *
 * Manages signal profile configuration including:
 * - Parsing profile JSON
 * - Validating state names
 * - Looking up output configurations
 * - Managing button cycle arrays
 * - Providing profile metadata
 *
 * Uses ArduinoJson for JSON parsing with a 2KB document size.
 */
class ProfileManager {
public:
    // Light modes (matches backend enum)
    enum LightMode {
        LIGHT_OFF = 0,
        LIGHT_ON = 1,
        LIGHT_SHORT_BLINK = 2,  // 500ms on/off
        LIGHT_LONG_BLINK = 3    // 1500ms on/off
    };

    // Buzzer modes (matches backend enum)
    enum BuzzerMode {
        BUZZER_OFF = 0,
        BUZZER_ON = 1,
        BUZZER_CHIRP = 2  // beep-pause-beep-pause-beep, 3s delay
    };

    // Output configuration for a state
    struct StateOutputs {
        LightMode redLight;
        LightMode yellowLight;
        LightMode greenLight;
        BuzzerMode primaryBuzzer;  // GPIO46 (was "buzzer")
        BuzzerMode towerBuzzer;    // DO4 (new)
    };

    ProfileManager(ProfileStorage* storage);

    /**
     * Load profile from storage and parse JSON
     * @return true if profile loaded and parsed successfully
     */
    bool loadProfile();

    /**
     * Update profile with new JSON
     * @param profileJson New profile JSON string
     * @param profileId Profile UUID
     * @param version Profile version number
     * @return true if saved and parsed successfully
     */
    bool updateProfile(const String& profileJson, const String& profileId, uint32_t version);

    /**
     * Check if a profile is loaded and ready
     * @return true if profile is loaded
     */
    bool hasProfile();

    /**
     * Check if a state name is valid (exists in profile)
     * @param stateName State name to check
     * @return true if state exists
     */
    bool isValidState(const char* stateName);

    /**
     * Get default state name from profile
     * @return Default state name or empty string if no profile
     */
    const char* getDefaultState();

    /**
     * Get output configuration for a specific state
     * @param stateName State name
     * @param outputs Output parameter for state outputs
     * @return true if state found and outputs populated
     */
    bool getStateOutputs(const char* stateName, StateOutputs& outputs);

    /**
     * Get next state in button press cycle
     * @param currentState Current state name
     * @param longPress true for long press cycle, false for short press
     * @return Next state name or current state if cycle empty
     */
    const char* getNextStateInCycle(const char* currentState, bool longPress);

    /**
     * Get profile ID
     * @return Profile UUID or empty string
     */
    String getProfileId();

    /**
     * Get profile version
     * @return Version number or 0 if no profile
     */
    uint32_t getProfileVersion();

    /**
     * Get profile name
     * @return Profile name or empty string
     */
    String getProfileName();

    /**
     * Get number of states in profile
     * @return State count
     */
    int getStateCount();

    /**
     * Clear loaded profile from memory
     */
    void clearProfile();

    /**
     * Get human-readable light mode name
     */
    static const char* lightModeToString(LightMode mode);

    /**
     * Get human-readable buzzer mode name
     */
    static const char* buzzerModeToString(BuzzerMode mode);

private:
    ProfileStorage* storage;
    JsonDocument profileDoc;
    bool profileLoaded;

    String profileId;
    uint32_t profileVersion;
    String profileName;
    String defaultState;

    // Helper methods
    LightMode parseLightMode(const char* mode);
    BuzzerMode parseBuzzerMode(const char* mode);
    bool parseProfile();
    int findStateIndex(const char* stateName);
    int findCurrentIndexInCycle(const char* currentState, JsonArrayConst cycle);
};
