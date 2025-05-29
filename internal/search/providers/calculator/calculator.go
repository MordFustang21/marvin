package calculator

import (
	"fmt"
	"math"
	"regexp"
	"strings"

	"fyne.io/fyne/v2/theme"
	"github.com/MordFustang21/marvin-go/internal/search"
	"github.com/MordFustang21/marvin-go/internal/util"
	"github.com/mnogu/go-calculator"
)

// Provider is a search provider that can handle basic calculations
type Provider struct {
	priority int
}

// NewProvider creates a new calculator provider
func NewProvider(priority int) *Provider {
	return &Provider{
		priority: priority,
	}
}

// Name returns the provider's name
func (p *Provider) Name() string {
	return "Calculator"
}

// Type returns the provider type
func (p *Provider) Type() search.ProviderType {
	return search.TypeCalculator
}

// Priority returns the provider's priority
func (p *Provider) Priority() int {
	return p.priority
}

// checkIfMathExpression checks if the query is likely a math expression
func (p *Provider) checkIfMathExpression(query string) bool {
	// A simple regex to check for common math operators and digits
	// This will match expressions like "1 + 2", "3.14 * 2", "5 - 3", etc.
	re := regexp.MustCompile(`^[\d\s\+\-\*/\(\)\.\^]+$`)
	return re.MatchString(query)
}

// CanHandle returns whether the provider can handle the given query
func (p *Provider) CanHandle(query string) bool {
	return p.checkIfMathExpression(query)
}

// parseAndCalculate parses and evaluates a mathematical expression
// using a recursive descent parser approach for more complex expressions
func (p *Provider) parseAndCalculate(expr string) (float64, error) {
	res, err := calculator.Calculate(expr)
	if err != nil {
		return 0, fmt.Errorf("invalid expression: %s", err.Error())
	}

	return res, nil
}

// isDigit checks if a character is a digit
func isDigit(c byte) bool {
	return c >= '0' && c <= '9'
}

// Search performs a calculation for the given query and returns results
func (p *Provider) Search(query string) ([]search.SearchResult, error) {
	// Remove any spaces to standardize input
	query = strings.TrimSpace(query)

	// Check if we can calculate it
	if !p.CanHandle(query) {
		return nil, nil
	}

	// Calculate the result
	result, err := p.parseAndCalculate(query)
	if err != nil {
		// Return the error as a result so the user can see it
		return []search.SearchResult{
			{
				Title:       fmt.Sprintf("Error: %s", err.Error()),
				Description: "Could not calculate the expression",
				Icon:        theme.ErrorIcon(),
				Type:        search.TypeCalculator,
				Action: func() {
					// Do nothing on error
				},
			},
		}, nil
	}

	// Format the result
	var resultStr string
	if result == float64(int(result)) {
		// If it's a whole number, format without decimal part
		resultStr = fmt.Sprintf("%d", int(result))
	} else {
		// If the decimal part is small, show more precision
		absResult := math.Abs(result)
		if absResult < 0.1 {
			resultStr = fmt.Sprintf("%.6f", result)
			// Remove trailing zeros
			resultStr = strings.TrimRight(strings.TrimRight(resultStr, "0"), ".")
		} else {
			resultStr = fmt.Sprintf("%.4f", result)
			// Remove trailing zeros
			resultStr = strings.TrimRight(strings.TrimRight(resultStr, "0"), ".")
		}
	}

	// Create a search result
	displayText := fmt.Sprintf("%s = %s", query, resultStr)

	return []search.SearchResult{
		{
			Title:       displayText,
			Description: "Calculator result - Press Enter to copy to clipboard",
			Icon:        theme.ContentAddIcon(),
			Type:        search.TypeCalculator,
			Action: func() {
				// Copy the result to clipboard
				if err := util.CopyToClipboard(resultStr); err != nil {
					fmt.Printf("Failed to copy to clipboard: %v\n", err)
				} else {
					fmt.Printf("Copied to clipboard: %s\n", resultStr)
				}
			},
		},
	}, nil
}

// Execute triggers an action for the given result
func (p *Provider) Execute(result search.SearchResult) error {
	if result.Type != search.TypeCalculator {
		return fmt.Errorf("not a calculator result")
	}

	if result.Action != nil {
		result.Action()
	}

	return nil
}
