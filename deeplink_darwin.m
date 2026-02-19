#import <Cocoa/Cocoa.h>

// Forward declaration of Go callback
extern void goHandleDeepLinkURL(char* url);

@interface TSDeepLinkHandler : NSObject
@end

@implementation TSDeepLinkHandler
- (void)handleGetURLEvent:(NSAppleEventDescriptor *)event
           withReplyEvent:(NSAppleEventDescriptor *)reply {
    NSString *urlString = [[event paramDescriptorForKeyword:keyDirectObject] stringValue];
    if (urlString) {
        goHandleDeepLinkURL((char *)[urlString UTF8String]);
    }
}
@end

static TSDeepLinkHandler *_handler = nil;

void RegisterURLHandler(void) {
    if (_handler != nil) return;
    _handler = [[TSDeepLinkHandler alloc] init];
    [[NSAppleEventManager sharedAppleEventManager]
        setEventHandler:_handler
            andSelector:@selector(handleGetURLEvent:withReplyEvent:)
          forEventClass:kInternetEventClass
             andEventID:kAEGetURL];
}
