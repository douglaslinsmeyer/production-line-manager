#!/bin/bash
# Generate MD5 checksum for firmware binary

FIRMWARE_PATH=".pio/build/esp32-s3-devkitc-1/firmware.bin"

if [ ! -f "$FIRMWARE_PATH" ]; then
    echo "ERROR: Firmware not found at $FIRMWARE_PATH"
    echo "Run 'pio run' first to build firmware"
    exit 1
fi

# Generate MD5 (works on both Linux and macOS)
if command -v md5sum &> /dev/null; then
    MD5=$(md5sum "$FIRMWARE_PATH" | cut -d' ' -f1)
elif command -v md5 &> /dev/null; then
    MD5=$(md5 -q "$FIRMWARE_PATH")
else
    echo "ERROR: No MD5 utility found (md5sum or md5)"
    exit 1
fi

# Get file size
if [[ "$OSTYPE" == "darwin"* ]]; then
    SIZE=$(stat -f%z "$FIRMWARE_PATH")
else
    SIZE=$(stat -c%s "$FIRMWARE_PATH")
fi

# Display results
echo "========================================="
echo "  Firmware MD5 Checksum"
echo "========================================="
echo "File: $FIRMWARE_PATH"
echo "Size: $SIZE bytes ($(echo "scale=2; $SIZE / 1024 / 1024" | bc) MB)"
echo "MD5:  $MD5"
echo ""
echo "Use this checksum when uploading via web interface"
echo "========================================="

# Save to file
echo "$MD5" > "$FIRMWARE_PATH.md5"
echo ""
echo "MD5 saved to: $FIRMWARE_PATH.md5"
