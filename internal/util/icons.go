package util

import (
	"bytes"
	"fmt"
	"image"
	"image/png"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

// IconCache stores loaded app icons to avoid extracting them repeatedly
var IconCache = make(map[string]fyne.Resource)

// GetAppIcon returns a fyne.Resource containing the icon for the specified application path
// If the icon cannot be extracted, it falls back to a default icon
func GetAppIcon(appPath string) fyne.Resource {
	// Check cache first
	if res, exists := IconCache[appPath]; exists {
		return res
	}

	// Ensure this is actually an app bundle
	if !strings.HasSuffix(appPath, ".app") {
		fmt.Printf("Not an app bundle: %s\n", appPath)
		return theme.ComputerIcon()
	}

	// Try to extract the icon using macOS-specific methods
	iconResource, err := extractMacOSAppIcon(appPath)
	if err != nil || iconResource == nil {
		fmt.Printf("Failed to extract icon for %s: %v\n", appPath, err)
		// Fallback to default icon
		return theme.ComputerIcon()
	}

	// Cache the result
	IconCache[appPath] = iconResource
	return iconResource
}

// extractMacOSAppIcon extracts an application icon from a .app bundle on macOS
func extractMacOSAppIcon(appPath string) (fyne.Resource, error) {
	// First, try to find the icon file within the app bundle
	iconPath := filepath.Join(appPath, "Contents", "Resources", "AppIcon.icns")
	if _, err := os.Stat(iconPath); os.IsNotExist(err) {
		// If AppIcon.icns doesn't exist, look for the icon specified in Info.plist
		infoPath := filepath.Join(appPath, "Contents", "Info.plist")
		if _, err := os.Stat(infoPath); os.IsNotExist(err) {
			return nil, fmt.Errorf("Info.plist not found in app bundle")
		}

		// Extract the icon file name from Info.plist
		cmd := exec.Command("defaults", "read", infoPath, "CFBundleIconFile")
		output, err := cmd.Output()
		if err != nil {
			return nil, fmt.Errorf("failed to read icon info: %w", err)
		}

		iconFile := strings.TrimSpace(string(output))
		if !strings.HasSuffix(iconFile, ".icns") {
			iconFile += ".icns"
		}

		iconPath = filepath.Join(appPath, "Contents", "Resources", iconFile)
		if _, err := os.Stat(iconPath); os.IsNotExist(err) {
			return nil, fmt.Errorf("icon file not found: %s", iconPath)
		}
	}

	// Use sips to convert the icon to PNG and extract it
	tmpFile := filepath.Join(os.TempDir(), fmt.Sprintf("marvin_icon_%d.png", os.Getpid()))
	defer os.Remove(tmpFile) // Clean up temp file when done

	// Convert icns to png using sips
	cmd := exec.Command("sips", "-s", "format", "png", iconPath, "--out", tmpFile)
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("failed to convert icon: %w", err)
	}

	// Read the resulting PNG file
	iconData, err := os.ReadFile(tmpFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read icon: %w", err)
	}

	// Create a fyne resource from the icon data
	iconName := filepath.Base(appPath) + "_icon"
	return fyne.NewStaticResource(iconName, iconData), nil
}

// GetSystemIcon tries to get a system-provided icon for a file or folder
func GetSystemIcon(path string) fyne.Resource {
	// For non-app files, try to get a system icon
	// Read file info to determine type
	fileInfo, err := os.Stat(path)
	if err != nil {
		// If we can't get file info, return a generic icon
		return theme.FileIcon()
	}

	if fileInfo.IsDir() {
		return theme.FolderIcon()
	}

	// Return appropriate icon based on file extension
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".pdf":
		return theme.DocumentIcon()
	case ".jpg", ".jpeg", ".png", ".gif", ".bmp":
		return theme.MediaPhotoIcon()
	case ".mp3", ".wav", ".aac", ".flac":
		return theme.MediaMusicIcon()
	case ".mp4", ".mov", ".avi", ".mkv":
		return theme.MediaVideoIcon()
	case ".txt", ".doc", ".docx", ".rtf":
		return theme.DocumentIcon()
	case ".zip", ".tar", ".gz", ".rar":
		return theme.FileIcon() // Use generic file icon for compressed files
	default:
		return theme.FileIcon()
	}
}

// GetIconForPath returns an appropriate icon for the given path
func GetIconForPath(path string) fyne.Resource {
	startTime := time.Now()
	var res fyne.Resource

	if strings.HasSuffix(path, ".app") {
		res = GetAppIcon(path)
	} else {
		res = GetSystemIcon(path)
	}

	fmt.Printf("GetIconForPath took %v for %s\n", time.Since(startTime), path)
	return res
}

// CreateResourceFromImage creates a fyne.Resource from an image
func CreateResourceFromImage(img image.Image, name string) (fyne.Resource, error) {
	var buf bytes.Buffer
	err := png.Encode(&buf, img)
	if err != nil {
		return nil, err
	}

	return fyne.NewStaticResource(name, buf.Bytes()), nil
}

// ResizeImage resizes an image to the specified dimensions
func ResizeImage(img image.Image, width, height int) image.Image {
	// In a real implementation, you'd use a library like "github.com/nfnt/resize"
	// or implement a proper resizing algorithm here
	// For simplicity, we'll just return the original image
	return img
}
