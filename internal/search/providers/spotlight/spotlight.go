package spotlight

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log/slog"
	"os/exec"
	"path/filepath"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
	"github.com/MordFustang21/marvin-go/internal/search"
	"github.com/MordFustang21/marvin-go/internal/util"
)

// CachedResultProvider is a search provider that uses macOS Spotlight
type Provider struct {
	priority   int
	maxResults int
	// cachedApps stores pre-indexed applications for quick access
	cachedApps map[string][]search.SearchResult // maps lowercase name -> results
}

// NewProvider creates a new Spotlight provider
func NewProvider(priority, maxResults int) *Provider {
	if maxResults <= 0 {
		maxResults = 20 // Default max results if invalid value provided
	}

	provider := &Provider{
		priority:   priority,
		maxResults: maxResults,
		cachedApps: make(map[string][]search.SearchResult),
	}

	// Cache applications in the background
	go provider.cacheApplications()

	return provider
}

// Name returns the provider's name
func (p *Provider) Name() string {
	return "Spotlight"
}

// Type returns the provider type
func (p *Provider) Type() search.ProviderType {
	return search.TypeFile
}

// Priority returns the provider's priority
func (p *Provider) Priority() int {
	return p.priority
}

// CanHandle returns whether the provider can handle the given query
func (p *Provider) CanHandle(query string) bool {
	return query != "" && len(query) > 2 // Only handle queries with at least 3 characters
}

// Search performs a Spotlight search with the given query
func (p *Provider) Search(query string) ([]search.SearchResult, error) {
	if query == "" {
		return []search.SearchResult{}, nil
	}

	// Normalize query for case-insensitive matching
	queryLower := strings.ToLower(query)

	// First check the cached applications and return any matches immediately
	cachedResults := p.searchCachedApps(queryLower)

	// If we have cached results, return them immediately
	if len(cachedResults) > 0 {
		// Return the cached results immediately
		return cachedResults, nil
	}

	// If no cached results, perform synchronous search
	return p.searchSpotlight(queryLower)
}

// searchCachedApps returns applications from the cache that match the query
func (p *Provider) searchCachedApps(queryLower string) []search.SearchResult {
	results := []search.SearchResult{}

	// Check for exact matches in cache first
	if cachedResults, ok := p.cachedApps[queryLower]; ok {
		return cachedResults
	}

	// Otherwise, do prefix matching on keys
	for appName, cachedResults := range p.cachedApps {
		if strings.HasPrefix(appName, queryLower) {
			results = append(results, cachedResults...)
			if len(results) >= p.maxResults {
				return results[:p.maxResults]
			}
		}
	}

	// If we still don't have enough, try substring matching
	if len(results) < p.maxResults {
		for appName, cachedResults := range p.cachedApps {
			if strings.Contains(appName, queryLower) && !strings.HasPrefix(appName, queryLower) {
				results = append(results, cachedResults...)
				if len(results) >= p.maxResults {
					return results[:p.maxResults]
				}
			}
		}
	}

	return results
}

// searchSpotlight performs the actual spotlight search
func (p *Provider) searchSpotlight(query string) ([]search.SearchResult, error) {
	// Format the mdfind query
	// We'll search for applications, files, folders that match the query
	mdFindQuery := fmt.Sprintf("kind:app %s", query)

	cmd := exec.Command("mdfind", mdFindQuery)
	var out bytes.Buffer
	cmd.Stdout = &out

	err := cmd.Run()
	if err != nil {
		return nil, fmt.Errorf("spotlight search failed: %w", err)
	}

	// Process results
	results := []search.SearchResult{}
	paths := strings.Split(strings.TrimSpace(out.String()), "\n")

	// Limit results to maxResults
	count := 0
	for _, path := range paths {
		if path == "" {
			continue
		}

		if count >= p.maxResults {
			break
		}

		// Skip if this is not an application (we're only looking for .app files)
		if !strings.HasSuffix(path, ".app") {
			continue
		}

		// Create a search result from the path
		result, err := p.createSearchResultFromPath(path)
		if err != nil {
			continue
		}

		results = append(results, result)
		count++
	}

	return results, nil
}

