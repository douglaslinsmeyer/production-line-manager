#include "profile_sync_manager.h"

ProfileSyncManager::ProfileSyncManager(ProfileManager* profileMgr, ProfileStorage* storage)
    : profileManager(profileMgr),
      storage(storage),
      syncing(false),
      lastError("") {
}

void ProfileSyncManager::begin() {
    Serial.println("[ProfileSyncManager] Initialized");
}

bool ProfileSyncManager::downloadProfile(const String& profileId, const String& apiBaseUrl) {
    if (syncing) {
        Serial.println("[ProfileSyncManager] Sync already in progress");
        return false;
    }

    syncing = true;
    lastError = "";

    Serial.printf("[ProfileSyncManager] Downloading profile %s from %s\n",
                 profileId.c_str(), apiBaseUrl.c_str());

    HTTPClient http;
    String url = buildProfileUrl(profileId, apiBaseUrl);

    http.begin(url);
    http.setTimeout(HTTP_TIMEOUT_MS);

    int httpCode = http.GET();

    if (httpCode != HTTP_CODE_OK) {
        lastError = String("HTTP error: ") + String(httpCode);
        Serial.printf("[ProfileSyncManager] Error: %s\n", lastError.c_str());
        http.end();
        syncing = false;
        return false;
    }

    String payload = http.getString();
    http.end();

    Serial.printf("[ProfileSyncManager] Downloaded %d bytes\n", payload.length());

    // Apply the downloaded profile
    bool success = applyProfileUpdate(payload);
    syncing = false;

    return success;
}

bool ProfileSyncManager::applyProfileUpdate(const String& profileJson) {
    Serial.println("[ProfileSyncManager] Applying profile update...");

    // Parse JSON
    JsonDocument profileDoc;
    DeserializationError error = deserializeJson(profileDoc, profileJson);

    if (error) {
        lastError = String("JSON parse error: ") + error.c_str();
        Serial.printf("[ProfileSyncManager] Error: %s\n", lastError.c_str());
        return false;
    }

    // Validate profile structure
    if (!validateProfile(profileDoc)) {
        return false;
    }

    // Extract profile metadata
    String newProfileId = profileDoc["id"].as<String>();
    uint32_t newVersion = profileDoc["version"].as<uint32_t>();
    uint32_t oldVersion = storage->getProfileVersion();

    Serial.printf("[ProfileSyncManager] New profile: ID=%s, Version=%d (current: %d)\n",
                 newProfileId.c_str(), newVersion, oldVersion);

    // Handle state migration
    bool stateChanged = false;
    if (!handleStateMigration(profileDoc, stateChanged)) {
        return false;
    }

    // Save profile to storage
    if (!storage->saveProfile(profileJson, newProfileId, newVersion)) {
        lastError = "Failed to save profile to NVS";
        Serial.printf("[ProfileSyncManager] Error: %s\n", lastError.c_str());
        return false;
    }

    // Reload profile in manager
    if (!profileManager->loadProfile()) {
        lastError = "Failed to load updated profile";
        Serial.printf("[ProfileSyncManager] Error: %s\n", lastError.c_str());
        return false;
    }

    Serial.printf("[ProfileSyncManager] Profile updated successfully: v%d -> v%d\n",
                 oldVersion, newVersion);

    return true;
}

bool ProfileSyncManager::forceSync(const String& profileId, const String& apiBaseUrl) {
    Serial.println("[ProfileSyncManager] Force sync requested");
    return downloadProfile(profileId, apiBaseUrl);
}

