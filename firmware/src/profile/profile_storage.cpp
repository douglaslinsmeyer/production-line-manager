#include "profile_storage.h"

// Static constants
const char* ProfileStorage::NAMESPACE = "profile";

const char* ProfileStorage::KEY_PROFILE_ID = "profile_id";
const char* ProfileStorage::KEY_PROFILE_VERSION = "profile_ver";
const char* ProfileStorage::KEY_PROFILE_JSON = "profile_json";
const char* ProfileStorage::KEY_CURRENT_STATE = "current_state";
const char* ProfileStorage::KEY_IS_OVERRIDE = "is_override";
const char* ProfileStorage::KEY_LAST_SYNC = "last_sync";

ProfileStorage::ProfileStorage() {
}

void ProfileStorage::begin() {
    // Preferences will be opened/closed on each operation
    // This avoids keeping the namespace locked
}

bool ProfileStorage::saveProfile(const String& profileJson, const String& profileId, uint32_t version) {
    // Validate input sizes
    if (profileJson.length() > MAX_PROFILE_JSON_SIZE) {
        Serial.printf("[ProfileStorage] Error: Profile JSON too large (%d bytes, max %d)\n",
                     profileJson.length(), MAX_PROFILE_JSON_SIZE);
        return false;
    }

    if (profileId.length() > MAX_PROFILE_ID_LENGTH) {
        Serial.printf("[ProfileStorage] Error: Profile ID too long (%d chars, max %d)\n",
                     profileId.length(), MAX_PROFILE_ID_LENGTH);
        return false;
    }

    // Open preferences in read-write mode
    if (!prefs.begin(NAMESPACE, false)) {
        Serial.println("[ProfileStorage] Error: Failed to open NVS namespace for writing");
        return false;
    }

    // Save all profile data
    bool success = true;

    if (!prefs.putString(KEY_PROFILE_JSON, profileJson)) {
        Serial.println("[ProfileStorage] Error: Failed to save profile JSON");
        success = false;
    }

    if (success && !prefs.putString(KEY_PROFILE_ID, profileId)) {
        Serial.println("[ProfileStorage] Error: Failed to save profile ID");
        success = false;
    }

    if (success && !prefs.putUInt(KEY_PROFILE_VERSION, version)) {
        Serial.println("[ProfileStorage] Error: Failed to save profile version");
        success = false;
    }

    // Update last sync timestamp
    if (success) {
        prefs.putULong64(KEY_LAST_SYNC, millis());
    }

    prefs.end();

    if (success) {
        Serial.printf("[ProfileStorage] Profile saved: ID=%s, Version=%d, Size=%d bytes\n",
                     profileId.c_str(), version, profileJson.length());
    }

    return success;
}

bool ProfileStorage::loadProfile(String& profileJson) {
    if (!prefs.begin(NAMESPACE, true)) { // Read-only
        Serial.println("[ProfileStorage] Error: Failed to open NVS namespace for reading");
        return false;
    }

    profileJson = prefs.getString(KEY_PROFILE_JSON, "");
    prefs.end();

    if (profileJson.length() == 0) {
        Serial.println("[ProfileStorage] No profile found in NVS");
        return false;
    }

    Serial.printf("[ProfileStorage] Profile loaded: Size=%d bytes\n", profileJson.length());
    return true;
}

String ProfileStorage::getProfileId() {
    if (!prefs.begin(NAMESPACE, true)) {
        return "";
    }

    String id = prefs.getString(KEY_PROFILE_ID, "");
    prefs.end();
    return id;
}

uint32_t ProfileStorage::getProfileVersion() {
    if (!prefs.begin(NAMESPACE, true)) {
        return 0;
    }

    uint32_t version = prefs.getUInt(KEY_PROFILE_VERSION, 0);
    prefs.end();
    return version;
}

