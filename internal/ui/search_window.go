package ui

import (
	"context"
	"fmt"
	"log/slog"
	"sort"
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
	// cancellation for current search
	currentCtx       context.Context
	cancelCurrentCtx context.CancelFunc
	// searchTimeout defines how long to wait before considering a search complete
	searchTimeout time.Duration
	// Track results by provider priority for proper ordering
	resultsByPriority map[int][]int
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

	// Create initial no-op context
	ctx, cancel := context.WithCancel(context.Background())

	// Create the main search window
	searchWindow := &SearchWindow{
		window:            window,
		searchInput:       searchInput,
		resultsList:       resultsList,
		registry:          registry,
		isFrameless:       true,
		resultMap:         make(map[string]int),
		currentCtx:        ctx,
		cancelCurrentCtx:  cancel,
		searchTimeout:     500 * time.Millisecond, // Default timeout for considering search complete
		resultsByPriority: make(map[int][]int),    // Track results by provider priority
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

		// Cancel any ongoing search
		searchWindow.cancelCurrentCtx()
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

	slog.Debug("Selecting result",
		"index", index,
		"title", sw.resultItems[index].Title,
		"previousIndex", sw.selectedIndex)

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
	// Cancel any ongoing searches
	if sw.cancelCurrentCtx != nil {
		sw.cancelCurrentCtx()
	}

	if sw.timer != nil {
		sw.timer.Stop()
	}

	fyne.Do(func() {
		sw.window.Close()
	})
}

// performSearch executes the search and updates the UI
func (sw *SearchWindow) performSearch(query string) {
	// Cancel any previous search
	sw.cancelCurrentCtx()

	// Create a new context for this search
	ctx, cancel := context.WithTimeout(context.Background(), sw.searchTimeout)
	sw.currentCtx = ctx
	sw.cancelCurrentCtx = cancel

	// Clean up when we're done with this search
	defer func() {
		// Don't cancel immediately - wait for timeout or next search
		go func() {
			<-time.After(sw.searchTimeout)
			cancel() // Cancel after timeout to clean up resources
		}()
	}()

	// Remove all previous results and reset state
	sw.results = nil
	sw.resultItems = nil
	sw.resultMap = make(map[string]int)
	sw.resultsByPriority = make(map[int][]int)
	sw.selectedIndex = 0 // Start with first item selected

	// Clear the UI right away
	fyne.Do(func() {
		sw.resultsList.RemoveAll()
	})

	if query == "" {
		return
	}

	resultsCh := make(chan []search.SearchResult)
	errCh := make(chan error)
	doneCh := make(chan struct{})

	// Start the search with context for cancellation
	go sw.registry.SearchAsync(ctx, query, resultsCh, errCh, doneCh)

	// Track if we've shown any results yet
	var anyResults bool

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

				// Process new results
				fyne.Do(func() {
					// Only set anyResults to true once we process results
					anyResults = true

					// Get the provider priority from the first result (all results in a batch have same priority)
					if len(results) > 0 {
						// Find provider for these results to get priority
						var providerPriority int
						for _, p := range sw.registry.GetProviders() {
							if p.Type() == results[0].Type {
								providerPriority = p.Priority()
								break
							}
						}

						// Process new results
						newItemIndices := make([]int, 0, len(results))

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

							// Add result to our data structures
							sw.resultItems = append(sw.resultItems, resultItem)
							itemIndex := len(sw.resultItems) - 1
							sw.resultMap[resultKey] = itemIndex

							// Track this result by provider priority
							newItemIndices = append(newItemIndices, itemIndex)
						}

						// Add indices to the priority map
						sw.resultsByPriority[providerPriority] = append(sw.resultsByPriority[providerPriority], newItemIndices...)

						// Rebuild the UI in priority order
						sw.resultsList.RemoveAll()

						// Get all priorities and sort them (lower number = higher priority)
						priorities := make([]int, 0, len(sw.resultsByPriority))
						for p := range sw.resultsByPriority {
							priorities = append(priorities, p)
						}
						sort.Ints(priorities)

						// Rebuild resultItems array to match display order
						orderedResults := make([]*SearchResultItem, 0, len(sw.resultItems))

						// Add results in priority order to both UI and ordered array
						for _, priority := range priorities {
							for _, idx := range sw.resultsByPriority[priority] {
								sw.resultsList.Add(sw.resultItems[idx])
								orderedResults = append(orderedResults, sw.resultItems[idx])
							}
						}

						// Replace resultItems with the ordered version
						sw.resultItems = orderedResults

						// Update resultMap to reflect new indices
						sw.resultMap = make(map[string]int)
						for i, item := range sw.resultItems {
							resultKey := fmt.Sprintf("%s:%s", string(item.searchResult.Type), item.Path)
							sw.resultMap[resultKey] = i
						}

						// Always select the first item after rebuilding the UI to ensure
						// the highest priority result is selected
						if len(sw.resultItems) > 0 {
							slog.Debug("Selecting first result after rebuild",
								"title", sw.resultItems[0].Title,
								"type", sw.resultItems[0].searchResult.Type,
								"totalResults", len(sw.resultItems))

							// Ensure all previous selections are cleared before setting new one
							for i, item := range sw.resultItems {
								item.IsSelected = (i == 0)
								item.Refresh()
							}
							sw.selectedIndex = 0
						}

						// Refresh the UI
						sw.resultsList.Refresh()
					}
				})

			case err, ok := <-errCh:
				if !ok || err == nil {
					continue
				}

				fyne.Do(func() {
					// Only show error if we don't have results yet
					if !anyResults {
						sw.resultsList.RemoveAll()
						sw.resultsList.Add(widget.NewLabel("Error: " + err.Error()))
						sw.resultsList.Refresh()
					}
				})

			case <-doneCh:
				fyne.Do(func() {
					// If no results were shown, show "No results found"
					if !anyResults {
						sw.resultsList.RemoveAll()
						sw.resultsList.Add(widget.NewLabel("No results found"))
						sw.resultsList.Refresh()
					}
				})
				return

			case <-ctx.Done():
				// Search was cancelled
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
	sw.searchInput.SelectAll()
}

// ClearSearch clears the current search query and results
func (sw *SearchWindow) ClearSearch() {
	// Cancel any ongoing searches
	if sw.cancelCurrentCtx != nil {
		sw.cancelCurrentCtx()
	}

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
