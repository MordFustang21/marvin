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
	"github.com/MordFustang21/marvin-go/internal/search/providers/commands"
	"github.com/MordFustang21/marvin-go/internal/search/providers/spotlight"
	"github.com/MordFustang21/marvin-go/internal/search/providers/web"
	"github.com/MordFustang21/marvin-go/internal/theme"
	"github.com/MordFustang21/marvin-go/internal/ui"
	"github.com/MordFustang21/marvin-go/internal/ui/assets"
	hook "github.com/robotn/gohook"
)

func main() {
	// Create a new Fyne application with custom ID
	marvin := app.NewWithID("com.mordfustang.marvin")
	marvin.SetIcon(assets.MarvinIcon)

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
	searchWindow := ui.NewSearchWindow(marvin, registry)

	// Setup shortcuts for cmd+space to toggle the window.
	go func() {
		hook.Register(hook.KeyDown, []string{"cmd", "space"}, func(e hook.Event) {
			fyne.Do(func() {
				if searchWindow.IsVisible() {
					searchWindow.Hide()
				} else {
					searchWindow.ShowWithKeyboardFocus()
				}
			})
		})

		hook.Process(hook.Start())
	}()

	// Do not show a window by default on startup

	// Set up a signal handler for graceful shutdown
	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-signalCh
		os.Exit(0)
	}()

	// Run the application
	marvin.Run()
}

// setupSearchProviders registers all search providers with the registry
func setupSearchProviders(registry *search.Registry) {
	// Register spotlight provider with highest priority (lowest number)
	spotlightProvider := spotlight.NewProvider(1, 20) // Priority 1, max 20 results
	registry.RegisterProvider(spotlightProvider)

	// Register calculator provider with medium priority
	calculatorProvider := calculator.NewProvider(2)
	registry.RegisterProvider(calculatorProvider)

	// Register custom commands provider with medium-high priority
	commandsProvider := commands.NewProvider(3, "")
	registry.RegisterProvider(commandsProvider)

	// Register web provider with lowest priority
	webProvider := web.NewProvider(10) // Much lower priority than other providers
	registry.RegisterProvider(webProvider)
}
