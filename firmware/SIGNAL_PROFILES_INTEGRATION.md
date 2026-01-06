# Signal Profiles Firmware Integration Guide

This guide explains how to integrate the Signal Profiles feature into the existing ESP32 firmware.

## Overview

The Signal Profiles feature adds configurable state-based control of outputs. Instead of hardcoded ON/OFF/MAINTENANCE/ERROR states, the device downloads a profile from the backend that defines:
- Available states and their names
- Output configuration for each state (lights and buzzer patterns)
- Button press cycles (short press and long press)

## New Components Created

### 1. Profile Storage (`src/profile/profile_storage.h/cpp`)
- NVS-based persistence for profile JSON, metadata, and current state
- Stores: profile_id, profile_version, profile_json, current_state, is_override, last_sync

### 2. Profile Manager (`src/profile/profile_manager.h/cpp`)
- Parses profile JSON with ArduinoJson
- Provides query interface for states, outputs, and button cycles
- Validates state names and lookups

### 3. Output Controller (`src/profile/output_controller.h/cpp`)
- Coordinates tower lights and buzzer based on profile configuration
- Applies state outputs by looking up configuration from ProfileManager

### 4. Tower Light Manager Refactored (`src/gpio/tower_light_refactored.h/cpp`)
- Supports blink patterns: OFF, ON, SHORT_BLINK (500ms), LONG_BLINK (1500ms)
- Non-blocking blink timers for each light
- update() method must be called in main loop

### 5. Buzzer Controller (`src/buzzer/buzzer_controller.h/cpp`)
- Supports modes: OFF, ON, CHIRP
- Chirp pattern: beep-pause-beep-pause-beep, 3s delay, repeat
- Non-blocking state machine
- update() method must be called in main loop

### 6. Line State Manager Refactored (`src/state/line_state_refactored.h/cpp`)
- Changed from enum-based to String-based state names
- Integrates with ProfileManager for button press cycles
- Backward compatible with legacy enum

### 7. Profile Sync Manager (`src/profile/profile_sync_manager.h/cpp`)
- Downloads profiles via HTTP from backend API
- Handles state migration when profile updated
- Confirms successful updates to backend

## Integration Steps

### Step 1: Update PlatformIO Build Configuration

Add ArduinoJson dependency to `platformio.ini` (if not already present):

```ini
lib_deps =
    ...existing dependencies...
    bblanchon/ArduinoJson@^7.0.0
```

### Step 2: Replace Existing Components

**Option A: Rename and Replace (Recommended for gradual migration)**
```bash
# Backup originals
mv src/state/line_state.h src/state/line_state_old.h
mv src/state/line_state.cpp src/state/line_state_old.cpp
mv src/gpio/tower_light.h src/gpio/tower_light_old.h
mv src/gpio/tower_light.cpp src/gpio/tower_light_old.cpp

# Use refactored versions
mv src/state/line_state_refactored.h src/state/line_state.h
mv src/state/line_state_refactored.cpp src/state/line_state.cpp
mv src/gpio/tower_light_refactored.h src/gpio/tower_light.h
mv src/gpio/tower_light_refactored.cpp src/gpio/tower_light.cpp
```

**Option B: Incremental Migration (Keep both versions)**
- Use refactored versions with different names
- Conditionally compile based on feature flag
- Update references gradually

### Step 3: Update main.cpp

Replace the existing initialization and loop code with the following:

#### Global Variables (add these)
```cpp
// Signal Profile components
ProfileStorage profileStorage;
ProfileManager profileManager(&profileStorage);
OutputController* outputController = nullptr;  // Initialized after other components
ProfileSyncManager profileSyncManager(&profileManager, &profileStorage);
BuzzerController buzzerController(BUZZER_PIN);  // GPIO46
```

#### setup() Function Changes

**After digital outputs are initialized:**
```cpp
// Initialize profile storage
profileStorage.begin();

// Load profile from NVS
if (!profileManager.loadProfile()) {
    Serial.println("No profile loaded - device will use default behavior");
    // TODO: Create default hardcoded profile or request from backend
}
```

**After tower lights are initialized:**
```cpp
// Initialize buzzer controller
buzzerController.begin();

// Initialize output controller
outputController = new OutputController(&towerLight, &buzzerController, &profileManager);
outputController->begin();

// Initialize profile sync manager
profileSyncManager.begin();
```

