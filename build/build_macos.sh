#!/bin/bash
set -e

# Define variables
APP_NAME="Marvin"
BUNDLE_NAME="${APP_NAME}.app"
VERSION="1.0.0"
BUILD_DIR="$(pwd)/build"
BUNDLE_DIR="${BUILD_DIR}/${BUNDLE_NAME}"
CONTENTS_DIR="${BUNDLE_DIR}/Contents"
MACOS_DIR="${CONTENTS_DIR}/MacOS"
RESOURCES_DIR="${CONTENTS_DIR}/Resources"
ICON_NAME="AppIcon.icns"

# Clean previous build if it exists
if [ -d "${BUNDLE_DIR}" ]; then
  echo "Removing previous build..."
  rm -rf "${BUNDLE_DIR}"
fi

# Create app bundle directory structure
echo "Creating app bundle structure..."
mkdir -p "${MACOS_DIR}"
mkdir -p "${RESOURCES_DIR}"

# Build the binary
echo "Building Marvin binary..."
cd "$(pwd)"
go build -o "${MACOS_DIR}/${APP_NAME}" ./cmd/marvin

# Copy resources
echo "Copying resources..."

# Create Info.plist
echo "Creating Info.plist..."
cat > "${CONTENTS_DIR}/Info.plist" << EOF
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>CFBundleDevelopmentRegion</key>
    <string>en</string>
    <key>CFBundleExecutable</key>
    <string>${APP_NAME}</string>
    <key>CFBundleIconFile</key>
    <string>${ICON_NAME}</string>
    <key>CFBundleIdentifier</key>
    <string>com.mordfustang.marvin</string>
    <key>CFBundleInfoDictionaryVersion</key>
    <string>6.0</string>
    <key>CFBundleName</key>
    <string>${APP_NAME}</string>
    <key>CFBundlePackageType</key>
    <string>APPL</string>
    <key>CFBundleShortVersionString</key>
    <string>${VERSION}</string>
    <key>CFBundleVersion</key>
    <string>${VERSION}</string>
    <key>LSMinimumSystemVersion</key>
    <string>10.13</string>
    <key>LSUIElement</key>
    <true/>
    <key>NSHighResolutionCapable</key>
    <true/>
    <key>NSHumanReadableCopyright</key>
    <string>Copyright © 2024 MordFustang21. All rights reserved.</string>
</dict>
</plist>
EOF

# Check if we have an icon file and copy it
if [ -f "internal/assets/${ICON_NAME}" ]; then
  cp "internal/assets/${ICON_NAME}" "${RESOURCES_DIR}/"
  echo "Icon copied to bundle."
else
  echo "Warning: Icon file internal/assets/${ICON_NAME} not found. App will use default icon."
  # Optionally create a simple placeholder icon
fi

# Create a simple README in the bundle for troubleshooting
cat > "${CONTENTS_DIR}/README.txt" << EOF
Marvin - Spotlight Alternative
Version: ${VERSION}

If you encounter any issues, please report them at:
https://github.com/MordFustang21/marvin-go/issues
EOF

# Make the binary executable
chmod +x "${MACOS_DIR}/${APP_NAME}"

# Create a DMG (optional)
# hdiutil create -volname "${APP_NAME}" -srcfolder "${BUNDLE_DIR}" -ov -format UDZO "${BUILD_DIR}/${APP_NAME}-${VERSION}.dmg"

echo "App bundle created at: ${BUNDLE_DIR}"
echo "To install, copy ${BUNDLE_NAME} to your Applications folder."
echo ""
echo "Note: If you receive a security warning when launching, go to System Preferences > Security & Privacy and click 'Open Anyway'."
