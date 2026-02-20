package main

import (
	"context"
	"fmt"
	"sync"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// deepLinkState holds the deep link URL and app context.
// URLs arriving before startup are stored in pendingURL.
// URLs arriving after startup are emitted directly as Wails events.
var deepLinkState struct {
	mu         sync.Mutex
	pendingURL string
	appCtx     context.Context // set once in startup()
}

// onDeepLinkReceived is called by platform-specific handlers when a URL arrives.
func onDeepLinkReceived(url string) {
	deepLinkState.mu.Lock()
	defer deepLinkState.mu.Unlock()

	fmt.Println("[DEEPLINK] onDeepLinkReceived called, url:", url, "appCtx is nil:", deepLinkState.appCtx == nil)

	if deepLinkState.appCtx != nil {
		// App is running — emit directly to frontend
		fmt.Println("[DEEPLINK] emitting deeplink:deploy event to frontend")
		runtime.EventsEmit(deepLinkState.appCtx, "deeplink:deploy", url)
	} else {
		// App not started yet — store for GetPendingDeepLink()
		fmt.Println("[DEEPLINK] storing as pendingURL (app not started yet)")
		deepLinkState.pendingURL = url
	}
}

// deepLinkStartup should be called from App.startup() to enable direct event emission.
func deepLinkStartup(ctx context.Context) {
	deepLinkState.mu.Lock()
	fmt.Println("[DEEPLINK] deepLinkStartup called, pendingURL:", deepLinkState.pendingURL)
	deepLinkState.appCtx = ctx
	deepLinkState.mu.Unlock()
}
