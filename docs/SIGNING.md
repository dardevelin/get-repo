# Code Signing for macOS

This document describes how to sign and notarize get-repo for macOS distribution.

## Prerequisites

1. Apple Developer Account ($99/year)
2. Developer ID Application certificate installed in Keychain
3. App-specific password for notarization

## Setup

1. Copy the signing configuration template:
   ```bash
   cp scripts/sign-config.sh.example scripts/sign-config.sh
   ```

2. Edit `scripts/sign-config.sh` with your credentials:
   - `DEVELOPER_ID`: Your Developer ID certificate name (find with `security find-identity -v -p codesigning`)
   - `APPLE_ID`: Your Apple ID email
   - `APP_PASSWORD`: App-specific password (create at https://appleid.apple.com)
   - `TEAM_ID`: Your Apple Developer Team ID

3. Keep `scripts/sign-config.sh` private - it's in .gitignore

## Signing Process

### Manual Signing

To sign a single binary:
```bash
./scripts/sign-macos-notarytool.sh
```

This will:
1. Build the binary
2. Sign it with hardened runtime
3. Submit for notarization
4. Wait for Apple's approval
5. Staple the notarization ticket
6. Create a signed distribution ZIP

### Release Builds

To build and sign for all platforms:
```bash
./scripts/build-release.sh v0.1.0
```

This creates signed macOS binaries and unsigned binaries for other platforms.

## Verification

To verify a signed binary:
```bash
# Check signature
codesign --verify --verbose dist/get-repo

# Check notarization
spctl -a -t exec -vvv dist/get-repo

# Check stapled ticket
xcrun stapler validate dist/get-repo
```

## GitHub Actions

The release workflow automatically builds binaries when you push a version tag:
```bash
git tag v0.1.0
git push origin v0.1.0
```

Note: GitHub Actions can't sign macOS binaries due to security constraints. 
For signed releases, run `build-release.sh` locally and upload the signed binaries manually.

## Homebrew Distribution

Homebrew builds from source on users' machines, so signing isn't required for the Homebrew formula.
The binary is compiled locally and inherits the user's code signing identity.