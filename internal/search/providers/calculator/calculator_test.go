package calculator

import (
	"testing"
)

func TestCalculator(t *testing.T) {
	provider := NewProvider(1)

	tests := []struct {
		expression string
		expected   float64
		shouldErr  bool
	}{
		// Basic arithmetic
		{"2 + 3", 5, false},
		{"10 - 4", 6, false},
		{"3 * 4", 12, false},
		{"15 / 3", 5, false},

		// Decimal numbers
		{"2.5 + 1.5", 4, false},
		{"10.0 / 4.0", 2.5, false},

		// Modulo operator
		{"10 % 3", 1, false},
		{"15 % 4", 3, false},
		{"7.5 % 2.5", 0, false},

		// Exponentiation
		{"2 ^ 3", 8, false},
		{"3 ^ 2", 9, false},
		{"2 ^ 0", 1, false},
		{"4 ^ 0.5", 2, false},

		// Parentheses
		{"(2 + 3) * 4", 20, false},
		{"2 * (3 + 4)", 14, false},
		{"(10 - 2) / (3 + 1)", 2, false},

		// Complex expressions
		{"2 + 3 * 4", 14, false},
		{"(2 + 3) * (4 - 1)", 15, false},
		{"2 ^ 3 + 1", 9, false},
		{"(2 + 3) ^ 2", 25, false},
		
		// Percentage tests
		{"50%", 0.5, false},
		{"100%", 1, false},
		{"25%", 0.25, false},
		{"10 - 50%", 5, false},
		{"20 + 10%", 22, false},
		{"100 * 50%", 50, false},
		
		// More contextual percentage tests
		{"50 + 20%", 60, false},    // 50 + (50 * 0.2) = 50 + 10 = 60
		{"80 - 25%", 60, false},    // 80 - (80 * 0.25) = 80 - 20 = 60
		{"200 + 5%", 210, false},   // 200 + (200 * 0.05) = 200 + 10 = 210
		{"40 - 75%", 10, false},    // 40 - (40 * 0.75) = 40 - 30 = 10

		// Unary operators
		{"-5", -5, false},
		{"+10", 10, false},
		{"-(2 + 3)", -5, false},

		// Nested expressions
		{"((2 + 3) * 2) - 1", 9, false},
		{"2 ^ (3 + 1)", 16, false},

		// Error cases
		{"5 / 0", 0, true},
		{"10 % 0", 0, true},
		{"2 +", 0, true},
		{"* 3", 0, true},
		{"(2 + 3", 0, true},
		{"2 + 3)", 0, true},
	}

	for _, test := range tests {
		t.Run(test.expression, func(t *testing.T) {
			result, err := provider.parseAndCalculate(test.expression)

			if test.shouldErr {
				if err == nil {
					t.Errorf("Expected error for expression %s, but got result: %f", test.expression, result)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error for expression %s: %v", test.expression, err)
					return
				}

				if result != test.expected {
					t.Errorf("Expression %s: expected %f, got %f", test.expression, test.expected, result)
				}
			}
		})
	}
}

func TestCanHandle(t *testing.T) {
	provider := NewProvider(1)

	tests := []struct {
		query    string
		expected bool
	}{
		{"2 + 3", true},
		{"(10 - 5) * 2", true},
		{"2.5 / 1.2", true},
		{"2 ^ 3", true},
		{"10 % 3", true},
		{"hello world", false},
		{"2 + abc", false},
		{"", false},
		{"   ", false},
		{"2+3*4", true},       // No spaces
		{"(2+3)*(4-1)", true}, // No spaces with parentheses
		{"50%", true},         // Percentage
		{"10-50%", true},      // Expression with percentage
	}

	for _, test := range tests {
		t.Run(test.query, func(t *testing.T) {
			result := provider.CanHandle(test.query)
			if result != test.expected {
				t.Errorf("CanHandle(%s): expected %t, got %t", test.query, test.expected, result)
			}
		})
	}
}

func TestOperatorPrecedence(t *testing.T) {
	provider := NewProvider(1)

	tests := []struct {
		expression string
		expected   float64
	}{
		// Multiplication before addition
		{"2 + 3 * 4", 14}, // Should be 2 + (3 * 4) = 14, not (2 + 3) * 4 = 20

		// Division before subtraction
		{"10 - 6 / 2", 7}, // Should be 10 - (6 / 2) = 7, not (10 - 6) / 2 = 2

		// Exponentiation before multiplication
		{"2 * 3 ^ 2", 18}, // Should be 2 * (3 ^ 2) = 18, not (2 * 3) ^ 2 = 36

		// Right associativity of exponentiation
		{"2 ^ 3 ^ 2", 512}, // Should be 2 ^ (3 ^ 2) = 2 ^ 9 = 512, not (2 ^ 3) ^ 2 = 64

		// Modulo same precedence as multiplication
		{"10 + 6 % 4", 12}, // Should be 10 + (6 % 4) = 12, not (10 + 6) % 4 = 0
	}

	for _, test := range tests {
		t.Run(test.expression, func(t *testing.T) {
			result, err := provider.parseAndCalculate(test.expression)
			if err != nil {
				t.Errorf("Unexpected error for expression %s: %v", test.expression, err)
				return
			}

			if result != test.expected {
				t.Errorf("Expression %s: expected %f, got %f", test.expression, test.expected, result)
			}
		})
	}
}
