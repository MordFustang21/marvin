package search

import (
	"fmt"
	"sort"
	"sync"
)

// Registry manages search providers and dispatches search requests
type Registry struct {
	providers []Provider
	// resultsByPriority helps maintain priority order of results
	resultsByPriority map[int][]SearchResult
	// Track which results have been sent to prevent duplicates
	sentResults map[string]bool
}

// NewRegistry creates a new provider registry
func NewRegistry() *Registry {
	return &Registry{
		providers:         []Provider{},
		resultsByPriority: make(map[int][]SearchResult),
		sentResults:       make(map[string]bool),
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

// SearchAsync performs a search and sends results in priority order.
func (r *Registry) SearchAsync(query string, resultsCh chan<- []SearchResult, errCh chan<- error, doneCh chan<- struct{}) {
	if query == "" {
		close(resultsCh)
		if doneCh != nil {
			doneCh <- struct{}{}
		}
		return
	}

	// Reset result collection
	r.resultsByPriority = make(map[int][]SearchResult)
	r.sentResults = make(map[string]bool)

	// Use a mutex to protect the resultsByPriority map during concurrent updates
	var resultsMutex sync.Mutex
	var wg sync.WaitGroup

	// Store provider priorities for sorting
	priorities := make([]int, 0)

	// Tracking active providers
	providerCount := 0

	for _, provider := range r.providers {
		if !provider.CanHandle(query) {
			continue
		}

		providerCount++
		priority := provider.Priority()

		// Track this priority for later sorting
		resultsMutex.Lock()
		if _, exists := r.resultsByPriority[priority]; !exists {
			priorities = append(priorities, priority)
		}
		resultsMutex.Unlock()

		wg.Add(1)
		go func(p Provider, prio int) {
			defer wg.Done()
			results, err := p.Search(query)
			if err != nil {
				if errCh != nil {
					errCh <- err
				}
				return
			}

			if len(results) > 0 {
				// Add results to priority map
				resultsMutex.Lock()
				r.resultsByPriority[prio] = append(r.resultsByPriority[prio], results...)

				// Only send each batch once, prioritizing higher priority providers
				// by not sending immediate results (only in priority order)
				resultsMutex.Unlock()
			}
		}(provider, priority)
	}

	go func() {
		wg.Wait()

		// If we have results, send them in proper priority order
		if len(r.resultsByPriority) > 0 {
			// Sort priorities (lower number = higher priority)
			sort.Ints(priorities)

			// Send results in priority order
			for _, priority := range priorities {
				results := r.resultsByPriority[priority]
				if len(results) > 0 {
					// Create a new slice to hold only unique results
					uniqueResults := make([]SearchResult, 0, len(results))

					// Filter out any duplicates based on Path
					for _, result := range results {
						// Create a unique key for this result
						resultKey := fmt.Sprintf("%s:%s", string(result.Type), result.Path)

						// Only add if we haven't sent this result yet
						if !r.sentResults[resultKey] {
							uniqueResults = append(uniqueResults, result)
							r.sentResults[resultKey] = true
						}
					}

					// Only send if we have unique results to send
					if len(uniqueResults) > 0 {
						resultsCh <- uniqueResults
					}
				}
			}
		}

		// Close channels when done
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
