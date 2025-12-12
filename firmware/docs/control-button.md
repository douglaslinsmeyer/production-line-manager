# Control Button Feature

## Overview

The ESP32-S3 firmware supports a lighted momentary button for physical control of production line state. The button provides:
- **Short press (<5s)**: Toggle production line state
- **Long press (≥5s)**: Enter maintenance mode
- **Visual feedback**: Integrated LED reflects current state

## Hardware Configuration

### Button Input
- **Channel**: DIN1 (Digital Input 1)
- **GPIO**: GPIO4
- **Type**: Momentary pushbutton (Normally Open contact)
- **Wiring**:
  - Button COM → DIN1 COM terminal
  - Button NO → DIN1 DGND terminal

### Button LED Output
- **Channel**: EXIO5 (Digital Output 5)
- **GPIO**: TCA9554PWR I2C expander, channel 4
- **Type**: Sinking transistor output
- **Wiring**:
  - External Power + → COM terminal
  - Button LED+ → COM terminal (same as power +)
  - Button LED- → DO5 terminal (EXIO5)
  - External Power - → DGND terminal

**LED Power Supply**: 5-40V DC (typically 12V or 24V)

## Digital Input Configuration - CRITICAL

### INPUT_PULLUP Required

**IMPORTANT**: Digital inputs MUST be configured with internal pull-up resistors:

```cpp
pinMode(DIN_PINS[i], INPUT_PULLUP);  // ✅ CORRECT
```

**NOT:**
```cpp
pinMode(DIN_PINS[i], INPUT);  // ❌ WRONG - causes floating inputs
```

### Why Pull-Ups Are Required

The Waveshare ESP32-S3-POE-ETH-8DI-8DO uses optocoupler-isolated inputs. Without pull-ups:

1. **Floating Input Problem**: When no external power is connected, inputs float
2. **Electrical Noise**: Floating inputs pick up electromagnetic interference
3. **Rapid Oscillation**: Inputs toggle between 0 and 1 continuously
4. **Debounce Failure**: Changes occur faster than 50ms debounce period
5. **No Stable State**: Callback never fires because state never stabilizes

### With INPUT_PULLUP Enabled

1. **Default State**: Internal ~45kΩ pull-up resistor pulls input HIGH
2. **Button Not Pressed**: Input reads HIGH (1)
3. **Button Pressed**: Contact closes, pulls input to DGND (LOW/0)
4. **Button Released**: Pull-up returns input to HIGH (1)
5. **Clean Transitions**: Debounce works correctly, callbacks fire reliably

### Reference

Based on official Waveshare demo code:
- File: `docs/demo-library/Arduino/examples/MAIN_MQTT_ALL/WS_DIN.cpp`
- Lines 134-141: All DIN pins initialized with `INPUT_PULLUP`

## Button Behavior

### State Transitions (Short Press)

| Current State | Short Press → New State |
|--------------|------------------------|
| ON | OFF |
| OFF | ON |
| MAINTENANCE | ON |
| ERROR | ON |
| UNKNOWN | ON |

### Long Press (5 seconds)

Any state → **MAINTENANCE**

## LED Patterns

| Production Line State | LED Behavior |
|----------------------|--------------|
| ON | Solid ON |
| OFF | OFF |
| MAINTENANCE | Blinking (500ms on/off) |
| ERROR | Fast blinking (200ms on/off) |
| UNKNOWN | OFF |

## State Persistence

Production line state is stored in NVS and survives:
- Power cycles
- Firmware updates (if NVS partition unchanged)
- ESP32 resets

**NVS Namespace**: `linestate`
**NVS Key**: `current`

## MQTT Integration

### Status Publishing

Line state is included in all MQTT status messages:

**Topic**: `devices/{MAC}/status`

**Payload**:
```json
{
  "device_id": "30:ED:A0:22:D3:A4",
  "line_state": "ON",
  "digital_inputs": 0x00,
  "digital_outputs": 0x10,
  "network_connected": true,
  "timestamp": 123456
}
```

### State Synchronization

