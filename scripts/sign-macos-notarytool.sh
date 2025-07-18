#!/bin/bash

# Modern script to sign and notarize get-repo for macOS using notarytool
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
if [ -z "$DEVELOPER_ID" ] || [ -z "$APPLE_ID" ] || [ -z "$APP_PASSWORD" ] || [ -z "$BUNDLE_ID" ] || [ -z "$TEAM_ID" ]; then
    echo "Error: Missing required configuration in sign-config.sh"
    exit 1
fi

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Create dist directory
mkdir -p dist

echo -e "${YELLOW}Building get-repo...${NC}"
go build -ldflags="-s -w" -o "dist/${BINARY_NAME}" ./cmd/get-repo

echo -e "${YELLOW}Signing binary with hardened runtime...${NC}"
codesign --force \
    --options runtime \
    --timestamp \
    --sign "${DEVELOPER_ID}" \
    --identifier "${BUNDLE_ID}" \
    "dist/${BINARY_NAME}"

echo -e "${YELLOW}Verifying signature...${NC}"
codesign --verify --deep --verbose "dist/${BINARY_NAME}"
spctl -a -t exec -vvv "dist/${BINARY_NAME}"

echo -e "${YELLOW}Creating ZIP for notarization...${NC}"
ditto -c -k --keepParent "dist/${BINARY_NAME}" "dist/${BINARY_NAME}-notarization.zip"

echo -e "${YELLOW}Submitting for notarization using notarytool...${NC}"
xcrun notarytool submit \
    "dist/${BINARY_NAME}-notarization.zip" \
    --apple-id "${APPLE_ID}" \
    --password "${APP_PASSWORD}" \
    --team-id "${TEAM_ID}" \
    --wait \
    --progress

echo -e "${YELLOW}Stapling notarization ticket to binary...${NC}"
xcrun stapler staple "dist/${BINARY_NAME}"

echo -e "${YELLOW}Verifying notarization...${NC}"
spctl -a -t exec -vvv "dist/${BINARY_NAME}"
xcrun stapler validate "dist/${BINARY_NAME}"

echo -e "${GREEN}✅ Binary signed and notarized successfully!${NC}"
echo -e "${GREEN}Binary location: dist/${BINARY_NAME}${NC}"

# Create final distribution ZIP
echo -e "${YELLOW}Creating distribution archive...${NC}"
cd dist
zip "${BINARY_NAME}-macos-signed.zip" "${BINARY_NAME}"
cd ..

echo -e "${GREEN}Distribution archive: dist/${BINARY_NAME}-macos-signed.zip${NC}"

# Cleanup
rm -f "dist/${BINARY_NAME}-notarization.zip"