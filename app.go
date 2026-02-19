package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/user"
	"path/filepath"
	"runtime/debug"
	"sync"
	"time"

	"github.com/go-logr/logr"
	api "github.com/tiny-systems/platform-api"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

var (
	BuildTime = "dev"
	Version   = "unknown"
)

const sdkModulePath = "github.com/tiny-systems/module"

type BuildInfo struct {
	BuildTime  string `json:"buildTime"`
	Version    string `json:"version"`
	SdkVersion string `json:"sdkVersion"`
}

func (a *App) GetBuildInfo() BuildInfo {
	sdkVersion := "unknown"
	if info, ok := debug.ReadBuildInfo(); ok {
		for _, dep := range info.Deps {
			if dep.Path == sdkModulePath {
				sdkVersion = dep.Version
				break
			}
		}
	}
	return BuildInfo{
		BuildTime:  BuildTime,
		Version:    Version,
		SdkVersion: sdkVersion,
	}
}

// App struct
type App struct {
	ctx    context.Context
	logger logr.Logger

	// watchMu protects watchCancel
	watchMu     sync.Mutex
	watchCancel context.CancelFunc
}

// Preferences stores user preferences
type Preferences struct {
	LastContext   string `json:"lastContext"`
	LastNamespace string `json:"lastNamespace"`
}

// getPreferencesPath returns the path to the preferences file
func getPreferencesPath() (string, error) {
	usr, err := user.Current()
	if err != nil {
		return "", err
	}

	configDir := filepath.Join(usr.HomeDir, ".config", "tinysystems")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return "", err
	}

	return filepath.Join(configDir, "preferences.json"), nil
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

	// Enable direct deep link event emission for URLs arriving while app is running
	deepLinkStartup(a.ctx)
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

// GetPreferences returns saved user preferences
func (a *App) GetPreferences() (*Preferences, error) {
	path, err := getPreferencesPath()
	if err != nil {
		return &Preferences{}, nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		// File doesn't exist yet - return empty preferences
		return &Preferences{}, nil
	}

	var prefs Preferences
	if err := json.Unmarshal(data, &prefs); err != nil {
		return &Preferences{}, nil
	}

	return &prefs, nil
}

// SavePreferences saves user preferences
func (a *App) SavePreferences(contextName, namespace string) error {
	path, err := getPreferencesPath()
	if err != nil {
		return err
	}

	prefs := Preferences{
		LastContext:   contextName,
		LastNamespace: namespace,
	}

	data, err := json.MarshalIndent(prefs, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

// SaveFile opens a save dialog and writes content to the selected file
func (a *App) SaveFile(defaultFilename, content string) (string, error) {
	filepath, err := runtime.SaveFileDialog(a.ctx, runtime.SaveDialogOptions{
		DefaultFilename: defaultFilename,
		Filters: []runtime.FileFilter{
			{DisplayName: "JSON Files", Pattern: "*.json"},
			{DisplayName: "All Files", Pattern: "*"},
		},
	})
	if err != nil {
		return "", err
	}
	if filepath == "" {
		// User cancelled
		return "", nil
	}

	if err := os.WriteFile(filepath, []byte(content), 0644); err != nil {
		return "", err
	}

	return filepath, nil
}

// OpenFile opens a file dialog and returns the file content
func (a *App) OpenFile() (string, error) {
	filepath, err := runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Filters: []runtime.FileFilter{
			{DisplayName: "JSON Files", Pattern: "*.json"},
			{DisplayName: "All Files", Pattern: "*"},
		},
	})
	if err != nil {
		return "", err
	}
	if filepath == "" {
		// User cancelled
		return "", nil
	}

	data, err := os.ReadFile(filepath)
	if err != nil {
		return "", err
	}

	return string(data), nil
}

// GetPendingDeepLink returns a deep link URL that arrived before the frontend was ready.
// Called by the frontend on mount to catch URLs from cold-start launches.
func (a *App) GetPendingDeepLink() string {
	deepLinkState.mu.Lock()
	defer deepLinkState.mu.Unlock()
	url := deepLinkState.pendingURL
	deepLinkState.pendingURL = "" // consume it
	fmt.Println("[DEEPLINK] GetPendingDeepLink called, returning:", url)
	return url
}

// FetchSolutionJSON downloads solution JSON from the given URL (legacy deep links).
func (a *App) FetchSolutionJSON(url string) (string, error) {
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return "", fmt.Errorf("failed to fetch solution: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("server returned %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	if !json.Valid(body) {
		return "", fmt.Errorf("response is not valid JSON")
	}

	return string(body), nil
}

// FetchSolutionExport downloads solution export JSON using a one-time token via the platform-api client.
func (a *App) FetchSolutionExport(token, apiBase string) (string, error) {
	client, err := api.NewClientWithResponses(apiBase)
	if err != nil {
		return "", fmt.Errorf("failed to create API client: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	resp, err := client.ExportSolutionWithResponse(ctx, &api.ExportSolutionParams{Token: token})
	if err != nil {
		return "", fmt.Errorf("failed to fetch solution: %w", err)
	}

	if resp.StatusCode() != http.StatusOK {
		return "", fmt.Errorf("server returned %d", resp.StatusCode())
	}

	body := resp.Body
	if !json.Valid(body) {
		return "", fmt.Errorf("response is not valid JSON")
	}

	return string(body), nil
}
