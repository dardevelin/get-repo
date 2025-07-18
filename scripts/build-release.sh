#!/bin/bash

# Build release binaries for multiple platforms

set -e

VERSION=${1:-"dev"}
BINARY_NAME="get-repo"

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

# Create dist directory
mkdir -p dist

# Build flags
LDFLAGS="-s -w -X main.version=${VERSION}"

echo -e "${YELLOW}Building release binaries for version ${VERSION}...${NC}"

# macOS ARM64
echo -e "${YELLOW}Building macOS ARM64...${NC}"
GOOS=darwin GOARCH=arm64 go build -ldflags="${LDFLAGS}" -o "dist/${BINARY_NAME}-darwin-arm64" ./cmd/get-repo

# macOS AMD64
echo -e "${YELLOW}Building macOS AMD64...${NC}"
GOOS=darwin GOARCH=amd64 go build -ldflags="${LDFLAGS}" -o "dist/${BINARY_NAME}-darwin-amd64" ./cmd/get-repo

# Linux AMD64
echo -e "${YELLOW}Building Linux AMD64...${NC}"
GOOS=linux GOARCH=amd64 go build -ldflags="${LDFLAGS}" -o "dist/${BINARY_NAME}-linux-amd64" ./cmd/get-repo

# Linux ARM64
echo -e "${YELLOW}Building Linux ARM64...${NC}"
GOOS=linux GOARCH=arm64 go build -ldflags="${LDFLAGS}" -o "dist/${BINARY_NAME}-linux-arm64" ./cmd/get-repo

# Windows AMD64
echo -e "${YELLOW}Building Windows AMD64...${NC}"
GOOS=windows GOARCH=amd64 go build -ldflags="${LDFLAGS}" -o "dist/${BINARY_NAME}-windows-amd64.exe" ./cmd/get-repo

# Create universal binary for macOS
echo -e "${YELLOW}Creating macOS universal binary...${NC}"
lipo -create -output "dist/${BINARY_NAME}-darwin-universal" \
    "dist/${BINARY_NAME}-darwin-amd64" \
    "dist/${BINARY_NAME}-darwin-arm64"

# Sign macOS binaries if sign-config.sh exists
if [ -f "scripts/sign-config.sh" ]; then
    echo -e "${YELLOW}Signing macOS binaries...${NC}"
    
    # Sign the universal binary
    cp "dist/${BINARY_NAME}-darwin-universal" "dist/${BINARY_NAME}"
    ./scripts/sign-macos-notarytool.sh
    mv "dist/${BINARY_NAME}-macos-signed.zip" "dist/${BINARY_NAME}-${VERSION}-macos-universal-signed.zip"
    rm "dist/${BINARY_NAME}"
    
    # Also create unsigned archives
    cd dist
    zip "${BINARY_NAME}-${VERSION}-macos-universal.zip" "${BINARY_NAME}-darwin-universal"
    zip "${BINARY_NAME}-${VERSION}-macos-arm64.zip" "${BINARY_NAME}-darwin-arm64"
    zip "${BINARY_NAME}-${VERSION}-macos-amd64.zip" "${BINARY_NAME}-darwin-amd64"
    cd ..
else
    echo -e "${YELLOW}Skipping signing (sign-config.sh not found)${NC}"
    # Create unsigned archives
    cd dist
    zip "${BINARY_NAME}-${VERSION}-macos-universal.zip" "${BINARY_NAME}-darwin-universal"
    zip "${BINARY_NAME}-${VERSION}-macos-arm64.zip" "${BINARY_NAME}-darwin-arm64" 
    zip "${BINARY_NAME}-${VERSION}-macos-amd64.zip" "${BINARY_NAME}-darwin-amd64"
    cd ..
fi

# Create archives for other platforms
cd dist
tar czf "${BINARY_NAME}-${VERSION}-linux-amd64.tar.gz" "${BINARY_NAME}-linux-amd64"
tar czf "${BINARY_NAME}-${VERSION}-linux-arm64.tar.gz" "${BINARY_NAME}-linux-arm64"
zip "${BINARY_NAME}-${VERSION}-windows-amd64.zip" "${BINARY_NAME}-windows-amd64.exe"
cd ..

echo -e "${GREEN}✅ Release builds complete!${NC}"
echo -e "${GREEN}Binaries in: dist/${NC}"

# Generate checksums
echo -e "${YELLOW}Generating checksums...${NC}"
cd dist
shasum -a 256 *.zip *.tar.gz > "${BINARY_NAME}-${VERSION}-checksums.txt"
cd ..

echo -e "${GREEN}✅ Checksums generated: dist/${BINARY_NAME}-${VERSION}-checksums.txt${NC}"