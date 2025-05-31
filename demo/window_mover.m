// window_mover.m
#import <Cocoa/Cocoa.h>
#include <stdio.h> // For NSLog alternative if needed, but NSLog is fine.
#include <stdint.h> // For uintptr_t

// Helper to center window on a target screen's visible area
static void centerWindowOnScreen(NSWindow* window, NSScreen* targetScreen) {
    if (!window || !targetScreen) {
        fprintf(stderr, "centerWindowOnScreen: nil window or screen\n");
        return;
    }

    NSRect screenVisibleFrame = [targetScreen visibleFrame]; // Area excluding menu bar and Dock
    NSRect windowFrame = [window frame];

    // Calculate new origin (bottom-left) to center the window
    CGFloat newX = screenVisibleFrame.origin.x + (screenVisibleFrame.size.width - windowFrame.size.width) / 2.0;
    CGFloat newY = screenVisibleFrame.origin.y + (screenVisibleFrame.size.height - windowFrame.size.height) / 2.0;

    // Basic clamping to ensure window is mostly on screen (adjust as needed)
    // Check left/bottom edges
    if (newX < screenVisibleFrame.origin.x) newX = screenVisibleFrame.origin.x;
    if (newY < screenVisibleFrame.origin.y) newY = screenVisibleFrame.origin.y;
    // Check top/right edges (window origin is bottom-left)
    if (newX + windowFrame.size.width > screenVisibleFrame.origin.x + screenVisibleFrame.size.width) {
        newX = screenVisibleFrame.origin.x + screenVisibleFrame.size.width - windowFrame.size.width;
    }
    if (newY + windowFrame.size.height > screenVisibleFrame.origin.y + screenVisibleFrame.size.height) {
        newY = screenVisibleFrame.origin.y + screenVisibleFrame.size.height - windowFrame.size.height;
    }
    
    fprintf(stdout, "ObjC: Moving window to X: %.0f, Y: %.0f on screen '%@'\n", newX, newY, [targetScreen localizedName]);
    [window setFrameOrigin:NSMakePoint(newX, newY)];
    // Optional: Make the window key and front
    // [window makeKeyAndOrderFront:nil];
}

// Exported C function to move window to the main screen
void MoveToMainScreen(uintptr_t nsWindowPtr) {
    @autoreleasepool {
        NSWindow* window = (NSWindow*)nsWindowPtr;
        NSScreen* mainScreen = [NSScreen mainScreen];
        if (window && mainScreen) {
            fprintf(stdout, "ObjC: Request to move to Main Screen\n");
            centerWindowOnScreen(window, mainScreen);
        } else {
            fprintf(stderr, "ObjC: MoveToMainScreen: Could not get window or main screen.\n");
        }
    }
}

// Exported C function to move window to the screen with the mouse cursor
void MoveToScreenWithMouse(uintptr_t nsWindowPtr) {
    @autoreleasepool {
        NSWindow* window = (NSWindow*)nsWindowPtr;
        if (!window) {
            fprintf(stderr, "ObjC: MoveToScreenWithMouse: nil window pointer\n");
            return;
        }

        // Mouse location is in global screen coordinates (bottom-left (0,0) on primary screen)
        NSPoint mouseLoc = [NSEvent mouseLocation];
        
        NSScreen* targetScreen = nil;
        NSArray<NSScreen*>* screens = [NSScreen screens];
        for (NSScreen* screen in screens) {
            // NSMouseInRect checks if a point is in a rect.
            // The third argument 'flipped' refers to the coordinate system of the point.
            // Since mouseLoc is in global (unflipped) coordinates, flipped should be NO.
            if (NSMouseInRect(mouseLoc, [screen frame], NO)) {
                targetScreen = screen;
                break;
            }
        }

        if (!targetScreen) {
            fprintf(stderr, "ObjC: MoveToScreenWithMouse: Could not find screen for mouse. Defaulting to main screen.\n");
            targetScreen = [NSScreen mainScreen]; // Fallback
        }
        
        if (targetScreen) {
            fprintf(stdout, "ObjC: Request to move to Screen with Mouse (found: '%@')\n", [targetScreen localizedName]);
            centerWindowOnScreen(window, targetScreen);
        } else {
            fprintf(stderr, "ObjC: MoveToScreenWithMouse: Could not determine target screen even after fallback.\n");
        }
    }
}

// Exported C function to move window to a screen by its index
void MoveToScreenAtIndex(uintptr_t nsWindowPtr, int screenIndex) {
    @autoreleasepool {
        NSWindow* window = (NSWindow*)nsWindowPtr;
        if (!window) {
            fprintf(stderr, "ObjC: MoveToScreenAtIndex: nil window pointer\n");
            return;
        }

        NSArray<NSScreen*>* screens = [NSScreen screens];
        if (screenIndex < 0 || screenIndex >= [screens count]) {
            fprintf(stderr, "ObjC: MoveToScreenAtIndex: Invalid screen index %d. Available: 0 to %lu\n", screenIndex, (unsigned long)([screens count] - 1));
            return;
        }

        NSScreen* targetScreen = [screens objectAtIndex:screenIndex];
        if (targetScreen) {
            fprintf(stdout, "ObjC: Request to move to Screen at Index %d ('%@')\n", screenIndex, [targetScreen localizedName]);
            centerWindowOnScreen(window, targetScreen);
        } else {
            // Should not happen if index is valid
            fprintf(stderr, "ObjC: MoveToScreenAtIndex: Could not get screen at index %d.\n", screenIndex);
        }
    }
}

// Helper to get the number of screens
int GetScreenCount() {
    @autoreleasepool {
        return (int)[[NSScreen screens] count];
    }
}
