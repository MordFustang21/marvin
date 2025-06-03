package search

import (
	"context"
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

// SearchAsync performs a search and sends results as they arrive, ordered by provider priority.
// It accepts a context for cancellation to stop the search when needed.
func (r *Registry) SearchAsync(ctx context.Context, query string, resultsCh chan<- []SearchResult, errCh chan<- error, doneCh chan<- struct{}) {
	if query == "" {
		close(resultsCh)
		if doneCh != nil {
			doneCh <- struct{}{}
			close(doneCh)
		}
		if errCh != nil {
			close(errCh)
		}
		return
	}

	// Reset tracking maps
	r.sentResults = make(map[string]bool)

	// Create a mutex to protect the sentResults map
	var sentMutex sync.Mutex

	// Create channels for collecting results from providers
	type priorityResult struct {
		priority int
		results  []SearchResult
	}
	
	// Channel for results from individual providers
	resultCollector := make(chan priorityResult)
	
	// Use WaitGroup to track when all providers are done
	var wg sync.WaitGroup

	// Tracking active providers
	activeProviders := 0

	// Start provider searches
	for _, provider := range r.providers {
		if !provider.CanHandle(query) {
			continue
		}

		activeProviders++
		priority := provider.Priority()

		wg.Add(1)
		go func(p Provider, prio int) {
			defer wg.Done()
			
			// Check if context is cancelled before starting search
			select {
			case <-ctx.Done():
				return
			default:
				// Continue with search
			}
			
			results, err := p.Search(query)
			
			// Check context again after search
			select {
			case <-ctx.Done():
				return
			default:
				// Continue processing results
			}
			
			if err != nil {
				if errCh != nil {
					select {
					case errCh <- err:
					case <-ctx.Done():
						return
					}
				}
				return
			}

			if len(results) > 0 {
				select {
				case resultCollector <- priorityResult{priority: prio, results: results}:
				case <-ctx.Done():
					return
				}
			}
		}(provider, priority)
	}

	// If no providers are active, clean up and return
	if activeProviders == 0 {
		close(resultsCh)
		if errCh != nil {
			close(errCh)
		}
		if doneCh != nil {
			doneCh <- struct{}{}
			close(doneCh)
		}
		return
	}

	// Start a goroutine to collect results and close channels when done
	go func() {
		// Close the result collector when all providers are done
		go func() {
			wg.Wait()
			close(resultCollector)
		}()

		// Process results as they come in
		for pr := range resultCollector {
			// Filter for unique results
			uniqueResults := make([]SearchResult, 0, len(pr.results))
			
			sentMutex.Lock()
			for _, result := range pr.results {
				// Create a unique key for this result
				resultKey := fmt.Sprintf("%s:%s", string(result.Type), result.Path)
				
				// Only add if we haven't sent this result yet
				if !r.sentResults[resultKey] {
					uniqueResults = append(uniqueResults, result)
					r.sentResults[resultKey] = true
				}
			}
			sentMutex.Unlock()
			
			// Send unique results if we have any
			if len(uniqueResults) > 0 {
				select {
				case resultsCh <- uniqueResults:
				case <-ctx.Done():
					// Cleanup and exit if context is cancelled
					close(resultsCh)
					if errCh != nil {
						close(errCh)
					}
					if doneCh != nil {
						doneCh <- struct{}{}
						close(doneCh)
					}
					return
				}
			}
		}

		// All providers are done, clean up channels
		select {
		case <-ctx.Done():
			// Context was cancelled
		default:
			// Normal completion
			close(resultsCh)
			if errCh != nil {
				close(errCh)
			}
			if doneCh != nil {
				doneCh <- struct{}{}
				close(doneCh)
			}
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
