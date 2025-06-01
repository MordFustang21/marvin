package ui

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	"github.com/MordFustang21/marvin-go/internal/search"
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
