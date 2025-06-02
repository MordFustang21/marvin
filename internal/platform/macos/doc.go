// Package macos provides macOS-specific native functionality for the Marvin application.
//
// This package includes CGO bindings to macOS Cocoa and Carbon frameworks to handle
// native event integration. The main functionality includes:
//   - Global hotkey registration and handling
//   - Keyboard event monitoring
//   - Native window management
//
// The code is designed to provide a Go-friendly API that abstracts the complex
// Objective-C and C interactions needed for native macOS integration.
//
// Build tags ensure this package only compiles on macOS, with stub implementations
// provided for other platforms.
package macos