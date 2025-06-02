#import "events_darwin.h"
#include <stdlib.h>

// Global variables to keep track of hotkeys and event monitors
static id eventMonitor = nil;
static CFMachPortRef eventTap = NULL;
static EventHotKeyRef hotkeyRef = NULL;
static EventHotKeyID hotkeyID;

// Use the function type defined in the header
static GoCallbackFunction registeredCallback = NULL;

// Map to store key combinations and their callbacks
static NSMutableDictionary *keyCallbacks = nil;

// Initialize the callbacks dictionary
__attribute__((constructor)) static void initialize() {
    keyCallbacks = [[NSMutableDictionary alloc] init];
}

// Convert Carbon modifier flags to Cocoa modifier flags
static NSEventModifierFlags carbonToCocoaModifiers(int carbonModifiers) {
    NSEventModifierFlags cocoaModifiers = 0;
    
    if (carbonModifiers & cmdKey) cocoaModifiers |= NSEventModifierFlagCommand;
    if (carbonModifiers & optionKey) cocoaModifiers |= NSEventModifierFlagOption;
    if (carbonModifiers & controlKey) cocoaModifiers |= NSEventModifierFlagControl;
    if (carbonModifiers & shiftKey) cocoaModifiers |= NSEventModifierFlagShift;
    
    return cocoaModifiers;
}

// Main event handler called when the hotkey is triggered
static OSStatus hotkeyHandler(EventHandlerCallRef nextHandler, EventRef event, void *userData) {
    // Call the Go callback function if registered
    if (registeredCallback != NULL) {
        registeredCallback();
    }
    return noErr;
}

// Global hotkey registration function
bool registerHotkey(int keyCode, int modifiers, GoCallbackFunction callbackFn) {
    // Store the Go callback function
    registeredCallback = callbackFn;
    
    // Unregister any existing hotkey
    unregisterHotkey();
    
    // Create an event type spec for the hotkey
    EventTypeSpec eventType;
    eventType.eventClass = kEventClassKeyboard;
    eventType.eventKind = kEventHotKeyPressed;
    
    // Install the event handler
    InstallEventHandler(GetApplicationEventTarget(),
                        NewEventHandlerUPP(hotkeyHandler),
                        1, &eventType, NULL, NULL);
    
    // Register the hotkey
    hotkeyID.signature = 'MRVN'; // Custom signature for Marvin app
    hotkeyID.id = 1; // Unique ID for this hotkey
    
    OSStatus status = RegisterEventHotKey(keyCode,
                                         modifiers,
                                         hotkeyID,
                                         GetApplicationEventTarget(),
                                         0,
                                         &hotkeyRef);
    
    return status == noErr;
}

// Unregister the global hotkey
bool unregisterHotkey() {
    if (hotkeyRef != NULL) {
        OSStatus status = UnregisterEventHotKey(hotkeyRef);
        hotkeyRef = NULL;
        return status == noErr;
    }
    return true;
}

// CGEventCallback for the event tap
static CGEventRef eventTapCallback(CGEventTapProxy proxy, CGEventType type, CGEventRef event, void *userInfo) {
    if (type != kCGEventKeyDown && type != kCGEventFlagsChanged) {
        return event;
    }
    
    // Convert CGEvent to NSEvent
    NSEvent *nsEvent = [NSEvent eventWithCGEvent:event];
    
    // Check if this key combination has a registered callback
    NSString *key = [NSString stringWithFormat:@"%ld-%lu", (long)nsEvent.keyCode, (unsigned long)nsEvent.modifierFlags];
    id callbackObj = [keyCallbacks objectForKey:key];
    GoCallbackFunction callback = (__bridge GoCallbackFunction)([callbackObj isKindOfClass:[NSValue class]] ? [(NSValue *)callbackObj pointerValue] : NULL);
    
    if (callback != NULL) {
        callback();
    }
    
    return event;
}

// Start monitoring keyboard events
void startEventMonitor() {
    if (eventTap != NULL) {
        // Event tap already active
        return;
    }
    
    // Create event tap to monitor all keyboard events
    eventTap = CGEventTapCreate(kCGSessionEventTap,
                                kCGHeadInsertEventTap,
                                kCGEventTapOptionDefault,
                                CGEventMaskBit(kCGEventKeyDown) | CGEventMaskBit(kCGEventFlagsChanged),
                                eventTapCallback,
                                NULL);
    
    if (eventTap != NULL) {
        // Create a run loop source
        CFRunLoopSourceRef runLoopSource = CFMachPortCreateRunLoopSource(kCFAllocatorDefault, eventTap, 0);
        
        // Add to the current run loop
        CFRunLoopAddSource(CFRunLoopGetCurrent(), runLoopSource, kCFRunLoopCommonModes);
        
        // Enable the event tap
        CGEventTapEnable(eventTap, true);
        
        CFRelease(runLoopSource);
    }
}

// Stop monitoring keyboard events
void stopEventMonitor() {
    if (eventTap != NULL) {
        CGEventTapEnable(eventTap, false);
        CFRelease(eventTap);
        eventTap = NULL;
    }
}

// Check if a key is currently pressed
bool isKeyPressed(int keyCode) {
    // Get the key state from the IOKit HID Manager
    return CGEventSourceKeyState(kCGEventSourceStateCombinedSessionState, keyCode);
}

// Add a callback for a specific key combination
void addKeyCallback(int keyCode, int modifiers, GoCallbackFunction callbackFn) {
    NSEventModifierFlags cocoaModifiers = carbonToCocoaModifiers(modifiers);
    NSString *key = [NSString stringWithFormat:@"%d-%lu", keyCode, (unsigned long)cocoaModifiers];
    [keyCallbacks setObject:[NSValue valueWithPointer:(void *)callbackFn] forKey:key];
}