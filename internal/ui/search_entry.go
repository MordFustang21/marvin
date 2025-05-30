package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
)

// SearchEntry is a custom Entry widget that allows handling special keys (like navigation and escape)
// while still supporting normal text input.
type SearchEntry struct {
	widget.Entry
	OnSpecialKey func(key *fyne.KeyEvent)
}

// NewSearchEntry creates a new SearchEntry widget.
func NewSearchEntry() *SearchEntry {
	entry := &SearchEntry{}
	entry.ExtendBaseWidget(entry)
	return entry
}

// TypedKey intercepts key events and calls OnSpecialKey for navigation/escape/return keys.
func (e *SearchEntry) TypedKey(key *fyne.KeyEvent) {
	switch key.Name {
	case fyne.KeyEscape, fyne.KeyDown, fyne.KeyUp, fyne.KeyReturn:
		if e.OnSpecialKey != nil {
			e.OnSpecialKey(key)
			return // Don't pass to default handler
		}
	}
	e.Entry.TypedKey(key) // Default behavior for other keys
}

func (e *SearchEntry) SelectAll() {
	// This is to programatically select all text in the entry.
}
