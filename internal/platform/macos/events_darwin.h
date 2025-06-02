#ifndef EVENTS_DARWIN_H
#define EVENTS_DARWIN_H

#import <Cocoa/Cocoa.h>
#import <Carbon/Carbon.h>

// Function type definition for callbacks from C to Go
typedef void (*GoCallbackFunction)();

// Register global hotkey
bool registerHotkey(int keyCode, int modifiers, GoCallbackFunction callbackFn);

// Unregister global hotkey
bool unregisterHotkey();

// Start event monitoring
void startEventMonitor();

// Stop event monitoring
void stopEventMonitor();

// Check if a key is currently pressed
bool isKeyPressed(int keyCode);

// Add a new callback for specific key+modifier combinations
void addKeyCallback(int keyCode, int modifiers, GoCallbackFunction callbackFn);

#endif // EVENTS_DARWIN_H