# Signal Profiles Specification

## Overview

Signal Profiles are configurable templates that define how assembly line devices respond to different operational states. Each profile specifies a collection of states, the visual and audio signals for each state, and how the device's button cycles through those states.

## Core Concepts

### Signal Profile
A named configuration template containing:
- A collection of **States** (e.g., "On", "Off", "Maintenance", "Error")
- Visual and audio **Output Configuration** for each state
- **Button Behavior** defining how button presses cycle through states

### State
A named operational condition (e.g., "On", "Off", "Maintenance") with specific output configurations:
- Tower Red Light behavior
- Tower Yellow Light behavior
- Tower Green Light behavior
- Buzzer behavior

### Profile Assignment
- Profiles are assigned to **Lines** in the web backend
- Devices inherit their profile when assigned to a line
- Devices store the profile configuration locally for offline operation

## Output Configuration

### Light States
Each light (Red, Yellow, Green) can be configured to:
- **Off** - Light is off
- **On** - Light is steady on
- **Short Blink** - Light blinks with 500ms intervals (500ms on, 500ms off)
- **Long Blink** - Light blinks with 1500ms intervals (1500ms on, 1500ms off)

### Buzzer States
The buzzer can be configured to:
- **Off** - Silent
- **On** - Continuous tone
- **Chirp** - Pattern: beep, pause, beep, pause, beep, then 3000ms delay before repeating

## Button Behavior

The device has one primary button (DI1) with two press types:

### Short Press
Cycles through a configured list of states in sequence. When the last state is reached, the next short press returns to the first state in the list.

Example: `["On", "Off"]` - Short press toggles between "On" and "Off"

### Long Press
Cycles through a separate configured list of states. When the last state is reached, the next long press returns to the first state in the list.

Example: `["Maintenance", "Error"]` - Long press toggles between "Maintenance" and "Error"

### Press Detection
- **Short Press**: Button released in less than 1 second
- **Long Press**: Button held for 1 second or more

## Example Profile

### Assembly Line Profile

**States:**
1. On
2. Off
3. Maintenance
4. Error

**State Configurations:**

#### On
- Tower Red Light: Off
- Tower Yellow Light: Off
- Tower Green Light: On (steady)
- Buzzer: Off

#### Off
- Tower Red Light: On (steady)
- Tower Yellow Light: Off
- Tower Green Light: Off
- Buzzer: Off

#### Maintenance
- Tower Red Light: Off
- Tower Yellow Light: On (steady)
- Tower Green Light: Off
- Buzzer: Off

#### Error
- Tower Red Light: Short Blink (500ms intervals)
- Tower Yellow Light: Off
- Tower Green Light: Off
- Buzzer: On

**Button Behavior:**
- Short Press Cycle: `["On", "Off"]`
- Long Press Cycle: `["Maintenance"]`

## Profile Management

### Backend Management
- Profiles are created and edited in the web backend
- Profiles can be assigned to specific lines
- The web UI provides a profile editor with visual preview of each state

### Device Adoption
When a device is assigned to a line:
1. Device receives the line's signal profile configuration via API
2. Device stores the profile locally in persistent storage
3. Device defaults to the first state in the short press cycle (or a designated default state)

### Offline Operation
- Devices continue operating with their stored profile when offline
- State changes via button presses work offline
- When connection is restored, device syncs current state to backend

## Device Overrides

### Manual Override
The device's configuration web UI allows users to:
- View the current signal profile configuration
- View the current active state
- Manually select any state from the profile
- Directly control individual outputs (Red/Yellow/Green/Buzzer) for testing

### Override Persistence
- Manual state overrides persist through device reboot
- Device maintains an "override flag" to track if current state differs from line default
- Override flag is stored in persistent storage

### Override Management in Web Backend
The web backend management interface should:
- Display an indicator for devices that have active overrides
- Show what state the device is currently in vs. what the line default would be
- Provide a "Reset to Line Profile" action that:
  - Clears the override flag
  - Returns device to the line's default state
  - Can be performed on individual devices or bulk operation on multiple devices

## Profile Versioning

### Overview
Profile versioning ensures all devices using a signal profile stay synchronized with the latest configuration. When a profile is updated in the backend, all assigned devices automatically receive and apply the changes.

### Version Numbering
- Each profile has a **version number** (integer, starting at 1)
- Version increments automatically when profile configuration changes
- Version is part of the profile's immutable metadata
- Each version increment creates a new version record in history