**After line state manager is initialized:**
```cpp
// Initialize line state with profile manager
lineState.begin(&profileManager);

// If profile loaded, apply initial state outputs
if (profileManager.hasProfile()) {
    String currentState = profileStorage.getCurrentState();
    if (currentState.length() == 0) {
        // No saved state, use default
        currentState = profileManager.getDefaultState();
        profileStorage.setCurrentState(currentState.c_str());
    }

    // Apply outputs for current state
    outputController->applyStateOutputs(currentState.c_str());
}
```

#### loop() Function Changes

**Add to main loop (after existing updates):**
```cpp
// Update output patterns (blink/chirp)
if (outputController != nullptr) {
    outputController->update();
}
```

### Step 4: Update Button Press Callbacks

**Replace existing callbacks:**

```cpp
void onControlButtonShortPress() {
    Serial.println("=== SHORT PRESS ===");

    // Use profile-based state transition
    const char* newState = lineState.handleShortPress();

    // Apply outputs for new state
    if (outputController != nullptr) {
        outputController->applyStateOutputs(newState);
    }

    // Set override flag (manual state change)
    profileStorage.setOverrideFlag(true);

    // Publish status immediately
    if (mqtt.isConnected()) {
        mqtt.publishStatus(
            digitalInputs.getInputStates(),
            digitalOutputs.getOutputStates(),
            networkManager.isConnected(),
            lineState.getStateLegacy()  // For backward compatibility
        );
    }
}

void onControlButtonLongPress() {
    Serial.println("=== LONG PRESS ===");

    // Use profile-based state transition
    const char* newState = lineState.handleLongPress();

    // Apply outputs for new state
    if (outputController != nullptr) {
        outputController->applyStateOutputs(newState);
    }

    // Set override flag
    profileStorage.setOverrideFlag(true);

    // Publish status
    if (mqtt.isConnected()) {
        mqtt.publishStatus(
            digitalInputs.getInputStates(),
            digitalOutputs.getOutputStates(),
            networkManager.isConnected(),
            lineState.getStateLegacy()
        );
    }
}
```

### Step 5: Extend MQTT Client

#### Add to mqtt_client.cpp in `publishStatus()` method:

**After existing fields, add profile sync metadata:**
```cpp
// Add profile sync metadata (if profile loaded)
if (profileManager.hasProfile()) {
    doc["profile_id"] = profileStorage.getProfileId();
    doc["profile_version"] = profileStorage.getProfileVersion();
    doc["current_state"] = profileStorage.getCurrentState();
    doc["is_overridden"] = profileStorage.getOverrideFlag();
}
```

#### Add to mqtt_client.cpp in `handleCommand()` method:

**Add new command handlers before existing commands:**
```cpp
// Handle set_state command (from signal profiles API)
if (strcmp(command, "set_state") == 0) {
    const char* stateName = doc["state"] | "";

    Serial.printf("Set state command: %s\n", stateName);

    if (lineState.setState(stateName, "mqtt")) {
        // Apply outputs for new state
        if (outputController != nullptr) {
            outputController->applyStateOutputs(stateName);
        }

        // Backend control = no override
        profileStorage.setOverrideFlag(false);
        profileStorage.setCurrentState(stateName);
    } else {
        Serial.printf("Invalid state: %s\n", stateName);
    }
    return;
}

// Handle update_profile command
if (strcmp(command, "update_profile") == 0) {
    Serial.println("Profile update command received");

    // Profile JSON included in command or trigger download
    if (doc.containsKey("profile")) {
        JsonDocument profileDoc;
        String profileJson;
        serializeJson(doc["profile"], profileJson);

        if (profileSyncManager.applyProfileUpdate(profileJson)) {
            Serial.println("Profile updated from MQTT command");

            // Apply outputs for current state with new configuration
            if (outputController != nullptr) {
                outputController->applyStateOutputs(lineState.getState());
            }
        }
    }
    return;
}

// Handle clear_override command
if (strcmp(command, "clear_override") == 0) {
    Serial.println("Clear override command");

    profileStorage.setOverrideFlag(false);

    // Return to default state
    if (profileManager.hasProfile()) {
        const char* defaultState = profileManager.getDefaultState();
        lineState.setState(defaultState, "override_clear");

        if (outputController != nullptr) {
            outputController->applyStateOutputs(defaultState);
        }
    }
    return;
}
```

