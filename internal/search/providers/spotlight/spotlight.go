package spotlight

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
	"github.com/MordFustang21/marvin-go/internal/search"
	"github.com/MordFustang21/marvin-go/internal/util"
)

// Provider is a search provider that uses macOS Spotlight
type Provider struct {
	priority   int
	maxResults int
}

// NewProvider creates a new Spotlight provider
func NewProvider(priority, maxResults int) *Provider {
	if maxResults <= 0 {
		maxResults = 20 // Default max results if invalid value provided
	}
	
	return &Provider{
		priority:   priority,
		maxResults: maxResults,
	}
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
		
		// Get file metadata
		_, iconResource := p.determineKindAndIcon(path)
		
		name := p.extractNameFromPath(path)
		
		// Create closure for the item's path for action handling
		pathCopy := path // Copy to avoid closure capturing loop variable
		
		// Create a more user-friendly description
		var description string
		if strings.HasSuffix(pathCopy, ".app") {
			description = "Application"
		} else {
			// Get parent directory for files
			parentDir := p.getParentDirectory(pathCopy)
			if parentDir != "" {
				description = "in " + parentDir
			} else {
				description = pathCopy
			}
		}

		results = append(results, search.SearchResult{
			Title:       name,
			Description: description,
			Path:        path,
			Icon:        iconResource,
			Type:        search.TypeFile,
			Action: func() {
				OpenFile(pathCopy)
			},
		})
		
		count++
	}
	
	return results, nil
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
	} else if strings.Contains(path, ".") {
		// Get an appropriate icon based on file type
		return "file", util.GetSystemIcon(path)
	} else {
		// Default to folder
		return "folder", theme.FolderIcon()
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

// OpenFile opens a file or application using the 'open' command
func OpenFile(path string) error {
	cmd := exec.Command("open", path)
	return cmd.Run()
}