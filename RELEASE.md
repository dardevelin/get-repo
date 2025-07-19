# Release Process

## Pre-release Checklist

- [ ] All tests pass: `make test`
- [ ] Code is properly formatted: `gofmt -l .`
- [ ] Version updated in `pkg/version/version.go`
- [ ] CHANGELOG updated (if exists)
- [ ] README is up to date
- [ ] Man page generated: `make man`

## Build Release

1. **Clean build directory**
   ```bash
   make clean
   ```

2. **Build for all platforms**
   ```bash
   make build-all
   ```

3. **Create release directory**
   ```bash
   mkdir -p dist/v1.0.0
   ```

4. **Copy binaries**
   ```bash
   cp get-repo-darwin-amd64 dist/v1.0.0/
   cp get-repo-darwin-arm64 dist/v1.0.0/
   cp get-repo-linux-amd64 dist/v1.0.0/
   cp get-repo-linux-arm64 dist/v1.0.0/
   ```

## macOS Signing and Notarization

1. **Sign the macOS binaries**
   ```bash
   ./scripts/sign-macos.sh dist/v1.0.0/get-repo-darwin-amd64
   ./scripts/sign-macos.sh dist/v1.0.0/get-repo-darwin-arm64
   ```

2. **Create ZIP archives for notarization**
   ```bash
   cd dist/v1.0.0
   zip get-repo-darwin-amd64.zip get-repo-darwin-amd64
   zip get-repo-darwin-arm64.zip get-repo-darwin-arm64
   cd ../..
   ```

3. **Notarize the archives**
   ```bash
   ./scripts/sign-macos-notarytool.sh dist/v1.0.0/get-repo-darwin-amd64.zip
   ./scripts/sign-macos-notarytool.sh dist/v1.0.0/get-repo-darwin-arm64.zip
   ```

4. **Staple the notarization**
   ```bash
   xcrun stapler staple dist/v1.0.0/get-repo-darwin-amd64
   xcrun stapler staple dist/v1.0.0/get-repo-darwin-arm64
   ```

## Create Release Archives

```bash
cd dist/v1.0.0

# macOS Intel
tar czf get-repo-v1.0.0-darwin-amd64.tar.gz get-repo-darwin-amd64

# macOS Apple Silicon
tar czf get-repo-v1.0.0-darwin-arm64.tar.gz get-repo-darwin-arm64

# Linux AMD64
tar czf get-repo-v1.0.0-linux-amd64.tar.gz get-repo-linux-amd64

# Linux ARM64
tar czf get-repo-v1.0.0-linux-arm64.tar.gz get-repo-linux-arm64

cd ../..
```

## Calculate SHA256 Checksums

```bash
cd dist/v1.0.0
shasum -a 256 *.tar.gz > SHA256SUMS
cd ../..
```

## GitHub Release

1. **Create git tag**
   ```bash
   git tag -a v1.0.0 -m "Release version 1.0.0"
   git push origin v1.0.0
   ```

2. **Create GitHub release**
   - Go to https://github.com/dardevelin/get-repo/releases
   - Click "Draft a new release"
   - Choose tag: v1.0.0
   - Release title: "v1.0.0 - Initial Release"
   - Upload all `.tar.gz` files and `SHA256SUMS`
   - Add release notes

## Update Homebrew Tap

1. **Calculate formula values**
   ```bash
   shasum -a 256 dist/v1.0.0/get-repo-v1.0.0-darwin-amd64.tar.gz
   shasum -a 256 dist/v1.0.0/get-repo-v1.0.0-darwin-arm64.tar.gz
   ```

2. **Update formula in homebrew-get-repo**
   - Update version
   - Update download URLs
   - Update SHA256 values

## Post-release

- [ ] Verify Homebrew installation works
- [ ] Update version in main branch to next version (e.g., 1.1.0-dev)
- [ ] Announce release