### When Versions Increment
Version numbers increment when any of the following changes:
- State added, removed, or renamed
- Output configuration for any state changes (light or buzzer settings)
- Button behavior cycles modified
- Default state changes

Version numbers **do not** increment for:
- Profile name or description changes (metadata only)
- Assignment to different lines

### Device Version Tracking
Each device stores:
- **Current Profile ID** - The UUID of the assigned profile
- **Current Profile Version** - The version number of the stored profile
- **Profile Hash** (optional) - MD5 hash of profile JSON for validation
- **Last Version Check** - Timestamp of last sync with backend

### Version Sync Mechanism

#### Periodic Version Check
Devices check for profile updates on a regular interval:
1. **Device sends heartbeat** to backend including:
   - Device ID
   - Current profile ID
   - Current profile version
   - Current state
2. **Backend responds** with:
   - Latest profile version number
   - Update available flag (boolean)
   - If update available, the full updated profile configuration
3. **Device compares versions**:
   - If versions match: no action needed
   - If backend version is higher: apply update (see Update Process below)

**Recommended sync interval**: Every 60 seconds when online

#### On-Connection Sync
When a device connects to the backend (after being offline):
1. Device immediately sends heartbeat with current version
2. Backend responds with latest version
3. If out of sync, device downloads and applies update

#### Push Updates (Optional Enhancement)
For real-time updates, backend can push profile changes to online devices:
- Backend publishes profile update event via WebSocket/MQTT
- Online devices receive notification immediately
- Devices download and apply update without waiting for next heartbeat

### Update Process

When a device detects a profile version mismatch:

#### Step 1: Download New Profile
- Device requests full profile configuration: `GET /api/profiles/{id}`
- Backend returns complete profile with new version number
- Device validates JSON structure and required fields

#### Step 2: Handle Current State
The device must determine if its current state still exists in the new profile:

**Case A: Current state still exists in new profile**
- Device keeps current state name
- Device applies new output configuration for that state
- Override flag remains unchanged

**Case B: Current state removed from new profile**
- Device logs warning about missing state
- Device transitions to profile's default state
- If device had override flag, it is cleared (since state no longer valid)
- Device sends notification to backend about forced state change

**Case C: Current state renamed**
This is treated as removed + added, so follows Case B logic.

#### Step 3: Store New Profile
- Device saves complete new profile to persistent storage
- Device updates stored version number
- Device updates last version check timestamp

#### Step 4: Apply Outputs
- Device immediately applies output configuration for current state
- Hardware outputs update to reflect new settings

#### Step 5: Confirm to Backend
- Device sends confirmation: `POST /api/devices/{id}/profile-updated`
- Includes new version number and any state changes
- Backend marks device as synchronized

### Backend Version Management

#### Profile Update Flow
When an admin updates a profile in the backend:

1. **Validate Changes**
   - Ensure all state references in button cycles are valid
   - Check that default state exists in states array
   - Validate output configuration values

2. **Create New Version**
   - Increment version number
   - Store new version in profile_versions table
   - Update current profile record
   - Log change with timestamp and user

3. **Identify Affected Devices**
   - Query all lines using this profile
   - Query all devices assigned to those lines
   - Create list of devices needing update

4. **Track Update Progress**
   - Mark all affected devices as "update pending"
   - Record expected version for each device
   - Monitor as devices report updated versions

5. **Notify/Push Updates** (if real-time updates enabled)
   - Send push notification to all online affected devices
   - Devices receive and apply update immediately

#### Version History
Backend maintains complete version history:
- Profile ID
- Version number
- Full profile configuration (JSON snapshot)
- Timestamp
- User who made change
- Change description/notes
- List of what changed (diff)

This enables:
- Audit trail of all profile changes
- Rollback to previous versions if needed
- Comparison between versions
- Analysis of when devices updated

### Device Status Dashboard

The backend management UI should display:

#### Profile Version Status
For each profile, show:
- Current version number
- Last updated timestamp
- Number of devices using this profile
- Number of devices on current version
- Number of devices on outdated versions
- Number of devices with failed updates

#### Device Version Status
For each device, show:
- Current profile name and version
- Latest available version
- Version status indicator:
  - ✓ **Up to date** - Green
  - ↑ **Update pending** - Yellow
  - ⚠ **Update failed** - Red
  - ⏸ **Offline** - Gray
- Last version check timestamp

