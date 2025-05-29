package theme

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

// GitHubDarkTheme is a custom theme that implements the GitHub Dark Default theme
type GitHubDarkTheme struct{
	noTitleBar bool
}

var _ fyne.Theme = (*GitHubDarkTheme)(nil)

// Color returns the color for a specific theme resource
func (t *GitHubDarkTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	// Special handling for title bar and window elements to make them minimal
	if name == theme.ColorNameShadow {
		return color.NRGBA{R: 0, G: 0, B: 0, A: 0} // Transparent shadow
	}
	
	// We want the title text to be visible, but not the window title
	if name == theme.ColorNameForeground && t.noTitleBar {
		return color.NRGBA{R: 230, G: 237, B: 243, A: 255} // Make text visible
	}
	
	switch name {
	case theme.ColorNameBackground:
		return color.NRGBA{R: 13, G: 17, B: 23, A: 255} // #0d1117
	case theme.ColorNameHeaderBackground:
		return color.NRGBA{R: 13, G: 17, B: 23, A: 255} // Solid header
	case theme.ColorNameButton:
		return color.NRGBA{R: 35, G: 134, B: 54, A: 255} // #238636
	case theme.ColorNameDisabled:
		return color.NRGBA{R: 110, G: 118, B: 129, A: 255} // #6e7681
	case theme.ColorNameDisabledButton:
		return color.NRGBA{R: 48, G: 54, B: 61, A: 255} // #30363d
	case theme.ColorNameError:
		return color.NRGBA{R: 248, G: 81, B: 73, A: 255} // #f85149
	case theme.ColorNameFocus:
		return color.NRGBA{R: 31, G: 111, B: 235, A: 255} // #1f6feb
	case theme.ColorNameForeground:
		return color.NRGBA{R: 230, G: 237, B: 243, A: 255} // #e6edf3
	case theme.ColorNameHover:
		return color.NRGBA{R: 110, G: 118, B: 129, A: 102} // #6e768166
	case theme.ColorNameInputBackground:
		return color.NRGBA{R: 1, G: 4, B: 9, A: 255} // #010409
	case theme.ColorNamePlaceHolder:
		return color.NRGBA{R: 125, G: 133, B: 144, A: 255} // #7d8590
	case theme.ColorNamePressed:
		return color.NRGBA{R: 88, G: 166, B: 255, A: 255} // #58a6ff
	case theme.ColorNamePrimary:
		return color.NRGBA{R: 88, G: 166, B: 255, A: 255} // #58a6ff
	case theme.ColorNameScrollBar:
		return color.NRGBA{R: 139, G: 148, B: 158, A: 51} // #8b949e33
	case theme.ColorNameShadow:
		return color.NRGBA{R: 0, G: 0, B: 0, A: 128} // Semi-transparent shadow
	default:
		return theme.DefaultTheme().Color(name, variant)
	}
}

// Font returns the specified font resource
func (t *GitHubDarkTheme) Font(style fyne.TextStyle) fyne.Resource {
	return theme.DefaultTheme().Font(style)
}

// Icon returns the specified icon resource
func (t *GitHubDarkTheme) Icon(name fyne.ThemeIconName) fyne.Resource {
	return theme.DefaultTheme().Icon(name)
}

// Size returns the configured size for the given theme resource
func (t *GitHubDarkTheme) Size(name fyne.ThemeSizeName) float32 {
	switch name {
	case theme.SizeNamePadding:
		return 8
	case theme.SizeNameScrollBar:
		return 10
	case theme.SizeNameScrollBarSmall:
		return 5
	case theme.SizeNameText:
		return 13
	case theme.SizeNameInputBorder:
		return 1
	default:
		return theme.DefaultTheme().Size(name)
	}
}

// NewGitHubDarkTheme creates and returns a new instance of the GitHub Dark Default theme
func NewGitHubDarkTheme() fyne.Theme {
	return &GitHubDarkTheme{
		noTitleBar: true,
	}
}