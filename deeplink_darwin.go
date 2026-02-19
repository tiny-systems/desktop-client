//go:build darwin

package main

// #cgo CFLAGS: -x objective-c
// #cgo LDFLAGS: -framework Cocoa
// extern void RegisterURLHandler(void);
import "C"

//export goHandleDeepLinkURL
func goHandleDeepLinkURL(curl *C.char) {
	url := C.GoString(curl)
	select {
	case pendingDeepLink <- url:
	default:
		// Channel full — drop old, send new
		select {
		case <-pendingDeepLink:
		default:
		}
		pendingDeepLink <- url
	}
}

func init() {
	C.RegisterURLHandler()
}
