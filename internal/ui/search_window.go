package ui

import (
	"fmt"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/widget"
	"github.com/MordFustang21/marvin-go/internal/search"
	"github.com/MordFustang21/marvin-go/internal/util"
)

const (
	searchDelay     = 50 * time.Millisecond
	defaultWidth    = 650
	defaultHeight   = 450
	searchBarHeight = 50
	resultRowHeight = 70
	cornerRadius    = 10
	iconWidth       = 48 // Increased icon width for better app icon display
	descMaxLines    = 2
)

// SearchWindow represents the main search window of the application
type SearchWindow struct {
	window        fyne.Window
	searchInput   *SearchEntry
	resultsList   *fyne.Container
	registry      *search.Registry
	timer         *time.Timer
	isFrameless   bool
	selectedIndex int
	results       []search.SearchResult
	resultItems   []*SearchResultItem
	// show is used to track if the window is currently visible.
	show bool
	// resultMap tracks results by their unique identifiers to prevent duplicates
	resultMap map[string]int
}

// NewSearchWindow creates a new search window
func NewSearchWindow(app fyne.App, registry *search.Registry) *SearchWindow {
	var window fyne.Window
	if drv, ok := app.Driver().(desktop.Driver); ok {
		window = drv.CreateSplashWindow()
	} else {
		window = app.NewWindow("Marvin")
	}

	// Set show menu item with shortcut to stop alert sound when showing the window.
	showMenuItem := fyne.NewMenuItem("Show Marvin", func() {})
	showMenuItem.Shortcut = &util.ShortcutLauncher{}
	m := fyne.NewMenu("show", showMenuItem)
	mm := fyne.NewMainMenu(m)

	window.SetMainMenu(mm)

	window.SetFixedSize(true)
	window.Resize(fyne.NewSize(defaultWidth, defaultHeight))
	// Window positioning is now handled in ShowWithKeyboardFocus

	// Create a custom styled search input
	searchInput := NewSearchEntry()
	searchInput.SetPlaceHolder(util.GetRandomQuote())
	searchInput.TextStyle = fyne.TextStyle{Bold: true}
	// Entry text size is controlled by theme

	resultsList := container.NewVBox()
	resultsScroll := container.NewScroll(resultsList)

	// Create the main search window
	searchWindow := &SearchWindow{
		window:      window,
		searchInput: searchInput,
		resultsList: resultsList,
		registry:    registry,
		isFrameless: true,
		resultMap:   make(map[string]int),
	}

	// Create trigger for search on input submission.
	// This is so we can launch a selected search result.
	searchInput.OnSubmitted = func(text string) {
		searchWindow.launchSelectedResult()
	}

	// Handle navigation keys even when entry has focus
	searchInput.OnSpecialKey = func(key *fyne.KeyEvent) {
		switch key.Name {
		case fyne.KeyEscape:
			// Clear the search input or hide the window if empty.
			if searchInput.Text != "" {
				searchInput.SetText("")
			} else {
				searchWindow.Hide()
			}
		case fyne.KeyDown:
			searchWindow.selectNextResult()
		case fyne.KeyUp:
			searchWindow.selectPreviousResult()
		case fyne.KeyReturn:
			searchWindow.launchSelectedResult()
		}
	}

	// Set up the search delay timer
	searchWindow.timer = time.NewTimer(searchDelay)
	go func() {
		for range searchWindow.timer.C {
			searchWindow.performSearch(searchInput.Text)
		}
	}()

	// Hook up events
	searchInput.OnChanged = func(text string) {
		// Reset the timer for each keystroke
		if searchWindow.timer != nil {
			searchWindow.timer.Reset(searchDelay)
		}
	}

	// Create a cleaner layout with custom padding and styling
	searchInputContainer := container.NewPadded(searchInput)

	// Create main content layout with better spacing
	searchBox := widget.NewCard("", "", searchInputContainer)

	mainContainer := container.NewBorder(
		searchBox,
		nil, nil, nil,
		resultsScroll,
	)

	// Set the content
	window.SetContent(mainContainer)

	// Focus the search input when the window opens
	window.Canvas().Focus(searchInput)

	return searchWindow
}

// selectResult sets the visual selection to the specified index
func (sw *SearchWindow) selectResult(index int) {
	if len(sw.resultItems) == 0 {
		return
	}

	// Ensure index is valid
	if index < 0 {
		index = 0
	} else if index >= len(sw.resultItems) {
		index = len(sw.resultItems) - 1
	}

	// Deselect current selection
	if sw.selectedIndex >= 0 && sw.selectedIndex < len(sw.resultItems) {
		sw.resultItems[sw.selectedIndex].IsSelected = false
		sw.resultItems[sw.selectedIndex].Refresh()
	}

	// Select the new item
	sw.selectedIndex = index
	sw.resultItems[index].IsSelected = true
	sw.resultItems[index].Refresh()

	// Refresh the results list container to ensure UI updates
	fyne.Do(func() {
		sw.resultsList.Refresh()
	})
}

// selectNextResult selects the next result in the list
func (sw *SearchWindow) selectNextResult() {
	if len(sw.resultItems) == 0 {
		return
	}

	nextIndex := sw.selectedIndex + 1
	if nextIndex >= len(sw.resultItems) {
		nextIndex = 0 // Wrap around
	}

	sw.selectResult(nextIndex)
}

// selectPreviousResult selects the previous result in the list
func (sw *SearchWindow) selectPreviousResult() {
	if len(sw.resultItems) == 0 {
		return
	}

	prevIndex := sw.selectedIndex - 1
	if prevIndex < 0 {
		prevIndex = len(sw.resultItems) - 1 // Wrap around
	}

	sw.selectResult(prevIndex)
}

