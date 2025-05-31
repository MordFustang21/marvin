package util

import "fyne.io/fyne/v2"

// ShortcutLauncher describes a shortcut selectAll action.
type ShortcutLauncher struct{}

var _ fyne.KeyboardShortcut = (*ShortcutLauncher)(nil)

// Key returns the [KeyName] for this shortcut.
//
// Implements: [KeyboardShortcut]
func (se *ShortcutLauncher) Key() fyne.KeyName {
	return fyne.KeySpace
}

// Mod returns the [KeyModifier] for this shortcut.
//
// Implements: [KeyboardShortcut]
func (se *ShortcutLauncher) Mod() fyne.KeyModifier {
	return fyne.KeyModifierShortcutDefault
}

// ShortcutName returns the shortcut name
func (se *ShortcutLauncher) ShortcutName() string {
	return "Launcher"
}
