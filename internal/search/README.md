# Marvin Search Architecture

The search functionality in Marvin is built around a flexible and extensible architecture consisting of providers, results, and a registry. This design allows Marvin to incorporate results from a variety of sources and display them in a cohesive, prioritized manner.

## Core Components

### Provider Interface

The `Provider` interface is the foundation of Marvin's search functionality. Each provider is responsible for handling a specific type of search and returning standardized results. Providers implement the following interface:

```go
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
```

### Search Result

All search providers return results in a standardized format through the `SearchResult` struct:

```go
type SearchResult struct {
    Title       string        // Display title for the result
    Description string        // Secondary description text
    Path        string        // Path or identifier (if applicable)
    Icon        fyne.Resource // Icon to display with the result
    Type        ProviderType  // Type of the provider that generated this result
    Action      func()        // Function to execute when the result is selected
}
```

### Registry

The `Registry` manages multiple providers and coordinates the search process. It:

- Maintains a list of registered providers
- Sorts providers by priority
- Dispatches searches to relevant providers based on the query
- Collects, deduplicates, and orders results from all providers
- Provides an asynchronous search API through channels

## Provider Types

Marvin supports several types of providers:

- `TypeFile`: For file and application searches (e.g., Spotlight provider)
- `TypeCalculator`: For mathematical calculations
- `TypeWeb`: For web searches and URL handling
- `TypeSystem`: For system commands and actions
- `TypeApp`: For application-specific searches

## Included Providers

### Spotlight Provider

The Spotlight provider leverages macOS's built-in Spotlight search index to find applications and files. It:

- Caches applications for faster results
- Extracts rich metadata from found items
- Handles launching applications and opening files

### Calculator Provider

The Calculator provider performs mathematical calculations directly in the search bar. It:

- Detects and evaluates mathematical expressions
- Formats and displays calculation results
- Allows copying results to the clipboard

### Web Provider

The Web provider handles URL opening and web searches. It:

- Detects direct URLs for opening
- Provides web search functionality with configurable search engines
- Formats web search results

### Commands Provider

The Commands provider enables custom user-defined commands and shortcuts. It:

- Loads command definitions from JSON files
- Supports multiple command types (shell, URL, application)
- Allows organization of commands into logical groups
- Handles custom icons for commands

### System Settings Provider

The System Settings provider enables quick access to macOS system settings panels. It:

- Provides comprehensive coverage of macOS System Settings/System Preferences
- Uses URL schemes to directly open specific settings panels
- Supports fuzzy search through setting names, descriptions, and keywords
- Works with both legacy System Preferences and newer System Settings
- Categorizes settings for better organization and discoverability

## Implementing a Custom Provider

To implement a custom search provider:

1. Create a new package under `internal/search/providers/`
2. Implement the `Provider` interface
3. Register the provider in `cmd/marvin/main.go`

### Example Provider Implementation

Here's a minimal example of a custom provider:

```go
package example

import (
    "github.com/MordFustang21/marvin-go/internal/search"
    "fyne.io/fyne/v2/theme"
)

type Provider struct {
    priority int
}

func NewProvider(priority int) *Provider {
    return &Provider{priority: priority}
}

func (p *Provider) Name() string {
    return "Example"
}

func (p *Provider) Type() search.ProviderType {
    return search.TypeSystem
}

func (p *Provider) Priority() int {
    return p.priority
}

func (p *Provider) CanHandle(query string) bool {
    return query == "example"
}

func (p *Provider) Search(query string) ([]search.SearchResult, error) {
    result := search.SearchResult{
        Title:       "Example Result",
        Description: "This is an example search result",
        Icon:        theme.InfoIcon(),
        Type:        search.TypeSystem,
        Action:      func() { fmt.Println("Example action executed") },
    }
    return []search.SearchResult{result}, nil
}

func (p *Provider) Execute(result search.SearchResult) error {
    if result.Action != nil {
        result.Action()
    }
    return nil
}
```

### Registration

Register your provider in `cmd/marvin/main.go`:

```go
func setupSearchProviders(registry *search.Registry) {
    // Existing providers
    
    // Register your new provider
    exampleProvider := example.NewProvider(5)
    registry.RegisterProvider(exampleProvider)
}
```

## Search Flow

1. User types a query in the search bar
2. UI triggers a search through the registry with a small delay for debouncing
3. Registry checks which providers can handle the query via `CanHandle()`
4. Registry dispatches the search to applicable providers in parallel
5. Providers return results asynchronously through channels
6. Registry collects, deduplicates, and sorts results by provider priority
7. UI receives and displays results in batches as they become available
8. When a result is selected, its associated action is executed

## Best Practices

When creating a provider:

1. **Respect Priority**: Use appropriate priority values to ensure results appear in the correct order
2. **Optimize Performance**: Keep search operations fast, especially for frequently used providers
3. **Be Selective**: The `CanHandle()` method should efficiently filter queries that your provider can't handle
4. **Clear Descriptions**: Provide informative title and description for each result
5. **Meaningful Icons**: Choose appropriate icons to help users quickly identify result types
6. **Error Handling**: Handle errors gracefully and provide informative error messages
7. **Resource Management**: Clean up resources properly when they're no longer needed

## Future Directions

Potential extensions to the search architecture could include:

- User-configurable provider priorities
- Learning from user selections to improve result ranking
- More sophisticated result filtering and categorization
- Plugin system for third-party providers