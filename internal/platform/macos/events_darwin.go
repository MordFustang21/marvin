package macos

/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework Cocoa -framework Carbon
#include "events_darwin.h"
#include <stdlib.h>

// Forward declaration of Go callback functions
extern void goHotkeyCallback();
extern void goKeyEventCallback(int keyCode, int modifiers);
*/
import "C"
import (
	"log/slog"
	"sync"
)

// Key codes for common keys (macOS virtual key codes)
const (
	KeySpace   = 49
	KeyReturn  = 36
	KeyEscape  = 53
	KeyCommand = 0x37
	KeyOption  = 0x3A
	KeyControl = 0x3B
	KeyShift   = 0x38
	KeyTab     = 48
	KeyDelete  = 51
	KeyLeft    = 123
	KeyRight   = 124
	KeyUp      = 126
	KeyDown    = 125
)

// Modifier keys
const (
	ModCommand = 1 << 8  // Command key
	ModShift   = 1 << 9  // Shift key
	ModOption  = 1 << 11 // Option key
	ModControl = 1 << 12 // Control key
)

// Singleton event handler
var (
	handler     *EventHandler
	handlerOnce sync.Once
)

// KeyCallback represents a function that will be called when a key combination is pressed
type KeyCallback func()

// EventHandler manages macOS keyboard events
type EventHandler struct {
	isMonitoring    bool
	registeredKeys  map[string]KeyCallback
	globalHotkeySet bool
	mu              sync.RWMutex
}

// GetEventHandler returns the singleton EventHandler instance
func GetEventHandler() *EventHandler {
	handlerOnce.Do(func() {
		handler = &EventHandler{
			registeredKeys: make(map[string]KeyCallback),
		}
	})
	return handler
}

// RegisterGlobalHotkey registers a system-wide hotkey
func (h *EventHandler) RegisterGlobalHotkey(keyCode int, modifiers int, callback KeyCallback) bool {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.globalHotkeySet {
		// Unregister existing hotkey first
		h.UnregisterGlobalHotkey()
	}

	// Store the callback
	globalHotkeyCallback = callback

	// Register via CGO
	success := bool(C.registerHotkey(C.int(keyCode), C.int(modifiers), (C.GoCallbackFunction)(C.goHotkeyCallback)))

	if success {
		h.globalHotkeySet = true
		slog.Debug("Registered global hotkey",
			slog.Int("keyCode", keyCode),
			slog.Int("modifiers", modifiers))
	} else {
		slog.Error("Failed to register global hotkey",
			slog.Int("keyCode", keyCode),
			slog.Int("modifiers", modifiers))
	}

	return success
}

// UnregisterGlobalHotkey removes the global hotkey
func (h *EventHandler) UnregisterGlobalHotkey() bool {
	h.mu.Lock()
	defer h.mu.Unlock()

	if !h.globalHotkeySet {
		return true
	}

	success := bool(C.unregisterHotkey())
	if success {
		h.globalHotkeySet = false
		globalHotkeyCallback = nil
		slog.Debug("Unregistered global hotkey")
	} else {
		slog.Error("Failed to unregister global hotkey")
	}

	return success
}

// StartMonitoring begins monitoring keyboard events
func (h *EventHandler) StartMonitoring() {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.isMonitoring {
		return
	}

	C.startEventMonitor()
	h.isMonitoring = true
	slog.Debug("Started keyboard event monitoring")
}

// StopMonitoring stops monitoring keyboard events
func (h *EventHandler) StopMonitoring() {
	h.mu.Lock()
	defer h.mu.Unlock()

	if !h.isMonitoring {
		return
	}

	C.stopEventMonitor()
	h.isMonitoring = false
	slog.Debug("Stopped keyboard event monitoring")
}

// RegisterKeyCombo registers a callback for a specific key combination
func (h *EventHandler) RegisterKeyCombo(keyCode int, modifiers int, callback KeyCallback) {
	h.mu.Lock()
	defer h.mu.Unlock()

	key := keyComboString(keyCode, modifiers)
	h.registeredKeys[key] = callback

	// Make sure we're monitoring events
	if !h.isMonitoring {
		C.startEventMonitor()
		h.isMonitoring = true
	}

	// Register the callback with C code
	C.addKeyCallback(C.int(keyCode), C.int(modifiers), (C.GoCallbackFunction)(C.goKeyEventCallback))

	slog.Debug("Registered key combination",
		slog.Int("keyCode", keyCode),
		slog.Int("modifiers", modifiers))
}

// UnregisterKeyCombo removes a callback for a specific key combination
func (h *EventHandler) UnregisterKeyCombo(keyCode int, modifiers int) {
	h.mu.Lock()
	defer h.mu.Unlock()

	key := keyComboString(keyCode, modifiers)
	delete(h.registeredKeys, key)
}

// IsKeyPressed checks if a key is currently pressed
func (h *EventHandler) IsKeyPressed(keyCode int) bool {
	return bool(C.isKeyPressed(C.int(keyCode)))
}

// Helper function to create a string key for the key combination
func keyComboString(keyCode int, modifiers int) string {
	return string(rune(keyCode)) + "-" + string(rune(modifiers))
}

// Store the global hotkey callback
var (
	globalHotkeyCallback KeyCallback
	keyCallbacks         = make(map[string]KeyCallback)
	callbacksMutex       sync.RWMutex
)

//export goHotkeyCallback
func goHotkeyCallback() {
	if globalHotkeyCallback != nil {
		globalHotkeyCallback()
	}
}

//export goKeyEventCallback
func goKeyEventCallback(keyCode C.int, modifiers C.int) {
	key := keyComboString(int(keyCode), int(modifiers))

	callbacksMutex.RLock()
	callback, exists := keyCallbacks[key]
	callbacksMutex.RUnlock()

	if exists && callback != nil {
		callback()
	}
}

// Cleanup resources when the package is unloaded
func cleanup() {
	if handler != nil && handler.isMonitoring {
		handler.StopMonitoring()
	}

	if handler != nil && handler.globalHotkeySet {
		handler.UnregisterGlobalHotkey()
	}
}
