package web

import (
	"fmt"
	"net/url"
	"os/exec"
	"regexp"
	"strings"

	"fyne.io/fyne/v2/theme"
	"github.com/MordFustang21/marvin-go/internal/search"
)

// Provider is a search provider that handles web searches and direct URL opening
type Provider struct {
	priority    int
	searchURL   string // URL template for searches (with %s for query)
	directRegex *regexp.Regexp
}

// NewProvider creates a new web provider
func NewProvider(priority int) *Provider {
	// Default to Google search
	urlRegex := regexp.MustCompile(`^(https?:\/\/)?([\w\-]+(\.[\w\-]+)+)(\/[^\s]*)?$`)

	return &Provider{
		priority:    priority,
		searchURL:   "https://www.google.com/search?q=%s",
		directRegex: urlRegex,
	}
}

// Name returns the provider's name
func (p *Provider) Name() string {
	return "Web"
}

// Type returns the provider type
func (p *Provider) Type() search.ProviderType {
	return search.TypeWeb
}

// Priority returns the provider's priority
func (p *Provider) Priority() int {
	return p.priority
}

// CanHandle returns whether the provider can handle the given query
func (p *Provider) CanHandle(query string) bool {
	// Check if it's a URL (containing domain)
	if p.directRegex.MatchString(query) {
		return true
	}

	// Only handle queries that look like web searches
	// Check for web search indicators like "search", "how to", etc.
	webSearchIndicators := []string{"search", "how to", "what is", "where", "when", "who", "why", ".com", ".net", ".org"}
	for _, indicator := range webSearchIndicators {
		if strings.Contains(strings.ToLower(query), indicator) {
			return true
		}
	}

	// Otherwise, be more selective than the default providers
	return false
}

// prepareURL formats the URL for opening, ensuring it has a protocol
func (p *Provider) prepareURL(query string) string {
	// Check if it's already a URL
	if p.directRegex.MatchString(query) {
		// Make sure it has the http:// prefix
		if !strings.HasPrefix(query, "http://") && !strings.HasPrefix(query, "https://") {
			return "https://" + query
		}
		return query
	}

	// Otherwise, use the search URL template
	return fmt.Sprintf(p.searchURL, url.QueryEscape(query))
}

// Search performs a web-related search with the given query
func (p *Provider) Search(query string) ([]search.SearchResult, error) {
	if query == "" {
		return []search.SearchResult{}, nil
	}

	results := []search.SearchResult{}

	// First, check if this is directly a URL
	if p.directRegex.MatchString(query) {
		urlToOpen := p.prepareURL(query)
		results = append(results, search.SearchResult{
			Title:       "Open URL: " + query,
			Description: "Open in default browser",
			Path:        urlToOpen,
			Icon:        theme.ComputerIcon(),
			Type:        search.TypeWeb,
			Action: func() {
				p.openURL(urlToOpen)
			},
		})
		return results, nil
	}

	// Add a web search result
	urlToOpen := p.prepareURL(query)
	results = append(results, search.SearchResult{
		Title:       "Search: " + query,
		Description: "Search on Google",
		Path:        urlToOpen,
		Icon:        theme.SearchIcon(),
		Type:        search.TypeWeb,
		Action: func() {
			p.openURL(urlToOpen)
		},
	})

	return results, nil
}

// openURL opens the given URL in the default browser
func (p *Provider) openURL(url string) error {
	cmd := exec.Command("open", url)
	return cmd.Run()
}

// Execute triggers an action for the given result
func (p *Provider) Execute(result search.SearchResult) error {
	if result.Type != search.TypeWeb {
		return fmt.Errorf("not a web result")
	}

	if result.Action != nil {
		result.Action()
	}

	return nil
}

// SetSearchEngine allows changing the search engine used
func (p *Provider) SetSearchEngine(name string) {
	switch strings.ToLower(name) {
	case "google":
		p.searchURL = "https://www.google.com/search?q=%s"
	case "bing":
		p.searchURL = "https://www.bing.com/search?q=%s"
	case "duckduckgo", "ddg":
		p.searchURL = "https://duckduckgo.com/?q=%s"
	case "yahoo":
		p.searchURL = "https://search.yahoo.com/search?p=%s"
	default:
		// If unrecognized, stick with the default
	}
}
