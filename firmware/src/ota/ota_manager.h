#pragma once

#include <Arduino.h>
#include <Update.h>
#include <esp_ota_ops.h>
#include <Preferences.h>

/**
 * OTA Manager
 *
 * Manages Over-The-Air firmware updates with:
 * - Chunked upload support
 * - MD5 checksum validation
 * - Boot count tracking for rollback detection
 * - Progress callbacks for UI feedback
 * - Automatic partition management
 */
class OTAManager {
public:
    enum OTAState {
        OTA_IDLE,
        OTA_STARTING,
        OTA_IN_PROGRESS,
        OTA_SUCCESS,
        OTA_ERROR
    };

    enum OTAError {
        OTA_ERROR_NONE,
        OTA_ERROR_NO_SPACE,
        OTA_ERROR_INVALID_FIRMWARE,
        OTA_ERROR_WRITE_FAILED,
        OTA_ERROR_BEGIN_FAILED,
        OTA_ERROR_END_FAILED,
        OTA_ERROR_ABORT_FAILED,
        OTA_ERROR_SIZE_MISMATCH,
        OTA_ERROR_CHECKSUM_FAILED,
        OTA_ERROR_INVALID_STATE
    };

    typedef void (*ProgressCallback)(uint8_t percent, OTAState state);

    OTAManager();
    ~OTAManager();

    /**
     * Initialize OTA manager
     * Loads boot count from NVS
     * @return true if initialization successful
     */
    bool begin();

    /**
     * Start firmware update
     * @param expectedSize Total size of firmware in bytes
     * @param md5 Optional MD5 checksum for validation (32-character hex string)
     * @return true if update started successfully
     */
    bool startUpdate(size_t expectedSize, const char* md5 = nullptr);

    /**
     * Write firmware chunk
     * @param data Pointer to firmware data
     * @param len Length of data in bytes
     * @return true if write successful
     */
    bool writeChunk(uint8_t* data, size_t len);

    /**
     * Finalize firmware update
     * Validates checksum if MD5 was provided
     * Sets new boot partition
     * @return true if update finalized successfully
     */
    bool finishUpdate();

    /**
     * Abort update and clean up
     */
    void abortUpdate();

    // State queries
    OTAState getState() const { return state; }
    OTAError getError() const { return lastError; }
    size_t getBytesWritten() const { return bytesWritten; }
    size_t getTotalSize() const { return totalSize; }
    uint8_t getProgressPercent() const;
    const char* getErrorString() const;
    String getStateString() const;

    // Boot management
    uint32_t getBootCount() const { return bootCount; }
    void incrementBootCount();
    bool isRollbackNeeded() const;
    void markValidBoot();

    // Callbacks
    void setProgressCallback(ProgressCallback callback) { progressCallback = callback; }

private:
    OTAState state;
    OTAError lastError;
    size_t bytesWritten;
    size_t totalSize;
    String expectedMD5;
    unsigned long updateStartTime;
    uint32_t bootCount;
    ProgressCallback progressCallback;
    Preferences bootPrefs;
    uint8_t lastProgressPercent;

    bool checkSpace(size_t size);
    bool validateFirmwareHeader(uint8_t* data, size_t len);
    void setError(OTAError error);
    void notifyProgress();
};
