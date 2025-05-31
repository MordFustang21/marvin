package main

import (
	"log/slog"
	"os"
	"os/signal"
	"runtime"
	"syscall"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/driver"
	"fyne.io/fyne/v2/driver/desktop"
	"github.com/MordFustang21/marvin-go/internal/search"
	"github.com/MordFustang21/marvin-go/internal/search/providers/calculator"
	"github.com/MordFustang21/marvin-go/internal/search/providers/commands"
	"github.com/MordFustang21/marvin-go/internal/search/providers/spotlight"
	"github.com/MordFustang21/marvin-go/internal/search/providers/web"
	"github.com/MordFustang21/marvin-go/internal/theme"
	"github.com/MordFustang21/marvin-go/internal/ui"
	"github.com/MordFustang21/marvin-go/internal/ui/assets"
	screenmanager "github.com/MordFustang21/marvin-go/internal/util/screen_manager"
	hook "github.com/robotn/gohook"
)

var (
	// goNSWindowPtr will store the pointer to the NSWindow as a Go uintptr.
	// This is used to call C functions that require the NSWindow pointer.
	goNSWindowPtr uintptr
)

func main() {
	// Create a new Fyne application with custom ID
	marvin := app.NewWithID("com.mordfustang.marvin")

	// Set system tray so the app can run in the background.
	if desk, ok := marvin.(desktop.App); ok {
		desk.SetSystemTrayIcon(assets.MarvinIcon)
	}

	// Apply our custom GitHub Dark theme
	marvin.Settings().SetTheme(theme.NewGitHubDarkTheme())

	// Initialize search providers
	registry := search.NewRegistry()
	setupSearchProviders(registry)

	// Create the search window with the registry.
	searchWindow := ui.NewSearchWindow(marvin, registry)

	// Attempt to get the NSWindow pointer for the search window.
	marvin.Lifecycle().SetOnEnteredForeground(func() {
		nativeWin, ok := searchWindow.GetWindow().(driver.NativeWindow)
		if !ok {
			slog.Debug("Window does not support driver.NativeWindow")
			return
		}

		// RunNative must be called on the main Fyne goroutine
		nativeWin.RunNative(func(ctx any) { // ctx is platform specific
			if runtime.GOOS == "darwin" {
				macCtx, ok := ctx.(driver.MacWindowContext)
				if !ok {
					slog.Debug("Failed to get MacWindowContext from RunNative callback")
					return
				}
				// macCtx.NSWindow is unsafe.Pointer
				// Store it as Go's uintptr
				goNSWindowPtr = uintptr(macCtx.NSWindow)
				slog.Debug("Got NSWindow pointer", slog.Any("NSWindowPtr", goNSWindowPtr))
			} else {
				slog.Debug("Screen management is only supported on macOS")
			}
		})
	})

	// Setup shortcuts for cmd+space to toggle the window.
	go func() {
		hook.Register(hook.KeyDown, []string{"cmd", "space"}, func(e hook.Event) {
			fyne.Do(func() {
				if searchWindow.IsVisible() {
					searchWindow.Hide()
				} else {
					searchWindow.ShowWithKeyboardFocus()
					screenmanager.GoMoveToScreenWithMouse(goNSWindowPtr)
				}
			})
		})

		hook.Process(hook.Start())
	}()

	// Set up a signal handler for graceful shutdown
	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-signalCh
		marvin.Quit()
		os.Exit(0)
	}()

	// Run the application
	searchWindow.GetWindow().ShowAndRun()
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