### Step 6: Update Device Web Server

Add these endpoints to `device_webserver.cpp`:

#### GET /profile - View Current Profile
```cpp
webServer->on("/profile", HTTP_GET, [this]() {
    if (!profileManager.hasProfile()) {
        webServer->send(404, "application/json", "{\"error\":\"No profile loaded\"}");
        return;
    }

    String profileJson;
    if (profileStorage.loadProfile(profileJson)) {
        webServer->send(200, "application/json", profileJson);
    } else {
        webServer->send(500, "application/json", "{\"error\":\"Failed to load profile\"}");
    }
});
```

#### GET /state - View Current State
```cpp
webServer->on("/state", HTTP_GET, [this]() {
    JsonDocument doc;
    doc["current_state"] = profileStorage.getCurrentState();
    doc["is_overridden"] = profileStorage.getOverrideFlag();
    doc["profile_id"] = profileStorage.getProfileId();
    doc["profile_version"] = profileStorage.getProfileVersion();

    // List available states
    if (profileManager.hasProfile()) {
        JsonArray states = doc["available_states"].to<JsonArray>();
        // Parse profile to get state names
        String profileJson;
        profileStorage.loadProfile(profileJson);
        JsonDocument profileDoc;
        deserializeJson(profileDoc, profileJson);

        for (JsonObject state : profileDoc["states"].as<JsonArray>()) {
            states.add(state["name"].as<String>());
        }
    }

    String output;
    serializeJson(doc, output);
    webServer->send(200, "application/json", output);
});
```

#### POST /state/set - Set State Manually
```cpp
webServer->on("/state/set", HTTP_POST, [this]() {
    if (!webServer->hasArg("state")) {
        webServer->send(400, "application/json", "{\"error\":\"Missing 'state' parameter\"}");
        return;
    }

    String stateName = webServer->arg("state");

    if (lineState.setState(stateName.c_str(), "webui")) {
        profileStorage.setCurrentState(stateName.c_str());
        profileStorage.setOverrideFlag(true);

        // Apply outputs
        if (outputController != nullptr) {
            outputController->applyStateOutputs(stateName.c_str());
        }

        webServer->send(200, "application/json",
                       "{\"success\":true,\"state\":\"" + stateName + "\"}");
    } else {
        webServer->send(400, "application/json",
                       "{\"success\":false,\"error\":\"Invalid state\"}");
    }
});
```

#### POST /test/outputs - Test Outputs
```cpp
webServer->on("/test/outputs", HTTP_POST, [this]() {
    // Parse JSON body
    JsonDocument doc;
    deserializeJson(doc, webServer->arg("plain"));

    bool red = doc["red"] | false;
    bool yellow = doc["yellow"] | false;
    bool green = doc["green"] | false;
    bool buzzer = doc["buzzer"] | false;

    if (outputController != nullptr) {
        outputController->testOutputs(red, yellow, green, buzzer);
    }

    webServer->send(200, "application/json", "{\"success\":true}");
});
```

#### POST /override/clear - Clear Override
```cpp
webServer->on("/override/clear", HTTP_POST, [this]() {
    profileStorage.setOverrideFlag(false);

    // Return to default state
    if (profileManager.hasProfile()) {
        const char* defaultState = profileManager.getDefaultState();
        lineState.setState(defaultState, "webui_clear");

        profileStorage.setCurrentState(defaultState);

        if (outputController != nullptr) {
            outputController->applyStateOutputs(defaultState);
        }
    }

    webServer->send(200, "application/json", "{\"success\":true}");
});
```

## Compilation Notes

### Include Paths
Make sure these directories are in your include path:
- `src/profile/`
- `src/buzzer/`
- `src/gpio/`
- `src/state/`

### Build Order
Components have dependencies:
1. ProfileStorage (no dependencies)
2. ProfileManager (depends on ProfileStorage)
3. LineStateManager (depends on ProfileManager)
4. TowerLightManager (depends on ProfileManager, DigitalOutputManager)
5. BuzzerController (depends on ProfileManager)
6. OutputController (depends on all output components + ProfileManager)
7. ProfileSyncManager (depends on ProfileManager, ProfileStorage)

### Memory Considerations

**Heap Usage:**
- JsonDocument (2KB): ~2500 bytes
- String buffers: ~500 bytes
- ProfileManager: ~200 bytes
- Total added: ~3.2KB heap

