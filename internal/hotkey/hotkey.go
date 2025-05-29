package hotkey

import (
	"fmt"
	"runtime"
	"strings"
)

// Handler defines the interface for hotkey handling
type Handler interface {
	Register(shortcut string, callback func()) error
	Unregister(shortcut string) error
	Start() error
	Stop()
}

// MacOSHandler implements hotkey handling for macOS using an applescript approach
type MacOSHandler struct {
	shortcuts map[string]func()
	running   bool
	stopChan  chan struct{}
}

// NewHandler creates a platform-specific hotkey handler
func NewHandler() (Handler, error) {
	switch runtime.GOOS {
	case "darwin":
		return &MacOSHandler{
			shortcuts: make(map[string]func()),
			stopChan:  make(chan struct{}),
		}, nil
	default:
		return nil, fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}
}

// Register registers a global hotkey with the given shortcut and callback
func (h *MacOSHandler) Register(shortcut string, callback func()) error {
	// Normalize the shortcut string
	shortcut = normalizeShortcut(shortcut)
	
	// Store the callback
	h.shortcuts[shortcut] = callback
	
	return nil
}

// Unregister removes a registered hotkey
func (h *MacOSHandler) Unregister(shortcut string) error {
	shortcut = normalizeShortcut(shortcut)
	
	// Check if the shortcut is registered
	if _, exists := h.shortcuts[shortcut]; !exists {
		return fmt.Errorf("shortcut not registered: %s", shortcut)
	}
	
	// Remove the shortcut
	delete(h.shortcuts, shortcut)
	
	return nil
}

// Start starts listening for hotkeys
func (h *MacOSHandler) Start() error {
	if h.running {
		return fmt.Errorf("hotkey handler already running")
	}
	
	h.running = true
	
	// Poll for key combinations in a goroutine
	go func() {
		for {
			select {
			case <-h.stopChan:
				return
			default:
				// Simple polling approach for demo purposes
				// In a real application, you would use a more efficient method
				for shortcut, callback := range h.shortcuts {
					if isShortcutPressed(shortcut) {
						callback()
					}
				}
			}
		}
	}()
	
	return nil
}

// Stop stops the hotkey handler
func (h *MacOSHandler) Stop() {
	if h.running {
		h.running = false
		h.stopChan <- struct{}{}
	}
}

// normalizeShortcut normalizes a shortcut string for consistent representation
func normalizeShortcut(shortcut string) string {
	return strings.ToLower(shortcut)
}

// isShortcutPressed checks if a shortcut is currently pressed
// Note: This is just a placeholder implementation
// In a real application, you would use platform-specific APIs
func isShortcutPressed(shortcut string) bool {
	// This would need to be implemented using platform-specific APIs
	// For demonstration purposes, we'll always return false
	return false
}

// For an actual implementation, you would likely need to use CGEventTap on macOS
// or similar platform-specific APIs to capture global keyboard events