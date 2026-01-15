package main

import (
	"context"
	"os"
	"os/user"
	"path/filepath"
	"sync"

	"github.com/go-logr/logr"
)

// App struct
type App struct {
	ctx    context.Context
	logger logr.Logger

	// watchMu protects watchCancel
	watchMu     sync.Mutex
	watchCancel context.CancelFunc
}

// NewApp creates a new App application struct
func NewApp(l logr.Logger) *App {
	return &App{
		logger: l,
	}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx

	// Set PATH to include common locations for gcloud and other CLI tools
	// This is necessary because GUI apps on macOS don't inherit shell PATH
	if err := setupPATH(); err != nil {
		a.logger.Error(err, "Failed to setup PATH")
	}
}

// setupPATH adds common CLI tool locations to PATH environment variable
func setupPATH() error {
	currentPath := os.Getenv("PATH")

	// Get user home directory
	usr, err := user.Current()
	if err != nil {
		return err
	}

	// Common paths where gke-gcloud-auth-plugin and other tools might be located
	additionalPaths := []string{
		filepath.Join(usr.HomeDir, "google-cloud-sdk", "bin"),
		filepath.Join(usr.HomeDir, ".local", "bin"),
		filepath.Join(usr.HomeDir, "go", "bin"),
		filepath.Join(usr.HomeDir, ".krew", "bin"),
		"/usr/local/bin",
		"/opt/homebrew/bin",
		"/usr/local/go/bin",
	}

	// Build new PATH with additional paths prepended
	newPath := currentPath
	for _, p := range additionalPaths {
		// Only add if directory exists
		if _, err := os.Stat(p); err == nil {
			newPath = p + ":" + newPath
		}
	}

	// Set the updated PATH
	return os.Setenv("PATH", newPath)
}

func (a *App) shutdown(ctx context.Context) {
}
