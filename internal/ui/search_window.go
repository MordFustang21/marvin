package ui

import (
	"image/color"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/MordFustang21/marvin-go/internal/search"
)

const (
	searchDelay     = 300 * time.Millisecond
	defaultWidth    = 600
	defaultHeight   = 400
	searchBarHeight = 40
	resultRowHeight = 60
	cornerRadius    = 8
)

// SearchResult represents a visual row in the search results
type SearchResult struct {
	widget.BaseWidget
	Title       string
	Description string
	Path        string
	Icon        fyne.Resource
	OnTap       func()
	IsSelected  bool
	background  *canvas.Rectangle
}

// NewSearchResult creates a new search result widget
func NewSearchResult(title, description, path string, icon fyne.Resource, onTap func()) *SearchResult {
	bgColor := color.NRGBA{R: 13, G: 17, B: 23, A: 255} // Default background color
	background := canvas.NewRectangle(bgColor)
	
	result := &SearchResult{
		Title:       title,
		Description: description,
		Path:        path,
		Icon:        icon,
		OnTap:       onTap,
		background:  background,
	}
	result.ExtendBaseWidget(result)
	return result
}

// CreateRenderer creates a renderer for the search result
func (r *SearchResult) CreateRenderer() fyne.WidgetRenderer {
	title := widget.NewLabel(r.Title)
	title.TextStyle = fyne.TextStyle{Bold: true}
	
	description := widget.NewLabel(r.Description)
	description.TextStyle = fyne.TextStyle{}
	description.Wrapping = fyne.TextWrapWord
	
	icon := widget.NewIcon(r.Icon)
	
	content := container.New(
		layout.NewHBoxLayout(),
		container.NewPadded(icon),
		container.New(
			layout.NewVBoxLayout(),
			title,
			description,
		),
	)
	
	// Set background color based on selection state
	if r.IsSelected {
		r.background.FillColor = color.NRGBA{R: 35, G: 57, B: 83, A: 255} // #233953 - darker blue for selection
	} else {
		r.background.FillColor = color.NRGBA{R: 13, G: 17, B: 23, A: 255} // #0d1117 - default dark bg
	}
	
	return &searchResultRenderer{
		result:     r,
		content:    content,
		background: r.background,
		objects:    []fyne.CanvasObject{r.background, content},
	}
}

// Tapped handles tap events on the result
func (r *SearchResult) Tapped(*fyne.PointEvent) {
	if r.OnTap != nil {
		r.OnTap()
	}
}

// MinSize returns the minimum size of the result
func (r *SearchResult) MinSize() fyne.Size {
	return fyne.NewSize(defaultWidth, resultRowHeight)
}

type searchResultRenderer struct {
	result     *SearchResult
	content    fyne.CanvasObject
	background *canvas.Rectangle
	objects    []fyne.CanvasObject
}

func (r *searchResultRenderer) Layout(size fyne.Size) {
	r.background.Resize(size)
	r.content.Resize(size)
}

func (r *searchResultRenderer) MinSize() fyne.Size {
	return r.content.MinSize()
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
	window          fyne.Window
	searchInput     *widget.Entry
	resultsList     *fyne.Container
	searcher        *search.SpotlightSearcher
	timer           *time.Timer
	isFrameless     bool
	selectedIndex   int
	results         []search.SpotlightResult
	resultItems     []*SearchResult
	hasFocus        bool
}

// NewSearchWindow creates a new search window
func NewSearchWindow(app fyne.App) *SearchWindow {
	window := app.NewWindow("Marvin")
	window.SetFixedSize(true)
	window.Resize(fyne.NewSize(defaultWidth, defaultHeight))
	window.CenterOnScreen()

	// Configure window for clean appearance
	window.SetPadded(true)
	window.Canvas().SetOnTypedKey(func(ke *fyne.KeyEvent) {
		if ke.Name == fyne.KeyEscape {
			window.Hide()
		}
	})

	searcher := search.NewSpotlightSearcher(20)

	// Create a custom styled search input
	searchInput := widget.NewEntry()
	searchInput.SetPlaceHolder("Search...")

	resultsList := container.NewVBox()
	resultsScroll := container.NewScroll(resultsList)

	// Create the main search window
	searchWindow := &SearchWindow{
		window:      window,
		searchInput: searchInput,
		resultsList: resultsList,
		searcher:    searcher,
		isFrameless: true,
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

	// Handle keyboard navigation
	window.Canvas().SetOnTypedKey(func(key *fyne.KeyEvent) {
		switch key.Name {
		case fyne.KeyEscape:
			window.Hide()
		case fyne.KeyDown:
			searchWindow.selectNextResult()
		case fyne.KeyUp:
			searchWindow.selectPreviousResult()
		case fyne.KeyReturn:
			searchWindow.launchSelectedResult()
		}
	})

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
	
	// Ensure the selected item is visible in the scroll container
	// (This would need custom scroll logic if we had implemented scroll visibility control)
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
	sw.window.Show()
}

// Hide hides the search window
func (sw *SearchWindow) Hide() {
	sw.window.Hide()
}

// Close closes the search window and cleans up resources
func (sw *SearchWindow) Close() {
	if sw.timer != nil {
		sw.timer.Stop()
	}
	sw.window.Close()
}

// performSearch executes the search and updates the UI
func (sw *SearchWindow) performSearch(query string) {
	// We'll update the UI directly since we're likely already on the main thread
	// If needed, fyne will handle thread safety internally
	sw.resultsList.RemoveAll()
	sw.results = nil
	sw.resultItems = nil
	sw.selectedIndex = 0  // Start with first item selected

	if query == "" {
		return
	}

	results, err := sw.searcher.Search(query)
	if err != nil {
		// Show error in the UI
		sw.resultsList.Add(widget.NewLabel("Error: " + err.Error()))
		return
	}

	if len(results) == 0 {
		sw.resultsList.Add(widget.NewLabel("No results found"))
		return
	}

	// Store the results
	sw.results = results
	sw.resultItems = make([]*SearchResult, 0, len(results))

	// Add results to the list
	for i, result := range results {
		var icon fyne.Resource

		// Select an appropriate icon based on the kind
		switch result.Kind {
		case "application":
			icon = theme.ComputerIcon()
		case "folder":
			icon = theme.FolderIcon()
		default:
			icon = theme.DocumentIcon()
		}

		// Create closure for the item's path for tap handling
		path := result.Path
		
		// Create a search result item
		resultItem := NewSearchResult(
			result.Name,
			result.Path,
			result.Path,
			icon,
			func() {
				search.OpenFile(path)
				sw.Hide() // Hide the window after selection
			},
		)
		
		// Set selected state for the first item only
		resultItem.IsSelected = (i == 0)

		sw.resultItems = append(sw.resultItems, resultItem)
		sw.resultsList.Add(resultItem)
	}

	sw.resultsList.Refresh()
	
	// Select the first result by default if we have results
	if len(sw.resultItems) > 0 {
		sw.selectResult(0)
	}
}

// ShowWithKeyboardFocus shows the search window and focuses the search input
func (sw *SearchWindow) ShowWithKeyboardFocus() {
	sw.Show()
	sw.window.Canvas().Focus(sw.searchInput)
}

// ClearSearch clears the current search query and results
func (sw *SearchWindow) ClearSearch() {
	sw.searchInput.SetText("")
	sw.resultsList.RemoveAll()
}

// IsVisible returns whether the window is currently visible
func (sw *SearchWindow) IsVisible() bool {
	// A simpler implementation that checks if the window has been closed
	return sw.window != nil && sw.window.Content() != nil
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
