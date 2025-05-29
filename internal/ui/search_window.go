package ui

import (
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
}

// NewSearchResult creates a new search result widget
func NewSearchResult(title, description, path string, icon fyne.Resource, onTap func()) *SearchResult {
	result := &SearchResult{
		Title:       title,
		Description: description,
		Path:        path,
		Icon:        icon,
		OnTap:       onTap,
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

	return &searchResultRenderer{
		result:     r,
		content:    content,
		background: canvas.NewRectangle(fyne.CurrentApp().Settings().Theme().Color(theme.ColorNameBackground, theme.VariantDark)),
		objects:    []fyne.CanvasObject{content},
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
	window      fyne.Window
	searchInput *widget.Entry
	resultsList *fyne.Container
	searcher    *search.SpotlightSearcher
	timer       *time.Timer
	isFrameless bool
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

	// Add results to the list
	for _, result := range results {
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

		// Create a search result item
		resultItem := NewSearchResult(
			result.Name,
			result.Path,
			result.Path,
			icon,
			func(path string) func() {
				return func() {
					search.OpenFile(path)
					sw.Hide() // Hide the window after selection
				}
			}(result.Path),
		)

		sw.resultsList.Add(resultItem)
	}

	sw.resultsList.Refresh()
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
