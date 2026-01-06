#pragma once

#include <Arduino.h>
#include <Preferences.h>

/**
 * Signal Profile Storage Manager
 *
 * Manages persistent storage of signal profiles in ESP32 NVS (Non-Volatile Storage).
 * Stores profile JSON, current state, override flag, and sync metadata.
 *
 * NVS Namespace: "profile"
 *
 * Storage Keys:
 * - profile_id: UUID of assigned profile (String, max 40 chars)
 * - profile_ver: Version number (uint32_t)
 * - profile_json: Full profile JSON (String, max 4096 bytes)
 * - current_state: Current state name (String, max 32 chars)
 * - is_override: Override flag (bool)
 * - last_sync: Last sync timestamp in millis (uint64_t)
 */
class ProfileStorage {
public:
    ProfileStorage();

    /**
     * Initialize storage (must be called before use)
     */
    void begin();

    /**
     * Save complete profile JSON to NVS
     * @param profileJson JSON string (max 4096 bytes)
     * @param profileId Profile UUID
     * @param version Profile version number
     * @return true if saved successfully
     */
    bool saveProfile(const String& profileJson, const String& profileId, uint32_t version);

    /**
     * Load profile JSON from NVS
     * @param profileJson Output parameter for JSON string
     * @return true if profile exists and was loaded
     */
    bool loadProfile(String& profileJson);

    /**
     * Get stored profile ID
     * @return Profile UUID or empty string if not set
     */
    String getProfileId();

    /**
     * Get stored profile version
     * @return Version number or 0 if not set
     */
    uint32_t getProfileVersion();

    /**
     * Set current active state name
     * @param stateName State name (max 32 chars)
     * @return true if saved successfully
     */
    bool setCurrentState(const char* stateName);

    /**
     * Get current active state name
     * @return State name or empty string if not set
     */
    String getCurrentState();

    /**
     * Set override flag (indicates manual state change)
     * @param isOverridden Override flag
     * @return true if saved successfully
     */
    bool setOverrideFlag(bool isOverridden);

    /**
     * Get override flag
     * @return true if device state is overridden
     */
    bool getOverrideFlag();

    /**
     * Update last sync timestamp
     * @param timestamp Timestamp in milliseconds
     * @return true if saved successfully
     */
    bool setLastSync(uint64_t timestamp);

    /**
     * Get last sync timestamp
     * @return Timestamp in milliseconds or 0 if not set
     */
    uint64_t getLastSync();

    /**
     * Check if a profile is stored
     * @return true if profile JSON exists in NVS
     */
    bool hasProfile();

    /**
     * Clear all profile data from NVS
     * Useful for factory reset or profile unassignment
     */
    void clearProfile();

    /**
     * Get NVS usage statistics
     * @param totalBytes Output: total NVS space used
     * @param freeBytes Output: free NVS space
     */
    void getStorageStats(size_t& totalBytes, size_t& freeBytes);

private:
    Preferences prefs;
    static const char* NAMESPACE;

    // NVS Keys
    static const char* KEY_PROFILE_ID;
    static const char* KEY_PROFILE_VERSION;
    static const char* KEY_PROFILE_JSON;
    static const char* KEY_CURRENT_STATE;
    static const char* KEY_IS_OVERRIDE;
    static const char* KEY_LAST_SYNC;

    // Size limits
    static const size_t MAX_PROFILE_JSON_SIZE = 4096;
    static const size_t MAX_STATE_NAME_LENGTH = 32;
    static const size_t MAX_PROFILE_ID_LENGTH = 40;
};
