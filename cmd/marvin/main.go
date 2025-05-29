package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"fyne.io/fyne/v2/app"
	"github.com/MordFustang21/marvin-go/internal/hotkey"
	"github.com/MordFustang21/marvin-go/internal/search"
	"github.com/MordFustang21/marvin-go/internal/search/providers/calculator"
	"github.com/MordFustang21/marvin-go/internal/search/providers/spotlight"
	"github.com/MordFustang21/marvin-go/internal/theme"
	"github.com/MordFustang21/marvin-go/internal/ui"
)

func main() {
	// Create a new Fyne application with custom ID
	marvin := app.NewWithID("com.mordfustang.marvin")

	// Apply our custom GitHub Dark theme
	marvin.Settings().SetTheme(theme.NewGitHubDarkTheme())

	// Set custom flags for the app window
	// This tells the windowing system to treat this as a special UI element
	marvin.Lifecycle().SetOnStarted(func() {
		setupAppAttributes()
	})

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

	// Show the window by default since our hotkey isn't working yet
	searchWindow.ShowWithKeyboardFocus()

	// Set up hotkey handling
	hotkeyHandler, err := hotkey.NewHandler()
	if err != nil {
		log.Fatalf("Failed to create hotkey handler: %v", err)
	}

	// Register Cmd+Space (or Alt+Space) as the activation hotkey
	// Note: In a real implementation, this would need proper platform detection
	err = hotkeyHandler.Register("cmd+space", func() {
		// Toggle window visibility when hotkey is pressed
		if searchWindow.IsVisible() {
			searchWindow.Hide()
		} else {
			searchWindow.ShowWithKeyboardFocus()
		}
	})

	if err != nil {
		log.Printf("Warning: Failed to register hotkey: %v", err)
		// Still show the window if we couldn't register the hotkey
		searchWindow.ShowWithKeyboardFocus()
	} else {
		// Start the hotkey handler
		if err := hotkeyHandler.Start(); err != nil {
			log.Printf("Warning: Failed to start hotkey handler: %v", err)
		}
	}

	// Set up a signal handler for graceful shutdown
	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-signalCh
		hotkeyHandler.Stop()
		searchWindow.Close()
		os.Exit(0)
	}()

	// Run the application
	marvin.Run()
}

// setupAppAttributes sets platform-specific window attributes
// This is a placeholder for platform-specific code to hide the title bar
func setupAppAttributes() {
	// On macOS, this would use Objective-C or CGo to set window level and style
	// This is just a placeholder - actual implementation would require platform-specific code
	// For a real implementation, you'd use CGo to call native APIs
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