// createSearchResultFromPath creates a SearchResult from a file path
func (p *Provider) createSearchResultFromPath(path string) (search.SearchResult, error) {
	// Get file metadata
	kind, iconResource := p.determineKindAndIcon(path)

	// Create closure for the item's path for action handling
	pathCopy := path // Copy to avoid closure capturing loop variable

	var title, description string

	// Try to get metadata using mdls first as it's faster
	mdlsInfo := p.extractMdlsMetadata(path)

	if kind == "application" && strings.HasSuffix(path, ".app") {
		// Extract app bundle information for better display
		appInfo := p.extractAppBundleInfo(path)

		// Use the bundle display name if available, otherwise fallback to filename
		if appInfo.DisplayName != "" {
			title = appInfo.DisplayName
		} else {
			title = p.extractNameFromPath(path)
		}

		// Use the bundle description if available
		if appInfo.Description != "" {
			description = appInfo.Description
		} else if appInfo.ShortVersion != "" {
			description = fmt.Sprintf("Version %s", appInfo.ShortVersion)
		} else {
			// Try to get more metadata using mdls as fallback
			if mdlsInfo.KindDisplayName != "" {
				description = mdlsInfo.KindDisplayName
			} else if mdlsInfo.Version != "" {
				description = fmt.Sprintf("Version %s", mdlsInfo.Version)
			} else if mdlsInfo.Description != "" {
				description = mdlsInfo.Description
			} else if mdlsInfo.Developer != "" {
				description = "By " + mdlsInfo.Developer
			} else {
				description = "Application"
			}
		}
	} else {
		// Use display name from mdls if available, otherwise use filename
		if mdlsInfo.DisplayName != "" {
			title = mdlsInfo.DisplayName
		} else {
			title = p.extractNameFromPath(path)
		}

		// Use the most informative description available
		if mdlsInfo.Description != "" {
			description = mdlsInfo.Description
		} else if mdlsInfo.KindDisplayName != "" {
			description = mdlsInfo.KindDisplayName
		} else {
			// Get parent directory for files
			parentDir := p.getParentDirectory(pathCopy)
			if parentDir != "" {
				description = "in " + parentDir
			} else {
				description = pathCopy
			}
		}
	}

	return search.SearchResult{
		Title:       title,
		Description: description,
		Path:        path,
		Icon:        iconResource,
		Type:        search.TypeFile,
		Action: func() {
			err := OpenFile(pathCopy)
			if err != nil {
				slog.Error("Failed to open file", slog.String("path", pathCopy), slog.Any("error", err))
			}
		},
	}, nil
}

// extractNameFromPath extracts the file or folder name from a path
func (p *Provider) extractNameFromPath(path string) string {
	parts := strings.Split(path, "/")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	return path
}

// getParentDirectory returns the parent directory name
func (p *Provider) getParentDirectory(path string) string {
	parts := strings.Split(path, "/")
	if len(parts) > 2 {
		return parts[len(parts)-2]
	}
	return ""
}

// determineKindAndIcon determines the kind and icon based on the file path
func (p *Provider) determineKindAndIcon(path string) (kind string, icon fyne.Resource) {
	if strings.HasSuffix(path, ".app") {
		// Use our custom icon extraction utility for app bundles
		return "application", util.GetAppIcon(path)
	} else if strings.HasSuffix(path, "/") || !strings.Contains(filepath.Base(path), ".") {
		// More reliable folder detection - ends with slash or has no extension
		return "folder", theme.FolderIcon()
	} else {
		// Get an appropriate icon based on file type
		return "file", util.GetSystemIcon(path)
	}
}

// Execute triggers an action for the given result
func (p *Provider) Execute(result search.SearchResult) error {
	if result.Type != search.TypeFile {
		return fmt.Errorf("not a file result")
	}

	if result.Action != nil {
		result.Action()
	}

	return nil
}

// AppBundleInfo stores information extracted from an app bundle's Info.plist
type AppBundleInfo struct {
	DisplayName      string
	BundleID         string
	Description      string
	ShortVersion     string
	Version          string
	MinimumOSVersion string
}

// MdlsMetadata stores metadata extracted from mdls command
type MdlsMetadata struct {
	DisplayName     string
	KindDisplayName string
	Version         string
	ContentType     string
	ContentTypeTree []string
	LastUsedDate    string
	Developer       string
	Description     string
}

// extractAppBundleInfo gets metadata from an app bundle's Info.plist
func (p *Provider) extractAppBundleInfo(appPath string) AppBundleInfo {
	info := AppBundleInfo{}

	// Path to the Info.plist file in the app bundle
	infoPlist := filepath.Join(appPath, "Contents", "Info.plist")

	// Extract display name (CFBundleDisplayName or CFBundleName)
	info.DisplayName = p.getPlistValue(infoPlist, "CFBundleDisplayName")
	if info.DisplayName == "" {
		info.DisplayName = p.getPlistValue(infoPlist, "CFBundleName")
	}

	// If still no display name, use the filename without .app
	if info.DisplayName == "" {
		name := p.extractNameFromPath(appPath)
		if strings.HasSuffix(name, ".app") {
			info.DisplayName = name[:len(name)-4]
		} else {
			info.DisplayName = name
		}
	}

	// Extract other useful information
	info.BundleID = p.getPlistValue(infoPlist, "CFBundleIdentifier")
	info.ShortVersion = p.getPlistValue(infoPlist, "CFBundleShortVersionString")
	info.Version = p.getPlistValue(infoPlist, "CFBundleVersion")
	info.MinimumOSVersion = p.getPlistValue(infoPlist, "LSMinimumSystemVersion")

	// Get copyright or other description that might be useful
	copyright := p.getPlistValue(infoPlist, "NSHumanReadableCopyright")
	if copyright != "" {
		info.Description = copyright
	}

	return info
}

