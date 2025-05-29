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

// SearchAsync performs a search and sends results as they arrive.
func (r *Registry) SearchAsync(query string, resultsCh chan<- []SearchResult, errCh chan<- error, doneCh chan<- struct{}) {
	if query == "" {
		close(resultsCh)
		if doneCh != nil {
			doneCh <- struct{}{}
		}
		return
	}

	var wg sync.WaitGroup
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
				if errCh != nil {
					errCh <- err
				}
				return
			}
			if len(results) > 0 {
				resultsCh <- results
			}
		}(provider)
	}

	go func() {
		wg.Wait()
		close(resultsCh)
		if errCh != nil {
			close(errCh)
		}
		if doneCh != nil {
			doneCh <- struct{}{}
		}
	}()
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
