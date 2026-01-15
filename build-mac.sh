#!/bin/bash

# Build script for macOS with proper entitlements

set -e

echo "Building Wails app..."
wails build -clean

echo "Signing app with entitlements..."
codesign --force --deep --sign - --entitlements build/darwin/entitlements.plist build/bin/tinysystems.app

echo "Verifying signature..."
codesign -d --entitlements - build/bin/tinysystems.app

echo ""
echo "Build complete! App is at: build/bin/tinysystems.app"
echo ""
echo "To run the app:"
echo "  open build/bin/tinysystems.app"
echo ""
echo "If macOS still blocks it, run:"
echo "  xattr -cr build/bin/tinysystems.app"
echo "  open build/bin/tinysystems.app"