// extractMdlsMetadata gets additional metadata using mdls command
func (p *Provider) extractMdlsMetadata(path string) MdlsMetadata {
	info := MdlsMetadata{}

	// Run mdls command to get metadata in JSON format
	cmd := exec.Command("mdls", "-name", "kMDItemDisplayName",
		"-name", "kMDItemVersion",
		"-name", "kMDItemKind",
		"-name", "kMDItemContentType",
		"-name", "kMDItemContentTypeTree",
		"-name", "kMDItemLastUsedDate",
		"-name", "kMDItemDeveloper",
		"-name", "kMDItemDescription",
		"-name", "kMDItemComment",
		"-json", path)

	var out bytes.Buffer
	cmd.Stdout = &out

	err := cmd.Run()
	if err != nil {
		return info
	}

	// Parse JSON response
	var result map[string]any
	if err := json.Unmarshal(out.Bytes(), &result); err != nil {
		return info
	}

	// Extract values
	if displayName, ok := result["kMDItemDisplayName"]; ok && displayName != nil {
		info.DisplayName = fmt.Sprintf("%v", displayName)
	}

	if kind, ok := result["kMDItemKind"]; ok && kind != nil {
		info.KindDisplayName = fmt.Sprintf("%v", kind)
	}

	if version, ok := result["kMDItemVersion"]; ok && version != nil {
		info.Version = fmt.Sprintf("%v", version)
	}

	if developer, ok := result["kMDItemDeveloper"]; ok && developer != nil {
		info.Developer = fmt.Sprintf("%v", developer)
	}

	if description, ok := result["kMDItemDescription"]; ok && description != nil {
		info.Description = fmt.Sprintf("%v", description)
	} else if comment, ok := result["kMDItemComment"]; ok && comment != nil {
		info.Description = fmt.Sprintf("%v", comment)
	}

	return info
}

// getPlistValue uses PlistBuddy to extract a value from a plist file
func (p *Provider) getPlistValue(plistPath, key string) string {
	cmd := exec.Command("/usr/libexec/PlistBuddy", "-c", "Print :"+key, plistPath)
	var out bytes.Buffer
	cmd.Stdout = &out

	// Ignore errors since some keys might not exist
	err := cmd.Run()
	if err != nil {
		return ""
	}

	value := strings.TrimSpace(out.String())

	// Handle PlistBuddy's "Does Not Exist" response
	if strings.Contains(value, "Does Not Exist") {
		return ""
	}

	return value
}

// OpenFile opens a file or application using the 'open' command
func OpenFile(path string) error {
	cmd := exec.Command("open", path)
	return cmd.Run()
}

// cacheApplications caches applications from /Applications and /System/Applications
func (p *Provider) cacheApplications() {
	// Path to the Applications directory
	standardApps := "/Applications"
	systemApps := "/System/Applications"

	// Cache applications from both directories
	p.cacheAppsFromDirectory(standardApps)
	p.cacheAppsFromDirectory(systemApps)

	slog.Debug("Application cache initialized", slog.Int("numEntries", len(p.cachedApps)))
}

// cacheAppsFromDirectory scans a directory for .app files and caches them
func (p *Provider) cacheAppsFromDirectory(dirPath string) {
	// Find all .app files in the directory
	cmd := exec.Command("find", dirPath, "-name", "*.app", "-maxdepth", "1")
	var out bytes.Buffer
	cmd.Stdout = &out

	err := cmd.Run()
	if err != nil {
		slog.Error("Error finding applications", slog.String("directory", dirPath), slog.Any("error", err))
		return
	}

	// Process the applications
	appPaths := strings.Split(strings.TrimSpace(out.String()), "\n")
	for _, path := range appPaths {
		if path == "" {
			continue
		}

		// Create a search result for this application
		result, err := p.createSearchResultFromPath(path)
		if err != nil {
			continue
		}

		// Create searchable index keys:
		// - Full app name
		// - Words in app name
		keys := []string{
			strings.ToLower(result.Title),
		}

		// Add individual words as keys
		words := strings.Fields(strings.ToLower(result.Title))
		for _, word := range words {
			if len(word) > 2 && !contains(keys, word) { // Only add words longer than 2 chars
				keys = append(keys, word)
			}
		}

		// Add to cache using all keys
		for _, key := range keys {
			p.cachedApps[key] = append(p.cachedApps[key], result)
		}
	}
}

// contains checks if a string is in a slice
func contains(slice []string, str string) bool {
	for _, s := range slice {
		if s == str {
			return true
		}
	}
	return false
}
