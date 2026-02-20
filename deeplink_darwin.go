//go:build darwin

package main

// No custom Apple Event handler needed — Wails v2 handles kAEGetURL
// via Mac.OnUrlOpen option and calls onDeepLinkReceived() from main.go.