#### Bulk Operations
- **Force Update** - Push update to selected devices immediately
- **View Outdated** - Filter to show only devices needing updates
- **Update History** - View timeline of when devices updated

### Rollback Scenarios

#### Rolling Back a Profile
If a profile update causes issues, admins can rollback:

1. **Select Previous Version**
   - View version history for profile
   - Select a previous version to restore

2. **Create Rollback Version**
   - Copy configuration from selected historical version
   - Increment version number (don't revert version number)
   - This creates a new version with old configuration

3. **Propagate to Devices**
   - Standard update mechanism applies
   - Devices detect new (higher) version and update

**Important**: Never decrement version numbers, always increment. A rollback is a new version with old content.

### Version Mismatch Scenarios

#### Device on Newer Version Than Backend
This should not happen in normal operation, but if it does:
- Log error - indicates clock skew or corruption
- Device should download profile from backend and overwrite
- Treat backend as source of truth

#### Device Unable to Update
If device fails to apply update after multiple attempts:
- Device marks update as failed
- Device continues operating with old version
- Backend marks device with error status
- Admin receives alert about failed device
- Manual intervention may be required

### API Endpoints for Versioning

Add these endpoints to support versioning:

#### Version Check
```
POST /api/devices/{id}/heartbeat
Request Body:
{
  "profileId": "uuid",
  "profileVersion": 5,
  "currentState": "On",
  "isOverridden": false
}

Response:
{
  "latestVersion": 7,
  "updateAvailable": true,
  "profile": { /* full profile JSON if updateAvailable */ }
}
```

#### Confirm Update
```
POST /api/devices/{id}/profile-updated
Request Body:
{
  "profileId": "uuid",
  "newVersion": 7,
  "previousVersion": 5,
  "currentState": "On",
  "stateChanged": false
}

Response:
{
  "success": true
}
```

#### Version History
```
GET /api/profiles/{id}/versions
Response:
{
  "versions": [
    {
      "version": 7,
      "timestamp": "2026-01-06T10:30:00Z",
      "changedBy": "admin@example.com",
      "changes": ["Modified 'On' state green light to Short Blink"],
      "config": { /* full profile snapshot */ }
    },
    ...
  ]
}
```

#### Rollback Profile
```
POST /api/profiles/{id}/rollback
Request Body:
{
  "targetVersion": 5,
  "reason": "New blink pattern causing issues"
}

Response:
{
  "success": true,
  "newVersion": 8,
  "affectedDevices": 12
}
```

#### Device Version Status
```
GET /api/profiles/{id}/device-status
Response:
{
  "profileId": "uuid",
  "currentVersion": 7,
  "devices": [
    {
      "deviceId": "device-001",
      "version": 7,
      "status": "up-to-date",
      "lastCheck": "2026-01-06T10:45:00Z"
    },
    {
      "deviceId": "device-002",
      "version": 5,
      "status": "update-pending",
      "lastCheck": "2026-01-06T10:30:00Z"
    }
  ],
  "summary": {
    "upToDate": 10,
    "updatePending": 2,
    "failed": 0,
    "offline": 1
  }
}
```

### Implementation Considerations

#### Version Storage Efficiency
- Devices only store current profile version, not full history
- Backend stores complete history for audit and rollback
- Consider compression for historical profile snapshots

#### Network Efficiency
- Heartbeat response only includes full profile if update needed
- Consider delta updates (only changed fields) for large profiles
- Cache profiles in backend to reduce database queries

#### Timing Considerations
- Stagger device heartbeats to avoid thundering herd
- Add jitter to sync intervals (e.g., 60s ± 10s random)
- Batch device notifications when pushing updates

#### State Migration Strategies
For more sophisticated state migrations, consider:
- **State mapping** - Provide mapping from old state names to new
- **Transition rules** - Define how devices should transition when states removed
- **Validation mode** - Test profile on subset of devices before full rollout

## Data Structure

### Profile Object
```json
{
  "id": "string (UUID)",
  "name": "string",
  "description": "string (optional)",
  "version": "integer (starts at 1, auto-increments)",
  "createdAt": "timestamp",
  "updatedAt": "timestamp",
  "updatedBy": "string (user ID or email)",
  "states": [
    {
      "name": "string",
      "outputs": {
        "redLight": "off" | "on" | "shortBlink" | "longBlink",
        "yellowLight": "off" | "on" | "shortBlink" | "longBlink",
        "greenLight": "off" | "on" | "shortBlink" | "longBlink",
        "buzzer": "off" | "on" | "chirp"
      }
    }
  ],
  "buttonBehavior": {
    "shortPressCycle": ["string (state names)"],
    "longPressCycle": ["string (state names)"]
  },
  "defaultState": "string (state name)"
}
```

### Profile Version History Object
```json
{
  "id": "string (UUID)",
  "profileId": "string (UUID)",
  "version": "integer",
  "timestamp": "timestamp",
  "changedBy": "string (user ID or email)",
  "changeDescription": "string (optional)",
  "changes": ["array of change descriptions"],
  "config": {
    // Full profile configuration snapshot at this version
  }
}
```

### Device State Object
```json
{
  "deviceId": "string",
  "lineId": "string",
  "profileId": "string",
  "profileVersion": "integer",
  "profileHash": "string (MD5, optional)",
  "currentState": "string (state name)",
  "isOverridden": "boolean",
  "lastStateChange": "timestamp",
  "lastSync": "timestamp",
  "lastVersionCheck": "timestamp",
  "versionStatus": "up-to-date" | "update-pending" | "failed" | "offline"
}
```

### Line Configuration
```json
{
  "lineId": "string",
  "signalProfileId": "string (UUID)",
  "defaultState": "string (state name, optional)"
}
```

## API Endpoints

### Profile Management
- `GET /api/profiles` - List all profiles
- `POST /api/profiles` - Create new profile
- `GET /api/profiles/{id}` - Get profile details
- `PUT /api/profiles/{id}` - Update profile
- `DELETE /api/profiles/{id}` - Delete profile (only if not assigned to any lines)

### Line Profile Assignment
- `PUT /api/lines/{id}/profile` - Assign profile to line
- `GET /api/lines/{id}/profile` - Get line's current profile

### Device Operations
- `GET /api/devices/{id}/state` - Get current device state
- `PUT /api/devices/{id}/state` - Manually set device state (creates override)
- `POST /api/devices/{id}/reset-override` - Clear override and return to line default
- `GET /api/devices/overridden` - List all devices with active overrides
- `POST /api/devices/bulk-reset-override` - Clear overrides for multiple devices

### Device-Side Endpoints
The device should expose local endpoints for configuration:
- `GET /config/profile` - View current profile configuration
- `GET /config/state` - View current state
- `PUT /config/state` - Set current state (creates override flag)
- `POST /config/reset-override` - Clear override flag
- `PUT /config/outputs/test` - Directly control outputs for testing

## Implementation Considerations

### State Transition Logic
1. When button is pressed, determine press type (short/long)
2. Get current state name
3. Find current state position in appropriate cycle array
4. Calculate next state (wrap to 0 if at end)
5. Look up new state's output configuration
6. Apply outputs to hardware
7. Save current state to persistent storage
8. Set override flag (since state changed manually)
9. Sync state change to backend when online

### Hardware Output Timing
- Light blink patterns run in non-blocking timers
- Buzzer chirp pattern runs on separate timer
- Each output controlled independently
- Timer state persists across state changes (e.g., blink phase continues)

### Profile Sync Strategy
1. When device connects to backend, check profile version
2. If profile has been updated on backend, download new version
3. If device has override, maintain current state but update available states
4. If no override, apply line's default state

### Error Handling
- If device receives profile with invalid state references in button cycles, log error and default to first state only
- If backend cannot reach device for override reset, mark as pending and retry on next connection
- If profile is deleted but still assigned to lines, prevent deletion and return error

### Migration & Versioning
See the dedicated **Profile Versioning** section above for complete details on:
- Version numbering and increment rules
- Device sync mechanisms and update processes
- Version history and rollback procedures
- API endpoints for version management

Additional considerations:
- When adding new output types in future versions, ensure backward compatibility
- Devices on older firmware should gracefully handle unknown output types
- Consider feature flags for experimental output patterns

## Future Enhancements

### Potential Future Features
- **Scheduled State Changes** - Automatic state transitions based on time
- **Sensor-Triggered States** - Automatic state changes based on sensor inputs
- **Multi-Device Coordination** - Synchronized state changes across multiple devices
- **Custom Blink Patterns** - User-defined timing for blink intervals
- **Audio Patterns** - More buzzer pattern options beyond chirp
- **State History** - Tracking and analytics of state changes over time
- **Conditional Cycles** - Different button cycles based on current state or conditions
