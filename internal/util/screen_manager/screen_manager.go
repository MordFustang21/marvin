package screenmanager

/*
#include <stdint.h> // <<< ADD THIS INCLUDE

#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework Cocoa

// These are the C function signatures as defined in window_mover.m
// The C compiler needs to know what 'uintptr_t' is from <stdint.h>
void MoveToMainScreen(uintptr_t nsWindowPtr);
void MoveToScreenWithMouse(uintptr_t nsWindowPtr);
void MoveToScreenAtIndex(uintptr_t nsWindowPtr, int screenIndex);
int GetScreenCount();
*/
import "C"

import (
	"log/slog"
)

// GoMoveToMainScreen moves the given window to the main screen.
func GoMoveToMainScreen(nsWindowPtr uintptr) {
	if nsWindowPtr == 0 {
		slog.Debug("Go: NSWindow pointer not yet available.")
		return
	}

	slog.Debug("Go: Calling MoveToMainScreen")
	C.MoveToMainScreen(C.uintptr_t(nsWindowPtr)) // Cast Go uintptr to C.uintptr_t
}

// GoMoveToScreenWithMouse moves the given window to the screen containing the mouse cursor.
func GoMoveToScreenWithMouse(nsWindowPtr uintptr) {
	if nsWindowPtr == 0 {
		slog.Debug("Go: NSWindow pointer not yet available.")
		return
	}

	slog.Debug("Go: Calling MoveToScreenWithMouse")
	C.MoveToScreenWithMouse(C.uintptr_t(nsWindowPtr)) // Cast Go uintptr to C.uintptr_t
}
