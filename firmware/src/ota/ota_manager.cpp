#include "ota_manager.h"
#include "config.h"

OTAManager::OTAManager()
    : state(OTA_IDLE),
      lastError(OTA_ERROR_NONE),
      bytesWritten(0),
      totalSize(0),
      expectedMD5(""),
      updateStartTime(0),
      bootCount(0),
      progressCallback(nullptr),
      lastProgressPercent(0) {
}

OTAManager::~OTAManager() {
    bootPrefs.end();
}

bool OTAManager::begin() {
    Serial.println("Initializing OTA manager...");

    // Open NVS namespace for boot tracking
    bootPrefs.begin("ota_boot", false);

    // Load and increment boot count
    bootCount = bootPrefs.getUInt("boot_cnt", 0);
    bootCount++;
    bootPrefs.putUInt("boot_cnt", bootCount);

    Serial.printf("  Boot count: %u\n", bootCount);

    // Check rollback threshold
    if (bootCount >= OTA_BOOT_COUNT_THRESHOLD) {
        Serial.println("⚠ WARNING: Multiple boot failures detected!");
        Serial.printf("  Boot count: %u (threshold: %u)\n",
                     bootCount, OTA_BOOT_COUNT_THRESHOLD);
        Serial.println("  Recent firmware may be unstable");
    }

    return true;
}

bool OTAManager::startUpdate(size_t expectedSize, const char* md5) {
    if (state != OTA_IDLE) {
        Serial.println("OTA: ERROR - Update already in progress!");
        setError(OTA_ERROR_INVALID_STATE);
        return false;
    }

    Serial.printf("OTA: Starting update (size: %u bytes, %.2f MB)\n",
                 expectedSize, expectedSize / 1024.0 / 1024.0);

    // Check if firmware size is valid
    if (expectedSize == 0 || expectedSize > OTA_MAX_SIZE) {
        Serial.printf("OTA: ERROR - Invalid size: %u bytes\n", expectedSize);
        setError(OTA_ERROR_INVALID_FIRMWARE);
        return false;
    }

    // Check available space
    if (!checkSpace(expectedSize)) {
        setError(OTA_ERROR_NO_SPACE);
        return false;
    }

    // Set MD5 if provided
    if (md5 && strlen(md5) == 32) {
        expectedMD5 = String(md5);
        Update.setMD5(md5);
        Serial.printf("OTA: MD5 validation enabled: %s\n", md5);
    } else {
        expectedMD5 = "";
        Serial.println("OTA: MD5 validation disabled");
    }

    // Begin update
    if (!Update.begin(expectedSize, U_FLASH)) {
        Serial.printf("OTA: ERROR - Begin failed: %s\n", Update.errorString());
        setError(OTA_ERROR_BEGIN_FAILED);
        return false;
    }

    state = OTA_IN_PROGRESS;
    bytesWritten = 0;
    totalSize = expectedSize;
    updateStartTime = millis();
    lastProgressPercent = 0;
    lastError = OTA_ERROR_NONE;

    Serial.println("OTA: Update started successfully");
    notifyProgress();

    return true;
}

bool OTAManager::writeChunk(uint8_t* data, size_t len) {
    if (state != OTA_IN_PROGRESS) {
        Serial.println("OTA: ERROR - No update in progress");
        return false;
    }

    // Validate firmware header on first chunk
    if (bytesWritten == 0) {
        if (!validateFirmwareHeader(data, len)) {
            setError(OTA_ERROR_INVALID_FIRMWARE);
            abortUpdate();
            return false;
        }
    }

    // Write chunk
    size_t written = Update.write(data, len);
    if (written != len) {
        Serial.printf("OTA: ERROR - Write failed (wrote %u of %u bytes)\n", written, len);
        setError(OTA_ERROR_WRITE_FAILED);
        abortUpdate();
        return false;
    }

    bytesWritten += written;

    // Notify progress if changed by 1% or more
    uint8_t currentPercent = getProgressPercent();
    if (currentPercent != lastProgressPercent) {
        lastProgressPercent = currentPercent;
        notifyProgress();
    }

    return true;
}

bool OTAManager::finishUpdate() {
    if (state != OTA_IN_PROGRESS) {
        Serial.println("OTA: ERROR - No update in progress");
        return false;
    }

    Serial.println("OTA: Finalizing update...");

    // Check size matches
    if (bytesWritten != totalSize) {
        Serial.printf("OTA: ERROR - Size mismatch (expected %u, got %u)\n",
                     totalSize, bytesWritten);
        setError(OTA_ERROR_SIZE_MISMATCH);
        abortUpdate();
        return false;
    }

    // Finalize update (true = set new boot partition)
    if (!Update.end(true)) {
        Serial.printf("OTA: ERROR - End failed: %s\n", Update.errorString());

        // Check if it was MD5 failure
        if (Update.hasError() && expectedMD5.length() > 0) {
            setError(OTA_ERROR_CHECKSUM_FAILED);
            Serial.println("OTA: MD5 checksum validation FAILED!");
        } else {
            setError(OTA_ERROR_END_FAILED);
        }
        return false;
    }

    // Verify MD5 if it was set
    if (expectedMD5.length() > 0) {
        if (Update.isFinished()) {
            Serial.println("OTA: MD5 checksum validation PASSED ✓");
        } else {
            setError(OTA_ERROR_CHECKSUM_FAILED);
            Serial.println("OTA: MD5 mismatch detected");
            return false;
        }
    }

    unsigned long duration = (millis() - updateStartTime) / 1000;
    Serial.printf("OTA: Update completed successfully in %lu seconds\n", duration);

    state = OTA_SUCCESS;
    lastError = OTA_ERROR_NONE;
    notifyProgress();

    return true;
}

