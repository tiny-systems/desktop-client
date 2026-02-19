package main

// pendingDeepLink receives deep link URLs from platform-specific handlers.
// Buffered to 1 so the sender never blocks (URL arrives before Wails startup).
var pendingDeepLink = make(chan string, 1)
