package search

import (
	"fmt"
	"sort"
)

// Registry manages search providers and dispatches search requests
type Registry struct {
	providers []Provider
}

// NewRegistry creates a new provider registry
func NewRegistry() *Registry {
	return &Registry{
		providers: []Provider{},
	}
}

// RegisterProvider adds a new search provider to the registry
func (r *Registry) RegisterProvider(provider Provider) {
	r.providers = append(r.providers, provider)

	// Sort providers by priority
	sort.Slice(r.providers, func(i, j int) bool {
		return r.providers[i].Priority() < r.providers[j].Priority()
	})
}

// Search performs a search across all registered providers
func (r *Registry) Search(query string) ([]SearchResult, error) {
	if query == "" {
		return []SearchResult{}, nil
	}
	
	var allResults []SearchResult
	var lastErr error
	
	for _, provider := range r.providers {
		// Skip providers that can't handle this query
		if !provider.CanHandle(query) {
			continue
		}
		
		// Perform search with this provider
		results, err := provider.Search(query)
		if err != nil {
			lastErr = err
			continue
		}
		
		// Add results
		allResults = append(allResults, results...)
	}
	
	// If we have results but also errors, we'll still return the results we found
	if len(allResults) == 0 && lastErr != nil {
		return nil, fmt.Errorf("search failed: %w", lastErr)
	}
	
	return allResults, nil
}

// ExecuteResult triggers the execution of a specific search result
func (r *Registry) ExecuteResult(result SearchResult) error {
	// Find the provider that can handle this result type
	for _, provider := range r.providers {
		if provider.Type() == result.Type {
			return provider.Execute(result)
		}
	}
	
	return fmt.Errorf("no provider found for result type %s", result.Type)
}

// GetProviders returns all registered providers
func (r *Registry) GetProviders() []Provider {
	return r.providers
}