bool ProfileStorage::setCurrentState(const char* stateName) {
    if (strlen(stateName) > MAX_STATE_NAME_LENGTH) {
        Serial.printf("[ProfileStorage] Error: State name too long (%d chars, max %d)\n",
                     strlen(stateName), MAX_STATE_NAME_LENGTH);
        return false;
    }

    if (!prefs.begin(NAMESPACE, false)) {
        Serial.println("[ProfileStorage] Error: Failed to open NVS namespace");
        return false;
    }

    bool success = prefs.putString(KEY_CURRENT_STATE, stateName);
    prefs.end();

    if (success) {
        Serial.printf("[ProfileStorage] Current state saved: %s\n", stateName);
    } else {
        Serial.println("[ProfileStorage] Error: Failed to save current state");
    }

    return success;
}

String ProfileStorage::getCurrentState() {
    if (!prefs.begin(NAMESPACE, true)) {
        return "";
    }

    String state = prefs.getString(KEY_CURRENT_STATE, "");
    prefs.end();
    return state;
}

bool ProfileStorage::setOverrideFlag(bool isOverridden) {
    if (!prefs.begin(NAMESPACE, false)) {
        Serial.println("[ProfileStorage] Error: Failed to open NVS namespace");
        return false;
    }

    bool success = prefs.putBool(KEY_IS_OVERRIDE, isOverridden);
    prefs.end();

    if (success) {
        Serial.printf("[ProfileStorage] Override flag saved: %s\n", isOverridden ? "true" : "false");
    } else {
        Serial.println("[ProfileStorage] Error: Failed to save override flag");
    }

    return success;
}

bool ProfileStorage::getOverrideFlag() {
    if (!prefs.begin(NAMESPACE, true)) {
        return false;
    }

    bool isOverridden = prefs.getBool(KEY_IS_OVERRIDE, false);
    prefs.end();
    return isOverridden;
}

bool ProfileStorage::setLastSync(uint64_t timestamp) {
    if (!prefs.begin(NAMESPACE, false)) {
        return false;
    }

    bool success = prefs.putULong64(KEY_LAST_SYNC, timestamp);
    prefs.end();
    return success;
}

uint64_t ProfileStorage::getLastSync() {
    if (!prefs.begin(NAMESPACE, true)) {
        return 0;
    }

    uint64_t timestamp = prefs.getULong64(KEY_LAST_SYNC, 0);
    prefs.end();
    return timestamp;
}

bool ProfileStorage::hasProfile() {
    if (!prefs.begin(NAMESPACE, true)) {
        return false;
    }

    bool exists = prefs.isKey(KEY_PROFILE_JSON);
    prefs.end();
    return exists;
}

void ProfileStorage::clearProfile() {
    if (!prefs.begin(NAMESPACE, false)) {
        Serial.println("[ProfileStorage] Error: Failed to open NVS namespace for clearing");
        return;
    }

    // Clear all keys
    prefs.remove(KEY_PROFILE_JSON);
    prefs.remove(KEY_PROFILE_ID);
    prefs.remove(KEY_PROFILE_VERSION);
    prefs.remove(KEY_CURRENT_STATE);
    prefs.remove(KEY_IS_OVERRIDE);
    prefs.remove(KEY_LAST_SYNC);

    prefs.end();

    Serial.println("[ProfileStorage] Profile data cleared from NVS");
}

void ProfileStorage::getStorageStats(size_t& totalBytes, size_t& freeBytes) {
    if (!prefs.begin(NAMESPACE, true)) {
        totalBytes = 0;
        freeBytes = 0;
        return;
    }

    totalBytes = prefs.getBytesLength(KEY_PROFILE_JSON);
    totalBytes += prefs.getBytesLength(KEY_PROFILE_ID);
    totalBytes += sizeof(uint32_t); // profile_ver
    totalBytes += prefs.getBytesLength(KEY_CURRENT_STATE);
    totalBytes += sizeof(bool); // is_override
    totalBytes += sizeof(uint64_t); // last_sync

    // NVS partition size is typically 20KB, but this is approximate
    freeBytes = 20480 - totalBytes; // Rough estimate

    prefs.end();
}
