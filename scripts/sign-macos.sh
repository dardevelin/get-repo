#!/bin/bash

# Script to sign and notarize get-repo for macOS
# Requires Apple Developer account and certificates

set -e

# Configuration
BINARY_NAME="get-repo"

# Load configuration
if [ -f "scripts/sign-config.sh" ]; then
    source scripts/sign-config.sh
else
    echo "Error: scripts/sign-config.sh not found!"
    echo "Copy scripts/sign-config.sh.example to scripts/sign-config.sh and update with your credentials."
    exit 1
fi

# Verify required variables
if [ -z "$DEVELOPER_ID" ] || [ -z "$APPLE_ID" ] || [ -z "$APP_PASSWORD" ] || [ -z "$BUNDLE_ID" ]; then
    echo "Error: Missing required configuration in sign-config.sh"
    exit 1
fi

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${YELLOW}Building get-repo...${NC}"
go build -ldflags="-s -w" -o dist/${BINARY_NAME} ./cmd/get-repo

echo -e "${YELLOW}Signing binary...${NC}"
codesign --force --options runtime --timestamp --sign "${DEVELOPER_ID}" "dist/${BINARY_NAME}"

echo -e "${YELLOW}Verifying signature...${NC}"
codesign --verify --verbose "dist/${BINARY_NAME}"

echo -e "${YELLOW}Creating ZIP for notarization...${NC}"
ditto -c -k --keepParent "dist/${BINARY_NAME}" "dist/${BINARY_NAME}.zip"

echo -e "${YELLOW}Submitting for notarization...${NC}"
xcrun altool --notarize-app \
    --primary-bundle-id "${BUNDLE_ID}" \
    --username "${APPLE_ID}" \
    --password "${APP_PASSWORD}" \
    --file "dist/${BINARY_NAME}.zip" \
    --output-format json > notarization.json

# Extract request UUID
REQUEST_UUID=$(python3 -c "import json; print(json.load(open('notarization.json'))['notarization-upload']['RequestUUID'])")
echo -e "${GREEN}Notarization request submitted. Request UUID: ${REQUEST_UUID}${NC}"

echo -e "${YELLOW}Waiting for notarization to complete...${NC}"
echo "This may take a few minutes..."

# Wait for notarization
while true; do
    sleep 30
    xcrun altool --notarization-info "${REQUEST_UUID}" \
        --username "${APPLE_ID}" \
        --password "${APP_PASSWORD}" \
        --output-format json > status.json
    
    STATUS=$(python3 -c "import json; print(json.load(open('status.json'))['notarization-info']['Status'])")
    
    if [ "$STATUS" = "success" ]; then
        echo -e "${GREEN}Notarization successful!${NC}"
        break
    elif [ "$STATUS" = "invalid" ]; then
        echo -e "${RED}Notarization failed!${NC}"
        cat status.json
        exit 1
    else
        echo "Status: $STATUS - waiting..."
    fi
done

echo -e "${YELLOW}Stapling notarization ticket to binary...${NC}"
xcrun stapler staple "dist/${BINARY_NAME}"

echo -e "${GREEN}Binary signed and notarized successfully!${NC}"
echo -e "${GREEN}Binary location: dist/${BINARY_NAME}${NC}"

# Cleanup
rm -f notarization.json status.json "dist/${BINARY_NAME}.zip"