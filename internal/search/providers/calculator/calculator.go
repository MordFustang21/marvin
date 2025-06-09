package calculator

import (
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"

	"fyne.io/fyne/v2/theme"
	"github.com/MordFustang21/marvin-go/internal/search"
	"github.com/MordFustang21/marvin-go/internal/util"
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
	// Trim whitespace and check if empty
	trimmed := strings.TrimSpace(query)
	if trimmed == "" {
		return false
	}

	// A simple regex to check for common math operators and digits
	// This will match expressions like "1 + 2", "3.14 * 2", "5 - 3", "50%", etc.
	re := regexp.MustCompile(`^[\d\s\+\-\*\%/\(\)\.\^]+$`)
	return re.MatchString(query)
}

// CanHandle returns whether the provider can handle the given query
func (p *Provider) CanHandle(query string) bool {
	return p.checkIfMathExpression(query)
}

// tokenType represents the type of a token
type tokenType int

const (
	tokenNumber tokenType = iota
	tokenOperator
	tokenLeftParen
	tokenRightParen
	tokenPercent
	tokenEnd
)

// token represents a token in the expression
type token struct {
	typ   tokenType
	value string
}

// tokenizer holds the state for tokenizing an expression
type tokenizer struct {
	expr     string
	position int
	current  rune
}

// newTokenizer creates a new tokenizer for the given expression
func newTokenizer(expr string) *tokenizer {
	t := &tokenizer{expr: expr, position: 0}
	if len(expr) > 0 {
		t.current = rune(expr[0])
	}
	return t
}

// advance moves to the next character
func (t *tokenizer) advance() {
	t.position++
	if t.position >= len(t.expr) {
		t.current = 0 // EOF
	} else {
		t.current = rune(t.expr[t.position])
	}
}

// skipWhitespace skips any whitespace characters
func (t *tokenizer) skipWhitespace() {
	for t.current == ' ' || t.current == '\t' {
		t.advance()
	}
}

// readNumber reads a number (integer or decimal) and checks for percentage
func (t *tokenizer) readNumber() (string, bool) {
	var result strings.Builder
	for (t.current >= '0' && t.current <= '9') || t.current == '.' {
		result.WriteRune(t.current)
		t.advance()
	}

	// Check if followed by %
	isPercent := t.current == '%'
	if isPercent {
		t.advance()
	}

	return result.String(), isPercent
}

// nextToken returns the next token from the expression
func (t *tokenizer) nextToken() token {
	for t.current != 0 {
		t.skipWhitespace()

		if t.current == 0 {
			break
		}

		// Numbers (and percentages)
		if (t.current >= '0' && t.current <= '9') || t.current == '.' {
			numStr, isPercent := t.readNumber()
			if isPercent {
				return token{tokenPercent, numStr}
			}
			return token{tokenNumber, numStr}
		}

		// Operators
		if t.current == '+' || t.current == '-' || t.current == '*' ||
			t.current == '/' || t.current == '%' || t.current == '^' {
			op := string(t.current)
			t.advance()
			return token{tokenOperator, op}
		}

		// Parentheses
		if t.current == '(' {
			t.advance()
			return token{tokenLeftParen, "("}
		}

		if t.current == ')' {
			t.advance()
			return token{tokenRightParen, ")"}
		}

		// Unknown character
		t.advance()
	}

	return token{tokenEnd, ""}
}

// parser holds the state for parsing an expression
type parser struct {
	tokenizer *tokenizer
	current   token
}

// newParser creates a new parser for the given expression
func newParser(expr string) *parser {
	p := &parser{tokenizer: newTokenizer(expr)}
	p.current = p.tokenizer.nextToken()
	return p
}

// eat consumes a token of the expected type
func (p *parser) eat(expectedType tokenType) error {
	if p.current.typ == expectedType {
		p.current = p.tokenizer.nextToken()
		return nil
	}
	return fmt.Errorf("expected token type %d, got %d", expectedType, p.current.typ)
}

