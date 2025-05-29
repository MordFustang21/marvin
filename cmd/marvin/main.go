package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"fyne.io/fyne/v2/app"
	"github.com/MordFustang21/marvin-go/internal/search"
	"github.com/MordFustang21/marvin-go/internal/search/providers/calculator"
	"github.com/MordFustang21/marvin-go/internal/search/providers/spotlight"
	"github.com/MordFustang21/marvin-go/internal/theme"
	"github.com/MordFustang21/marvin-go/internal/ui"
	hook "github.com/robotn/gohook"
)

func main() {
	// Create a new Fyne application with custom ID
	marvin := app.NewWithID("com.mordfustang.marvin")

	// Apply our custom GitHub Dark theme
	marvin.Settings().SetTheme(theme.NewGitHubDarkTheme())

	// Initialize search providers
	registry := search.NewRegistry()
	setupSearchProviders(registry)

	// Create our search window
	searchWindow := ui.NewSearchWindow(marvin, registry)

	// Hide decorations as much as possible
	w := searchWindow.GetWindow()
	if w != nil {
		w.SetPadded(false)
		w.SetTitle("")
	}

	// TODO: Setup shortcuts for cmd+space to toggle the window.
	go func() {
		hook.Register(hook.KeyDown, []string{"cmd", "space"}, func(e hook.Event) {
			fmt.Println("Toggling window")
			// Toggle the search window visibility
			if searchWindow.IsVisible() {
				searchWindow.Hide()
			} else {
				searchWindow.ShowWithKeyboardFocus()
			}
		})

		hook.Process(hook.Start())
	}()

	// Show the window by default since our hotkey isn't working yet
	searchWindow.ShowWithKeyboardFocus()

	// Set up a signal handler for graceful shutdown
	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-signalCh
		searchWindow.Close()
		os.Exit(0)
	}()

	// Run the application
	marvin.Run()
}

// setupSearchProviders registers all search providers with the registry
func setupSearchProviders(registry *search.Registry) {
	// Register calculator provider with higher priority (lower number)
	calculatorProvider := calculator.NewProvider(10)
	registry.RegisterProvider(calculatorProvider)

	// Register spotlight provider with lower priority
	spotlightProvider := spotlight.NewProvider(100, 20) // Priority 100, max 20 results
	registry.RegisterProvider(spotlightProvider)
}