bool ProfileSyncManager::confirmUpdate(const String& apiBaseUrl, const String& deviceMAC,
                                      uint32_t oldVersion, uint32_t newVersion, bool stateChanged) {
    Serial.printf("[ProfileSyncManager] Confirming update to backend: v%d -> v%d\n",
                 oldVersion, newVersion);

    HTTPClient http;
    String url = buildConfirmUrl(deviceMAC, apiBaseUrl);

    http.begin(url);
    http.setTimeout(10000);  // 10 second timeout for confirmation
    http.addHeader("Content-Type", "application/json");

    // Build confirmation payload
    JsonDocument doc;
    doc["profileId"] = storage->getProfileId();
    doc["newVersion"] = newVersion;
    doc["previousVersion"] = oldVersion;
    doc["currentState"] = storage->getCurrentState();
    doc["stateChanged"] = stateChanged;

    String payload;
    serializeJson(doc, payload);

    int httpCode = http.POST(payload);

    if (httpCode == HTTP_CODE_OK || httpCode == HTTP_CODE_CREATED) {
        Serial.println("[ProfileSyncManager] Update confirmation sent");
        http.end();
        return true;
    } else {
        Serial.printf("[ProfileSyncManager] Confirmation failed: HTTP %d\n", httpCode);
        http.end();
        return false;
    }
}

// ========== Private Helper Methods ==========

bool ProfileSyncManager::validateProfile(const JsonDocument& profileDoc) {
    // Check required fields
    if (profileDoc["id"].isNull()) {
        lastError = "Missing required field: id";
        Serial.printf("[ProfileSyncManager] Validation error: %s\n", lastError.c_str());
        return false;
    }

    if (profileDoc["version"].isNull()) {
        lastError = "Missing required field: version";
        Serial.printf("[ProfileSyncManager] Validation error: %s\n", lastError.c_str());
        return false;
    }

    if (profileDoc["states"].isNull() || !profileDoc["states"].is<JsonArray>()) {
        lastError = "Missing or invalid field: states";
        Serial.printf("[ProfileSyncManager] Validation error: %s\n", lastError.c_str());
        return false;
    }

    if (profileDoc["buttonBehavior"].isNull()) {
        lastError = "Missing required field: buttonBehavior";
        Serial.printf("[ProfileSyncManager] Validation error: %s\n", lastError.c_str());
        return false;
    }

    if (profileDoc["defaultState"].isNull()) {
        lastError = "Missing required field: defaultState";
        Serial.printf("[ProfileSyncManager] Validation error: %s\n", lastError.c_str());
        return false;
    }

    JsonArrayConst states = profileDoc["states"];
    if (states.size() == 0) {
        lastError = "Profile must have at least one state";
        Serial.printf("[ProfileSyncManager] Validation error: %s\n", lastError.c_str());
        return false;
    }

    Serial.println("[ProfileSyncManager] Profile validation passed");
    return true;
}

bool ProfileSyncManager::handleStateMigration(const JsonDocument& newProfile, bool& stateChanged) {
    String currentState = storage->getCurrentState();

    if (currentState.length() == 0) {
        // No current state, use default from new profile
        String defaultState = newProfile["defaultState"].as<String>();
        storage->setCurrentState(defaultState.c_str());
        storage->setOverrideFlag(false);
        stateChanged = true;

        Serial.printf("[ProfileSyncManager] No current state, using default: %s\n",
                     defaultState.c_str());
        return true;
    }

    // Check if current state exists in new profile
    JsonArrayConst states = newProfile["states"];
    bool stateExists = false;

    for (JsonObjectConst state : states) {
        String stateName = state["name"].as<String>();
        if (stateName.equals(currentState)) {
            stateExists = true;
            break;
        }
    }

    if (stateExists) {
        // Current state still exists, keep it
        Serial.printf("[ProfileSyncManager] Current state '%s' exists in new profile\n",
                     currentState.c_str());
        stateChanged = false;
    } else {
        // Current state removed, migrate to default
        String defaultState = newProfile["defaultState"].as<String>();

        Serial.printf("[ProfileSyncManager] State migration: '%s' -> '%s' (current state removed)\n",
                     currentState.c_str(), defaultState.c_str());

        storage->setCurrentState(defaultState.c_str());
        storage->setOverrideFlag(false);  // Clear override since state no longer valid
        stateChanged = true;
    }

    return true;
}

String ProfileSyncManager::buildProfileUrl(const String& profileId, const String& apiBaseUrl) {
    return apiBaseUrl + "/profiles/" + profileId;
}

String ProfileSyncManager::buildConfirmUrl(const String& deviceMAC, const String& apiBaseUrl) {
    return apiBaseUrl + "/devices/" + deviceMAC + "/profile-updated";
}
