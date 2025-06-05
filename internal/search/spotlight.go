package search

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
)

// SpotlightResult represents a single result from the Spotlight search
type SpotlightResult struct {
	Name     string `json:"name"`
	Path     string `json:"path"`
	Kind     string `json:"kind"`
	IconName string `json:"iconName,omitempty"`
}

// SpotlightSearcher provides methods to search using MacOS Spotlight
type SpotlightSearcher struct {
	maxResults int
}

// NewSpotlightSearcher creates a new SpotlightSearcher instance
func NewSpotlightSearcher(maxResults int) *SpotlightSearcher {
	if maxResults <= 0 {
		maxResults = 20 // Default max results if invalid value provided
	}

	return &SpotlightSearcher{
		maxResults: maxResults,
	}
}

// Search performs a Spotlight search with the given query
func (s *SpotlightSearcher) Search(query string) ([]SpotlightResult, error) {
	if query == "" {
		return []SpotlightResult{}, nil
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
	results := []SpotlightResult{}
	for path := range strings.SplitSeq(strings.TrimSpace(out.String()), "\n") {
		if path == "" {
			continue
		}

		// Get file metadata
		kind, iconName := determineKindAndIcon(path)

		name := extractNameFromPath(path)

		results = append(results, SpotlightResult{
			Name:     name,
			Path:     path,
			Kind:     kind,
			IconName: iconName,
		})
	}

	return results, nil
}

// extractNameFromPath extracts the file or folder name from a path
func extractNameFromPath(path string) string {
	parts := strings.Split(path, "/")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	return path
}

// determineKindAndIcon determines the kind and icon name based on the file path
func determineKindAndIcon(path string) (kind, iconName string) {
	if strings.HasSuffix(path, ".app") {
		return "application", "computer"
	} else if strings.Contains(path, ".") {
		// This is a simplistic check for files
		return "file", "document"
	} else {
		// Default to folder
		return "folder", "folder"
	}
}

// GetFileMetadata gets additional metadata about a file using mdls
func GetFileMetadata(path string) (map[string]interface{}, error) {
	cmd := exec.Command("mdls", "-json", path)
	var out bytes.Buffer
	cmd.Stdout = &out

	err := cmd.Run()
	if err != nil {
		return nil, fmt.Errorf("metadata lookup failed: %w", err)
	}

	var metadata map[string]interface{}
	if err := json.Unmarshal(out.Bytes(), &metadata); err != nil {
		return nil, fmt.Errorf("failed to parse metadata: %w", err)
	}

	return metadata, nil
}

// OpenFile opens a file or application using the 'open' command
func OpenFile(path string) error {
	cmd := exec.Command("open", path)
	return cmd.Run()
}