// launchSelectedResult launches the currently selected result
func (sw *SearchWindow) launchSelectedResult() {
	if sw.selectedIndex >= 0 && sw.selectedIndex < len(sw.resultItems) {
		if sw.resultItems[sw.selectedIndex].OnTap != nil {
			sw.resultItems[sw.selectedIndex].OnTap()
		}
	}
}

// Close closes the search window and cleans up resources
func (sw *SearchWindow) Close() {
	if sw.timer != nil {
		sw.timer.Stop()
	}

	fyne.Do(func() {
		sw.window.Close()
	})
}

// performSearch executes the search and updates the UI
func (sw *SearchWindow) performSearch(query string) {
	// Remove all previous results and reset state
	sw.results = nil
	sw.resultItems = nil
	sw.resultMap = make(map[string]int)
	sw.selectedIndex = 0 // Start with first item selected

	if query == "" {
		return
	}

	resultsCh := make(chan []search.SearchResult)
	errCh := make(chan error)
	doneCh := make(chan struct{})

	go sw.registry.SearchAsync(query, resultsCh, errCh, doneCh)

	// Track if we've shown any results yet
	var anyResults bool

	// Track results by provider type for proper ordering
	resultsByType := make(map[search.ProviderType][]int)

	go func() {
		for {
			select {
			case results, ok := <-resultsCh:
				if !ok {
					return
				}
				if len(results) == 0 {
					continue
				}

				// If this is the first set of results, ensure the list is clean
				if !anyResults {
					sw.resultsList.RemoveAll()
					// Don't reset resultItems completely, as we want to maintain ordered results
					sw.selectedIndex = 0
				}
				anyResults = true

				// Process new results
				currentItemsCount := len(sw.resultItems)
				for _, result := range results {
					// Create a unique key for this result to prevent duplicates
					resultKey := fmt.Sprintf("%s:%s", string(result.Type), result.Path)

					// Skip if we've already added this result
					if _, exists := sw.resultMap[resultKey]; exists {
						continue
					}

					resultItem := NewSearchResult(result)
					// Truncate long descriptions
					if len(resultItem.Description) > 120 {
						resultItem.Description = resultItem.Description[:117] + "..."
					}
					// Configure the action to hide the window after execution
					originalAction := resultItem.OnTap
					resultItem.OnTap = func() {
						if originalAction != nil {
							originalAction()
						}
						sw.Hide() // Hide the window after selection
					}

					// Track the item by provider type
					resultIndex := len(sw.resultItems)
					resultsByType[result.Type] = append(resultsByType[result.Type], resultIndex)

					// Add result to our data structures and mark as added
					sw.resultItems = append(sw.resultItems, resultItem)
					sw.resultMap[resultKey] = resultIndex
				}

				// If we just got a first batch of results, build the UI
				// This handles initial results
				if currentItemsCount == 0 && len(sw.resultItems) > 0 {
					// Add all items to UI in order
					for _, item := range sw.resultItems {
						sw.resultsList.Add(item)
					}
					// Select first item
					if len(sw.resultItems) > 0 {
						sw.resultItems[0].IsSelected = true
						sw.selectResult(0)
					}
				} else if currentItemsCount > 0 {
					// This handles subsequent results
					// Clear and rebuild the list to maintain proper ordering
					sw.resultsList.RemoveAll()
					for _, item := range sw.resultItems {
						sw.resultsList.Add(item)
					}
				}

				// Refresh the UI
				fyne.Do(func() {
					sw.resultsList.Refresh()
				})

			case err, ok := <-errCh:
				if ok && err != nil {
					sw.resultsList.RemoveAll()
					sw.resultsList.Add(widget.NewLabel("Error: " + err.Error()))
					sw.resultsList.Refresh()
				}
			case <-doneCh:
				// If no results were shown, show "No results found"
				if !anyResults {
					sw.resultsList.RemoveAll()
					sw.resultsList.Add(widget.NewLabel("No results found"))
					sw.resultsList.Refresh()
				}
				return
			}
		}
	}()
}

// Show displays the search window
func (sw *SearchWindow) Show() {
	fyne.Do(func() {
		sw.searchInput.SetPlaceHolder(util.GetRandomQuote())

		sw.window.Show()
	})
}

// Hide hides the search window
func (sw *SearchWindow) Hide() {
	sw.show = false
	fyne.Do(func() {
		sw.window.Hide()
	})
}

// ShowWithKeyboardFocus shows the search window and focuses the search input
// It will attempt to position the window on the active screen (where the mouse cursor is)
func (sw *SearchWindow) ShowWithKeyboardFocus() {
	sw.show = true
	sw.Show()

	// First focus the input field
	sw.window.Canvas().Focus(sw.searchInput)
	sw.searchInput.Entry.TypedShortcut(&fyne.ShortcutSelectAll{})
}

// ClearSearch clears the current search query and results
func (sw *SearchWindow) ClearSearch() {
	sw.searchInput.SetText("")
	sw.resultsList.RemoveAll()
}

// IsVisible returns whether the window is currently visible
func (sw *SearchWindow) IsVisible() bool {
	return sw.show
}

// GetWindow provides access to the underlying window object
func (sw *SearchWindow) GetWindow() fyne.Window {
	return sw.window
}

// SetFrameless configures the window to be as frameless as possible
func (sw *SearchWindow) SetFrameless(frameless bool) {
	sw.isFrameless = frameless
	sw.window.SetTitle("")
	sw.window.SetPadded(false)
}