**NVS Usage:**
- Profile JSON: up to 4096 bytes
- Metadata: ~500 bytes
- Total: ~4.6KB of 20KB NVS partition

**Should not cause issues on ESP32-S3 with 2MB PSRAM and 200KB+ heap.**

## Testing Checklist

### Local Testing (Without Backend)
1. ✅ Create default hardcoded profile in code
2. ✅ Test ProfileStorage save/load
3. ✅ Test ProfileManager parsing
4. ✅ Test button presses with profile cycles
5. ✅ Test blink patterns (visual confirmation)
6. ✅ Test buzzer chirp pattern (audio confirmation)
7. ✅ Test state persistence across reboots

### Backend Integration Testing
1. ✅ Assign profile to line via backend API
2. ✅ Device heartbeat with profile metadata
3. ✅ Backend responds with profile update
4. ✅ Device downloads and applies profile
5. ✅ Button press cycles through profile states
6. ✅ Verify outputs match profile configuration
7. ✅ Test profile update propagation
8. ✅ Test state migration (remove state from profile)
9. ✅ Test override flag and reset

### Web Interface Testing
1. ✅ Access /profile endpoint
2. ✅ Access /state endpoint
3. ✅ Manually set state via /state/set
4. ✅ Test output testing via /test/outputs
5. ✅ Clear override via /override/clear

## Troubleshooting

### Issue: Profile not loading
**Check:**
- NVS partition not corrupted: `profileStorage.hasProfile()`
- JSON valid: Test with online JSON validator
- ArduinoJson document size sufficient (2KB should be enough for most profiles)

### Issue: Blink patterns not working
**Check:**
- `outputController->update()` called in main loop
- `towerLight->update()` called in main loop
- Non-blocking timers not blocked by delay() calls

### Issue: State not persisting
**Check:**
- `profileStorage.setCurrentState()` called on state changes
- NVS namespace "profile" accessible
- NVS partition not full

### Issue: Profile download fails
**Check:**
- Network connected
- API base URL correct
- Profile ID valid
- HTTP timeout sufficient (30s default)

## Default Profile (Fallback)

If no profile is loaded, create a default hardcoded profile in ProfileManager:

```cpp
String createDefaultProfile() {
    return R"({
        "id": "default-profile",
        "name": "Default Profile",
        "version": 1,
        "states": [
            {
                "name": "On",
                "outputs": {
                    "redLight": "off",
                    "yellowLight": "off",
                    "greenLight": "on",
                    "buzzer": "off"
                }
            },
            {
                "name": "Off",
                "outputs": {
                    "redLight": "on",
                    "yellowLight": "off",
                    "greenLight": "off",
                    "buzzer": "off"
                }
            },
            {
                "name": "Maintenance",
                "outputs": {
                    "redLight": "off",
                    "yellowLight": "on",
                    "greenLight": "off",
                    "buzzer": "off"
                }
            }
        ],
        "buttonBehavior": {
            "shortPressCycle": ["On", "Off"],
            "longPressCycle": ["Maintenance"]
        },
        "defaultState": "Off"
    })";
}
```

## Migration Path

### For Existing Deployed Devices

1. **Firmware Update**: Deploy new firmware with signal profiles support
2. **Backend Setup**: Create default profile matching current behavior
3. **Line Assignment**: Assign default profile to all lines
4. **Device Sync**: Devices download profile on next heartbeat
5. **Verification**: Confirm all devices updated successfully
6. **Custom Profiles**: Create and assign custom profiles as needed

### Backward Compatibility

The refactored components maintain backward compatibility:
- LineState enum still exists (legacy methods convert to/from strings)
- Devices without profiles use default behavior
- MQTT messages include both legacy and new fields
- Gradual rollout possible (some devices with profiles, some without)

## Next Steps

After integration:
1. Test firmware compilation
2. Deploy to test device
3. Test basic functionality (button presses, outputs)
4. Connect to backend and test profile sync
5. Create custom profile via web UI
6. Assign to line and verify device updates
7. Test all edge cases (offline, state migration, override reset)

## Code Examples

See `/firmware/examples/signal_profiles_test/` for standalone testing examples (TODO: create these).

## Support

For issues or questions:
- Check logs via Serial monitor (115200 baud)
- Verify NVS partition with `profileStorage.getStorageStats()`
- Test profile JSON with online validators
- Review backend API logs for sync issues
