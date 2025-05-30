package assets

import (
	_ "embed"

	"fyne.io/fyne/v2"
)

//go:embed marvin.svg
var marvinIconData []byte

// MarvinIcon is the icon for the Marvin application
var MarvinIcon = fyne.NewStaticResource("marvin-icon.png", marvinIconData)
