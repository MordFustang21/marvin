package ui

import (
	"fmt"
	"image/color"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	"github.com/MordFustang21/marvin-go/internal/search"
	"github.com/MordFustang21/marvin-go/internal/util"
)

const (
	searchDelay     = 300 * time.Millisecond
	defaultWidth    = 650
	defaultHeight   = 450
	searchBarHeight = 50
	resultRowHeight = 70
	cornerRadius    = 10
	iconWidth       = 48 // Increased icon width for better app icon display
	descMaxLines    = 2
)

// SearchResultItem represents a visual row in the search results
type SearchResultItem struct {
	widget.BaseWidget
	Title        string
	Description  string
	Path         string
	Icon         fyne.Resource
	OnTap        func()
	IsSelected   bool
	background   *canvas.Rectangle
	searchResult search.SearchResult // Reference to the original search result
}

// NewSearchResult creates a new search result widget
func NewSearchResult(result search.SearchResult) *SearchResultItem {
	bgColor := color.NRGBA{R: 13, G: 17, B: 23, A: 255} // Default background color
	background := canvas.NewRectangle(bgColor)

	resultItem := &SearchResultItem{
		Title:        result.Title,
		Description:  result.Description,
		Path:         result.Path,
		Icon:         result.Icon,
		OnTap:        result.Action,
		background:   background,
		searchResult: result,
	}
	resultItem.ExtendBaseWidget(resultItem)
	return resultItem
}

// CreateRenderer creates a renderer for the search result
func (r *SearchResultItem) CreateRenderer() fyne.WidgetRenderer {
	title := widget.NewLabel(r.Title)
	title.TextStyle = fyne.TextStyle{Bold: true}
	// Title size is controlled by text style

	description := widget.NewLabel(r.Description)
	description.TextStyle = fyne.TextStyle{}
	description.Wrapping = fyne.TextWrapWord
	description.Truncation = fyne.TextTruncateEllipsis
	// Description size is controlled by text style

	// Ensure description is a lighter color to differentiate from title
	descriptionText := canvas.NewText(r.Description, color.NRGBA{R: 160, G: 170, B: 180, A: 220})
	descriptionText.TextStyle = fyne.TextStyle{}
	descriptionText.TextSize = 14

	// Create a properly sized icon
	icon := widget.NewIcon(r.Icon)

	// Create a fixed width container for text to prevent overflow
	textContainer := container.New(
		layout.NewVBoxLayout(),
		title,
		descriptionText,
	)

	// Limit the text container's width
	textWrapper := container.NewStack(textContainer)

	// Create a nicer icon container with fixed size and proper scaling
	iconSize := fyne.NewSize(iconWidth, iconWidth)
	icon.Resize(iconSize)

	// Create a container that properly centers and scales the icon
	iconWrapper := container.New(
		layout.NewCenterLayout(),
		icon,
	)

	iconContainer := container.New(
		layout.NewPaddedLayout(),
		iconWrapper,
	)
	iconContainer.Move(fyne.NewPos(8, 8))

	content := container.New(
		layout.NewHBoxLayout(),
		iconContainer,
		textWrapper,
	)

	// Set background color based on selection state and add rounded corners
	if r.IsSelected {
		r.background.FillColor = color.NRGBA{R: 35, G: 57, B: 83, A: 255} // #233953 - darker blue for selection
	} else {
		r.background.FillColor = color.NRGBA{R: 18, G: 23, B: 32, A: 255} // Slightly lighter than default for contrast
	}
	r.background.CornerRadius = 6 // Add rounded corners to results

	return &searchResultRenderer{
		result:     r,
		content:    content,
		background: r.background,
		objects:    []fyne.CanvasObject{r.background, content},
	}
}

// Tapped handles tap events on the result
func (r *SearchResultItem) Tapped(*fyne.PointEvent) {
	if r.OnTap != nil {
		r.OnTap()
	}
}

// MinSize returns the minimum size of the result
func (r *SearchResultItem) MinSize() fyne.Size {
	// Add a bit more height for better icon display and spacing between items
	return fyne.NewSize(defaultWidth, resultRowHeight)
}

// Refresh refreshes the widget
func (r *SearchResultItem) Refresh() {
	if r.IsSelected {
		r.background.FillColor = color.NRGBA{R: 35, G: 57, B: 83, A: 255} // selected
	} else {
		r.background.FillColor = color.NRGBA{R: 18, G: 23, B: 32, A: 255} // not selected
	}

	r.background.Refresh()
	r.BaseWidget.Refresh()
}

type searchResultRenderer struct {
	result     *SearchResultItem
	content    fyne.CanvasObject
	background *canvas.Rectangle
	objects    []fyne.CanvasObject
}

func (r *searchResultRenderer) Layout(size fyne.Size) {
	r.background.Resize(size)

	// Leave a small margin on the right and add some vertical padding
	contentSize := fyne.NewSize(size.Width-16, size.Height-8)
	r.content.Resize(contentSize)
	r.content.Move(fyne.NewPos(4, 4))

	// Add a subtle separator at the bottom of each result
	if !r.result.IsSelected {
		// Draw a subtle separator line except for selected items
		separator := canvas.NewLine(color.NRGBA{R: 50, G: 60, B: 80, A: 100})
		separator.Position1 = fyne.NewPos(10, size.Height-1)
		separator.Position2 = fyne.NewPos(size.Width-10, size.Height-1)
		separator.StrokeWidth = 1
		separator.Refresh()
	}
}

func (r *searchResultRenderer) MinSize() fyne.Size {
	// Ensure we have a minimum height for each result item
	minSize := r.content.MinSize()
	if minSize.Height < resultRowHeight {
		minSize.Height = resultRowHeight
	}
	return minSize
}

func (r *searchResultRenderer) Objects() []fyne.CanvasObject {
	return []fyne.CanvasObject{r.background, r.content}
}

func (r *searchResultRenderer) Refresh() {
	r.background.Refresh()
	r.content.Refresh()
}

func (r *searchResultRenderer) Destroy() {}

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

	window.SetFixedSize(true)
	window.Resize(fyne.NewSize(defaultWidth, defaultHeight))
	window.CenterOnScreen()

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

	// Set window attributes after creation
	window.SetCloseIntercept(func() {
		window.Hide()
	})

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
	fyne.DoAndWait(func() {
		window.SetContent(mainContainer)
	})

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

// Show displays the search window
func (sw *SearchWindow) Show() {
	fyne.Do(func() {
		sw.window.Show()
	})
}

// Hide hides the search window
func (sw *SearchWindow) Hide() {
	sw.show = false
	sw.window.Hide()
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
	sw.resultsList.RemoveAll()
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

// ShowWithKeyboardFocus shows the search window and focuses the search input
func (sw *SearchWindow) ShowWithKeyboardFocus() {
	sw.show = true
	sw.Show()
	sw.window.Canvas().Focus(sw.searchInput)
	// Fyne v2 does not support select-all; just focus the input and leave text as-is.
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
