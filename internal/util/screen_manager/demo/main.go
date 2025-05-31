package main

/*
#include <stdint.h> // <<< ADD THIS INCLUDE

#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework Cocoa

// These are the C function signatures as defined in window_mover.m
// The C compiler needs to know what 'uintptr_t' is from <stdint.h>
void MoveToMainScreen(uintptr_t nsWindowPtr);
void MoveToScreenWithMouse(uintptr_t nsWindowPtr);
void MoveToScreenAtIndex(uintptr_t nsWindowPtr, int screenIndex);
int GetScreenCount();
*/
import "C"

import (
	"fmt"
	"log"
	"runtime"
	"strconv"

	// Needed for unsafe.Pointer
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver"
	"fyne.io/fyne/v2/widget"
)

var (
	// nsWindowPtr will store the pointer value as a Go uintptr,
	// then cast to C.uintptr_t when calling C functions.
	goNSWindowPtr uintptr
)

func callMoveToMainScreen() {
	if goNSWindowPtr == 0 {
		log.Println("Go: NSWindow pointer not yet available.")
		return
	}
	log.Println("Go: Calling MoveToMainScreen")
	C.MoveToMainScreen(C.uintptr_t(goNSWindowPtr)) // Cast Go uintptr to C.uintptr_t
}

func callMoveToScreenWithMouse() {
	if goNSWindowPtr == 0 {
		log.Println("Go: NSWindow pointer not yet available.")
		return
	}
	log.Println("Go: Calling MoveToScreenWithMouse")
	C.MoveToScreenWithMouse(C.uintptr_t(goNSWindowPtr)) // Cast
}

func callMoveToScreenAtIndex(index int) {
	if goNSWindowPtr == 0 {
		log.Println("Go: NSWindow pointer not yet available.")
		return
	}
	log.Printf("Go: Calling MoveToScreenAtIndex: %d\n", index)
	C.MoveToScreenAtIndex(C.uintptr_t(goNSWindowPtr), C.int(index)) // Cast
}

func getScreenCount() int {
	// This function doesn't need the window pointer if it's just querying system screens
	return int(C.GetScreenCount())
}

func main() {
	runtime.LockOSThread()

	myApp := app.New()
	myWindow := myApp.NewWindow("macOS CGO Screen Mover")

	statusLabel := widget.NewLabel("NSWindow Ptr: Not yet available")

	myApp.Lifecycle().SetOnEnteredForeground(func() {
		log.Println("Go: App entered foreground, attempting to get NSWindow pointer.")
		nativeWin, ok := myWindow.(driver.NativeWindow)
		if !ok {
			log.Panicln("Window does not support driver.NativeWindow")
			return
		}

		// RunNative must be called on the main Fyne goroutine
		nativeWin.RunNative(func(ctx any) { // ctx is platform specific
			if runtime.GOOS == "darwin" {
				macCtx, ok := ctx.(driver.MacWindowContext)
				if !ok {
					log.Println("Failed to get MacWindowContext on RunNative callback")
					return
				}
				// macCtx.NSWindow is unsafe.Pointer
				// Store it as Go's uintptr
				goNSWindowPtr = uintptr(macCtx.NSWindow)
				log.Printf("Go: NSWindow pointer obtained (Go uintptr): %x", goNSWindowPtr)

				if goNSWindowPtr != 0 {
					statusLabel.SetText(fmt.Sprintf("NSWindow Ptr: %x (Ready)", goNSWindowPtr))
					// Update screen buttons now that we can get screen count
					updateScreenButtons(myWindow) // Pass window for RunExclusive
				}

			} else {
				log.Println("Go: Not on macOS, CGO screen moving is macOS specific.")
				statusLabel.SetText("Not on macOS")
			}
		})
	})

	moveToMainBtn := widget.NewButton("Move to Main Screen", callMoveToMainScreen)
	moveToMouseBtn := widget.NewButton("Move to Screen with Mouse", callMoveToScreenWithMouse)
	screenButtonsContainer := container.NewVBox()

	myWindow.SetContent(container.NewVBox(
		statusLabel,
		moveToMainBtn,
		moveToMouseBtn,
		widget.NewLabel("Move to specific screen by index:"),
		screenButtonsContainer,
	))

	// Initial call to populate buttons (will likely show 0 initially if goNSWindowPtr is not set)
	// The SetOnEnteredForeground will call it again once the pointer is available.
	updateScreenButtonsUI(screenButtonsContainer, 0, myWindow) // Show 0 buttons initially

	myWindow.Resize(fyne.NewSize(400, 350))
	myWindow.ShowAndRun()
}

// updateScreenButtons is called when the nsWindowPtr is likely available
func updateScreenButtons(win fyne.Window) {
	if goNSWindowPtr == 0 {
		log.Println("Go: updateScreenButtons called but NSWindow pointer not ready.")
		return // Can't get screen count yet
	}
	numScreens := getScreenCount()
	log.Printf("Go: System reports %d screens.\n", numScreens)

	// Find the VBox container for screen buttons in the window content
	// This is a bit brittle; a more robust way would be to pass the container directly
	// or have a dedicated function to rebuild that part of the UI.
	// For simplicity, assuming it's the last VBox in main VBox.
	mainVBox, _ := win.Content().(*fyne.Container)
	var screenButtonsContainer *fyne.Container
	if mainVBox != nil && len(mainVBox.Objects) > 0 {
		// Assuming screenButtonsContainer is the last element added in main()
		if sbc, ok := mainVBox.Objects[len(mainVBox.Objects)-1].(*fyne.Container); ok {
			screenButtonsContainer = sbc
		}
	}

	if screenButtonsContainer == nil {
		log.Println("Error: Could not find screenButtonsContainer to update.")
		return
	}

	updateScreenButtonsUI(screenButtonsContainer, numScreens, win)
}

// updateScreenButtonsUI rebuilds the screen buttons in the given container
func updateScreenButtonsUI(containerToUpdate *fyne.Container, numScreens int, win fyne.Window) {
	fyne.Do(func() { // Ensure UI updates are on Fyne's main goroutine
		var newButtons []fyne.CanvasObject
		for i := 0; i < numScreens; i++ {
			screenIdx := i // Capture for closure
			btn := widget.NewButton("Move to Screen "+strconv.Itoa(screenIdx), func() {
				callMoveToScreenAtIndex(screenIdx)
			})
			newButtons = append(newButtons, btn)
		}
		containerToUpdate.Objects = newButtons
		containerToUpdate.Refresh()
	})
}
