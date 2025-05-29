package search

import (
	"fmt"
	"sort"
	"sync"
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

// Search performs a search across all registered providers concurrently
func (r *Registry) Search(query string) ([]SearchResult, error) {
	if query == "" {
		return []SearchResult{}, nil
	}

	var (
		allResults []SearchResult
		mu         sync.Mutex
		wg         sync.WaitGroup
	)

	resultsCh := make(chan []SearchResult)
	errCh := make(chan error)
	providerCount := 0

	for _, provider := range r.providers {
		if !provider.CanHandle(query) {
			continue
		}
		providerCount++
		wg.Add(1)
		go func(p Provider) {
			defer wg.Done()
			results, err := p.Search(query)
			if err != nil {
				errCh <- err
				return
			}
			resultsCh <- results
		}(provider)
	}

	// Close channels when all goroutines are done
	go func() {
		wg.Wait()
		close(resultsCh)
		close(errCh)
	}()

	var lastErr error
	received := 0
	for received < providerCount {
		select {
		case results, ok := <-resultsCh:
			if ok {
				mu.Lock()
				allResults = append(allResults, results...)
				mu.Unlock()
				received++
			}
		case err, ok := <-errCh:
			if ok {
				lastErr = err
				received++
			}
		}
	}

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
