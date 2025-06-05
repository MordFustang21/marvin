package search

import (
	"fyne.io/fyne/v2"
)

// ProviderType indicates what kind of search provider is being used
type ProviderType string

const (
	// TypeApp indicates a provider that searches for applications
	TypeApp ProviderType = "app"
	// TypeFile indicates a provider that searches for files
	TypeFile ProviderType = "file"
	// TypeCalculator indicates a provider that performs calculations
	TypeCalculator ProviderType = "calculator"
	// TypeWeb indicates a provider that searches the web
	TypeWeb ProviderType = "web"
	// TypeSystem indicates a provider that interacts with system functions
	TypeSystem ProviderType = "system"
)

// SearchResult represents a single search result from any provider
type SearchResult struct {
	Title       string        // Display title for the result
	Description string        // Secondary description text
	Path        string        // Path or identifier (if applicable)
	Icon        fyne.Resource // Icon to display with the result
	Type        ProviderType  // Type of the provider that generated this result
	Action      func()        // Function to execute when the result is selected
}

// Provider defines the interface for search providers
type Provider interface {
	// Name returns the provider's name
	Name() string

	// Type returns the provider type
	Type() ProviderType

	// Priority returns the provider's priority (lower is higher priority)
	// Results from higher priority providers will be shown first
	Priority() int

	// CanHandle returns whether the provider can handle the given query
	CanHandle(query string) bool

	// Search performs a search with the given query and returns results
	Search(query string) ([]SearchResult, error)

	// Execute triggers an action for the given result, if applicable
	Execute(result SearchResult) error
}
