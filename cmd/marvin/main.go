package main

import (
	"os"
	"os/signal"
	"syscall"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/driver/desktop"
	"github.com/MordFustang21/marvin-go/internal/search"
	"github.com/MordFustang21/marvin-go/internal/search/providers/calculator"
	"github.com/MordFustang21/marvin-go/internal/search/providers/spotlight"
	"github.com/MordFustang21/marvin-go/internal/search/providers/web"
	"github.com/MordFustang21/marvin-go/internal/theme"
	"github.com/MordFustang21/marvin-go/internal/ui"
	hook "github.com/robotn/gohook"
)

func main() {
	// Create a new Fyne application with custom ID
	marvin := app.NewWithID("com.mordfustang.marvin")
	// Set system tray so the app can run in the background.
	if desk, ok := marvin.(desktop.App); ok {
		m := fyne.NewMenu("marvin")
		desk.SetSystemTrayMenu(m)
	}

	// Apply our custom GitHub Dark theme
	marvin.Settings().SetTheme(theme.NewGitHubDarkTheme())

	// Initialize search providers
	registry := search.NewRegistry()
	setupSearchProviders(registry)

	// Keep track of the last search window so we can close it if needed
	var lastWindow *ui.SearchWindow

	// Setup shortcuts for cmd+space to toggle the window.
	go func() {
		hook.Register(hook.KeyDown, []string{"cmd", "space"}, func(e hook.Event) {
			// Close previous window if it's still open
			if lastWindow != nil && lastWindow.IsVisible() {
				lastWindow.Close()
				lastWindow = nil
				return
			}

			// Create and show a new window
			searchWindow := ui.NewSearchWindow(marvin, registry)
			lastWindow = searchWindow

			// Hide decorations as much as possible
			w := searchWindow.GetWindow()
			if w != nil {
				w.SetPadded(false)
				w.SetTitle("")
			}

			searchWindow.ShowWithKeyboardFocus()
		})

		hook.Process(hook.Start())
	}()

	// Do not show a window by default on startup

	// Set up a signal handler for graceful shutdown
	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-signalCh
		if lastWindow != nil {
			lastWindow.Close()
		}
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

	webProvider := web.NewProvider(1)
	registry.RegisterProvider(webProvider)
}