// factor parses a factor (number or parenthesized expression)
func (p *parser) factor() (float64, error) {
	token := p.current

	if token.typ == tokenNumber {
		if err := p.eat(tokenNumber); err != nil {
			return 0, err
		}
		value, err := strconv.ParseFloat(token.value, 64)
		if err != nil {
			return 0, fmt.Errorf("invalid number: %s", token.value)
		}
		return value, nil
	}

	if token.typ == tokenPercent {
		if err := p.eat(tokenPercent); err != nil {
			return 0, err
		}
		value, err := strconv.ParseFloat(token.value, 64)
		if err != nil {
			return 0, fmt.Errorf("invalid percentage: %s", token.value)
		}
		return value / 100.0, nil
	}

	if token.typ == tokenLeftParen {
		if err := p.eat(tokenLeftParen); err != nil {
			return 0, err
		}
		result, err := p.expr()
		if err != nil {
			return 0, err
		}
		if err := p.eat(tokenRightParen); err != nil {
			return 0, err
		}
		return result, nil
	}

	// Handle unary minus
	if token.typ == tokenOperator && token.value == "-" {
		if err := p.eat(tokenOperator); err != nil {
			return 0, err
		}
		factor, err := p.factor()
		if err != nil {
			return 0, err
		}
		return -factor, nil
	}

	// Handle unary plus
	if token.typ == tokenOperator && token.value == "+" {
		if err := p.eat(tokenOperator); err != nil {
			return 0, err
		}
		return p.factor()
	}

	return 0, fmt.Errorf("unexpected token: %s", token.value)
}

// power parses exponentiation (right associative)
func (p *parser) power() (float64, error) {
	result, err := p.factor()
	if err != nil {
		return 0, err
	}

	if p.current.typ == tokenOperator && p.current.value == "^" {
		if err := p.eat(tokenOperator); err != nil {
			return 0, err
		}
		right, err := p.power() // Right associative
		if err != nil {
			return 0, err
		}
		result = math.Pow(result, right)
	}

	return result, nil
}

// term parses multiplication, division, and modulo
func (p *parser) term() (float64, error) {
	result, err := p.power()
	if err != nil {
		return 0, err
	}

	for p.current.typ == tokenOperator &&
		(p.current.value == "*" || p.current.value == "/" || p.current.value == "%") {

		op := p.current.value
		if err := p.eat(tokenOperator); err != nil {
			return 0, err
		}

		right, err := p.power()
		if err != nil {
			return 0, err
		}

		switch op {
		case "*":
			result *= right
		case "/":
			if right == 0 {
				return 0, fmt.Errorf("division by zero")
			}
			result /= right
		case "%":
			if right == 0 {
				return 0, fmt.Errorf("modulo by zero")
			}
			result = math.Mod(result, right)
		}
	}

	return result, nil
}

// expr parses addition and subtraction
func (p *parser) expr() (float64, error) {
	result, err := p.term()
	if err != nil {
		return 0, err
	}

	for p.current.typ == tokenOperator &&
		(p.current.value == "+" || p.current.value == "-") {

		op := p.current.value
		if err := p.eat(tokenOperator); err != nil {
			return 0, err
		}

		// Check if the next token is a percentage for contextual calculation
		if p.current.typ == tokenPercent {
			percentToken := p.current
			if err := p.eat(tokenPercent); err != nil {
				return 0, err
			}

			percentValue, err := strconv.ParseFloat(percentToken.value, 64)
			if err != nil {
				return 0, fmt.Errorf("invalid percentage: %s", percentToken.value)
			}

			// Apply percentage to the left operand (result)
			percentageAmount := result * (percentValue / 100.0)

			if op == "+" {
				result += percentageAmount
			} else {
				result -= percentageAmount
			}
		} else {
			right, err := p.term()
			if err != nil {
				return 0, err
			}

			if op == "+" {
				result += right
			} else {
				result -= right
			}
		}
	}

	return result, nil
}

// parseAndCalculate parses and evaluates a mathematical expression
// using a recursive descent parser approach for more complex expressions
func (p *Provider) parseAndCalculate(expr string) (float64, error) {
	parser := newParser(expr)
	result, err := parser.expr()
	if err != nil {
		return 0, err
	}

	// Check if we've consumed all tokens
	if parser.current.typ != tokenEnd {
		return 0, fmt.Errorf("unexpected token at end: %s", parser.current.value)
	}

	return result, nil
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
			Description: "Press Enter to copy result to clipboard",
			Icon:        theme.ContentAddIcon(),
			Type:        search.TypeCalculator,
			Action: func() {
				// Copy the result to clipboard
				if err := util.CopyToClipboard(resultStr); err != nil {
					fmt.Printf("Failed to copy to clipboard: %v\n", err)
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
