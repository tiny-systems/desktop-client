//go:build !darwin && !windows

package main

import (
	"os"
	"strings"
)

func init() {
	// On Windows/Linux, the URL is passed as a command-line argument
	// when the OS launches the app via the registered URL scheme.
	for _, arg := range os.Args[1:] {
		if strings.HasPrefix(arg, "tinysystems://") {
			select {
			case pendingDeepLink <- arg:
			default:
			}
			break
		}
	}
}