Button presses trigger immediate status publish (don't wait for 30s heartbeat).

API can send state commands back to firmware:

**Topic**: `devices/{MAC}/command`

**Payload**:
```json
{
  "command": "set_line_state",
  "state": "ON"
}
```

Valid states: `ON`, `OFF`, `MAINTENANCE`, `ERROR`

## Offline Operation

The control button works without network connectivity:
- State changes applied immediately (firmware-local control)
- State persisted to NVS
- Status published when MQTT reconnects
- API backend syncs automatically

## Testing

### Functional Test

1. **Short press test**:
   - Press button quickly (<5s)
   - Verify state toggles (ON↔OFF)
   - Verify LED changes immediately
   - Check MQTT status message includes new state

2. **Long press test**:
   - Press and hold button for 5+ seconds
   - Verify state changes to MAINTENANCE at 5s mark
   - Verify LED starts blinking (500ms)
   - Check serial log: "Long press detected (5 seconds)"

3. **Persistence test**:
   - Set state to ON via button
   - Power cycle device (unplug/replug)
   - Verify state loads as ON from NVS on boot
   - Verify LED shows correct state immediately

### Serial Monitor Output

Expected messages when pressing button:

```
Input change: CH1 = LOW
Control button detected on CH1
Control button pressed
Input change: CH1 = HIGH
Control button released after 1234 ms
Short press detected

========================================
  LINE STATE CHANGED: UNKNOWN -> ON
========================================

Button LED pattern changed: UNKNOWN -> ON
Button LED: Solid ON
Published status: line_state=ON inputs=0x00 outputs=0x10
```

## Implementation Files

### New Modules
- `firmware/src/state/line_state.h` - State manager interface
- `firmware/src/state/line_state.cpp` - State transitions and NVS persistence
- `firmware/src/gpio/control_button.h` - Button handler interface
- `firmware/src/gpio/control_button.cpp` - Press detection (short/long)
- `firmware/src/gpio/button_led.h` - LED controller interface
- `firmware/src/gpio/button_led.cpp` - Non-blocking LED patterns

### Modified Files
- `firmware/src/gpio/digital_input.cpp` - Changed INPUT to INPUT_PULLUP
- `firmware/src/main.cpp` - Integration and callbacks
- `firmware/src/config.h` - Control button constants
- `firmware/src/mqtt/mqtt_client.h` - Added LineState parameter
- `firmware/src/mqtt/mqtt_client.cpp` - Line state in status messages + command handler

## Troubleshooting

### Button not responding

**Symptom**: Pressing button has no effect, no serial output

**Cause**: Digital inputs not configured with INPUT_PULLUP

**Solution**: Verify firmware/src/gpio/digital_input.cpp line 27:
```cpp
pinMode(DIN_PINS[i], INPUT_PULLUP);  // Must be INPUT_PULLUP
```

### Inputs oscillating/noisy

**Symptom**: Serial monitor shows rapid input changes without button press:
```
[DEBUG] CH1 reading changed: 1 -> 0
[DEBUG] CH1 reading changed: 0 -> 1
[DEBUG] CH1 reading changed: 1 -> 0
...
```

**Cause**: INPUT_PULLUP not enabled (floating inputs)

**Solution**: Same as above - enable INPUT_PULLUP mode

### Button LED not lighting

**Symptom**: Button works but LED doesn't illuminate

**Possible causes**:
1. **No external power**: LED needs 5-40V power supply to COM/DGND
2. **Wrong polarity**: LED+ should connect to COM (+), LED- to DO5
3. **Wrong output channel**: Verify LED connected to EXIO5 (DO5), not another output
4. **State issue**: Check line_state in MQTT messages - LED follows state

**Debug**:
```bash
# Turn on EXIO5 manually via MQTT
curl -X POST http://localhost:15672/api/exchanges/%2f/amq.topic/publish \
  -u guest:guest \
  -H "Content-Type: application/json" \
  -d '{
    "routing_key": "devices.{MAC}.command",
    "payload": "{\"command\":\"set_output\",\"channel\":4,\"state\":true}"
  }'
```

### State not syncing to API

**Symptom**: Button changes state but API doesn't show it

**Possible causes**:
1. **MQTT not connected**: Check network connectivity
2. **Topic mismatch**: Verify MAC address format (30:ED:A0:22:D3:A4)
3. **API not listening**: Check API logs for status messages

**Debug**: Monitor MQTT topic `devices/{MAC}/status` for line_state field

## Configuration Constants

From `firmware/src/config.h`:

```cpp
// Control Button
#define CONTROL_BUTTON_CHANNEL 0           // DIN1 (0-indexed)
#define CONTROL_BUTTON_GPIO 4              // GPIO4
#define CONTROL_BUTTON_LONG_PRESS 5000     // 5 seconds
#define BUTTON_LED_CHANNEL 4               // EXIO5 (0-indexed)

// LED Patterns
#define BUTTON_LED_MAINTENANCE_PERIOD 500  // 500ms blink
#define BUTTON_LED_ERROR_PERIOD 200        // 200ms fast blink
```
