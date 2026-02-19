//go:build windows

package main

import (
	"fmt"
	"os"
	"strings"

	"golang.org/x/sys/windows/registry"
)

func init() {
	// Parse CLI arg (Windows passes URL as argument)
	for _, arg := range os.Args[1:] {
		if strings.HasPrefix(arg, "tinysystems://") {
			select {
			case pendingDeepLink <- arg:
			default:
			}
			break
		}
	}

	// Register URL protocol in the Windows registry (per-user, no admin needed)
	registerWindowsURLProtocol()
}

func registerWindowsURLProtocol() {
	exe, err := os.Executable()
	if err != nil {
		return
	}

	// Check if already registered with the correct path
	k, err := registry.OpenKey(registry.CURRENT_USER, `Software\Classes\tinysystems\shell\open\command`, registry.QUERY_VALUE)
	if err == nil {
		val, _, _ := k.GetStringValue("")
		k.Close()
		if strings.Contains(val, exe) {
			return // Already registered with correct path
		}
	}

	// Create tinysystems:// protocol key
	k, _, err = registry.CreateKey(registry.CURRENT_USER, `Software\Classes\tinysystems`, registry.SET_VALUE)
	if err != nil {
		return
	}
	k.SetStringValue("", "URL:TinySystems Protocol")
	k.SetStringValue("URL Protocol", "")
	k.Close()

	// Set command
	k, _, err = registry.CreateKey(registry.CURRENT_USER, `Software\Classes\tinysystems\shell\open\command`, registry.SET_VALUE)
	if err != nil {
		return
	}
	k.SetStringValue("", fmt.Sprintf(`"%s" "%%1"`, exe))
	k.Close()
}