void OTAManager::abortUpdate() {
    if (state == OTA_IN_PROGRESS) {
        Serial.println("OTA: Aborting update...");
        Update.abort();
        state = OTA_IDLE;
        bytesWritten = 0;
        totalSize = 0;
    }
}

uint8_t OTAManager::getProgressPercent() const {
    if (totalSize == 0) return 0;
    return (bytesWritten * 100) / totalSize;
}

const char* OTAManager::getErrorString() const {
    switch (lastError) {
        case OTA_ERROR_NONE:
            return "No error";
        case OTA_ERROR_NO_SPACE:
            return "Insufficient space in flash";
        case OTA_ERROR_INVALID_FIRMWARE:
            return "Invalid firmware file";
        case OTA_ERROR_WRITE_FAILED:
            return "Write to flash failed";
        case OTA_ERROR_BEGIN_FAILED:
            return "Failed to begin update";
        case OTA_ERROR_END_FAILED:
            return "Failed to finalize update";
        case OTA_ERROR_ABORT_FAILED:
            return "Failed to abort update";
        case OTA_ERROR_SIZE_MISMATCH:
            return "Firmware size mismatch";
        case OTA_ERROR_CHECKSUM_FAILED:
            return "MD5 checksum validation failed";
        case OTA_ERROR_INVALID_STATE:
            return "Invalid state for operation";
        default:
            return "Unknown error";
    }
}

String OTAManager::getStateString() const {
    switch (state) {
        case OTA_IDLE:
            return "idle";
        case OTA_STARTING:
            return "starting";
        case OTA_IN_PROGRESS:
            return "in_progress";
        case OTA_SUCCESS:
            return "success";
        case OTA_ERROR:
            return "error";
        default:
            return "unknown";
    }
}

void OTAManager::incrementBootCount() {
    bootCount++;
    bootPrefs.putUInt("boot_cnt", bootCount);
}

bool OTAManager::isRollbackNeeded() const {
    return (bootCount >= OTA_BOOT_COUNT_THRESHOLD);
}

void OTAManager::markValidBoot() {
    if (bootCount > 0) {
        bootCount = 0;
        bootPrefs.putUInt("boot_cnt", 0);
        Serial.println("✓ OTA: Boot validated - counter reset to 0");
    }
}

bool OTAManager::checkSpace(size_t size) {
    const esp_partition_t* partition = esp_ota_get_next_update_partition(NULL);

    if (!partition) {
        Serial.println("OTA: ERROR - No OTA partition found!");
        return false;
    }

    Serial.printf("OTA: Next partition: %s (0x%X, %u bytes)\n",
                 partition->label, partition->address, partition->size);

    if (size > partition->size) {
        Serial.printf("OTA: ERROR - Firmware too large!\n");
        Serial.printf("  Firmware size: %u bytes (%.2f MB)\n",
                     size, size / 1024.0 / 1024.0);
        Serial.printf("  Partition size: %u bytes (%.2f MB)\n",
                     partition->size, partition->size / 1024.0 / 1024.0);
        return false;
    }

    Serial.printf("OTA: Space check OK (%.1f%% of partition)\n",
                 (size * 100.0) / partition->size);
    return true;
}

bool OTAManager::validateFirmwareHeader(uint8_t* data, size_t len) {
    if (len < 4) {
        Serial.println("OTA: Header too short for validation");
        return false;
    }

    // ESP32 image magic byte
    if (data[0] != 0xE9) {
        Serial.printf("OTA: Invalid magic byte: 0x%02X (expected 0xE9)\n", data[0]);
        return false;
    }

    // Segment count
    uint8_t segments = data[1];
    if (segments == 0 || segments > 16) {
        Serial.printf("OTA: Invalid segment count: %u\n", segments);
        return false;
    }

    // Flash mode
    uint8_t flashMode = data[2];
    if (flashMode > 3) {
        Serial.printf("OTA: Invalid flash mode: %u\n", flashMode);
        return false;
    }

    Serial.println("✓ OTA: Firmware header validation passed");
    Serial.printf("  Segments: %u, Flash mode: %u\n", segments, flashMode);
    return true;
}

void OTAManager::setError(OTAError error) {
    lastError = error;
    state = OTA_ERROR;
    Serial.printf("OTA: ERROR - %s\n", getErrorString());
}

void OTAManager::notifyProgress() {
    if (progressCallback) {
        progressCallback(getProgressPercent(), state);
    }
}
