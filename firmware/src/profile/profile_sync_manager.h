#pragma once

#include <Arduino.h>
#include <HTTPClient.h>
#include <ArduinoJson.h>
#include "profile_manager.h"
#include "profile_storage.h"

/**
 * Profile Sync Manager
 *
 * Handles synchronization of signal profiles with backend API:
 * - Downloads profiles via HTTP GET when update available
 * - Validates downloaded profile JSON
 * - Handles state migration when current state removed from profile
 * - Confirms successful updates to backend
 *
 * Triggered by:
 * - Heartbeat response indicating update available
 * - MQTT command "update_profile"
 * - Manual force sync
 */
class ProfileSyncManager {
public:
    ProfileSyncManager(ProfileManager* profileMgr,
                      ProfileStorage* storage);

    /**
     * Initialize sync manager
     */
    void begin();

    /**
     * Download profile from backend API
     * @param profileId Profile UUID to download
     * @param apiBaseUrl Base URL for API (e.g., "http://192.168.1.100:8080/api/v1")
     * @return true if downloaded and applied successfully
     */
    bool downloadProfile(const String& profileId, const String& apiBaseUrl);

    /**
     * Handle profile update from heartbeat response
     * Backend includes full profile JSON in heartbeat when update available
     * @param profileJson Profile JSON from heartbeat response
     * @return true if profile applied successfully
     */
    bool applyProfileUpdate(const String& profileJson);

    /**
     * Force sync: Download profile from backend regardless of version
     * @param profileId Profile UUID
     * @param apiBaseUrl Base URL for API
     * @return true if successful
     */
    bool forceSync(const String& profileId, const String& apiBaseUrl);

    /**
     * Confirm profile update to backend
     * Sends POST request to backend confirming successful update
     * @param apiBaseUrl Base URL for API
     * @param deviceMAC Device MAC address
     * @param oldVersion Previous profile version
     * @param newVersion New profile version
     * @param stateChanged true if state was migrated
     * @return true if confirmation sent successfully
     */
    bool confirmUpdate(const String& apiBaseUrl, const String& deviceMAC,
                      uint32_t oldVersion, uint32_t newVersion, bool stateChanged);

    /**
     * Get last sync error message
     */
    String getLastError() const { return lastError; }

    /**
     * Check if a sync is currently in progress
     */
    bool isSyncing() const { return syncing; }

private:
    ProfileManager* profileManager;
    ProfileStorage* storage;

    bool syncing;
    String lastError;

    static const uint32_t HTTP_TIMEOUT_MS = 30000;  // 30 second timeout

    // Helper methods
    bool validateProfile(const JsonDocument& profileDoc);
    bool handleStateMigration(const JsonDocument& newProfile, bool& stateChanged);
    String buildProfileUrl(const String& profileId, const String& apiBaseUrl);
    String buildConfirmUrl(const String& deviceMAC, const String& apiBaseUrl);
};
