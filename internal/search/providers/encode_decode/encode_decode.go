package encodedecode

import (
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"log/slog"
	"strings"

	"github.com/MordFustang21/marvin-go/internal/search"
	"github.com/lithammer/fuzzysearch/fuzzy"
	"golang.design/x/clipboard"
)

var _ search.Provider = (*EncoderDecoder)(nil)

// EncoderDecoder is a search provider that handles encoding and decoding tasks.
type EncoderDecoder struct {
	priority int
}

// NewProvider creates a new EncoderDecoder provider.
func NewProvider(priority int) (*EncoderDecoder, error) {
	err := clipboard.Init()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize clipboard: %w", err)
	}

	return &EncoderDecoder{
		priority: priority,
	}, nil
}

// CanHandle implements search.Provider.
func (e *EncoderDecoder) CanHandle(query string) bool {
	canHandle := fuzzy.Match(query, "base64") || fuzzy.Match(query, "hex")
	slog.Debug("CanHandle check",
		slog.String("query", query),
		slog.Bool("result", canHandle))
	return canHandle
}

// Execute implements search.Provider.
func (e *EncoderDecoder) Execute(result search.SearchResult) error {
	if result.Action != nil {
		result.Action()
	}
	return nil
}

// Name implements search.Provider.
func (e *EncoderDecoder) Name() string {
	return "Encoder/Decoder"
}

// Priority implements search.Provider.
func (e *EncoderDecoder) Priority() int {
	return e.priority
}

// Search implements search.Provider.
func (e *EncoderDecoder) Search(query string) ([]search.SearchResult, error) {
	clipText := string(clipboard.Read(clipboard.FmtText))
	slog.Debug("Encode/Decode provider search",
		slog.String("query", query),
		slog.String("clipboardLength", fmt.Sprintf("%d", len(clipText))))

	if len(clipText) == 0 {
		return []search.SearchResult{{
			Title:       "No text in clipboard",
			Description: "Copy some text to encode or decode",
			Type:        search.TypeSystem,
			Path:        "clipboard:empty",
		}}, nil
	}

	var results []search.SearchResult
	// Show potential base64 operations.
	if fuzzy.Match(query, "base64") {
		// Add encode option
		encoded := base64.StdEncoding.EncodeToString([]byte(clipText))
		results = append(results, search.SearchResult{
			Title:       "Base64 Encode",
			Description: fmt.Sprintf("Result: %s", encoded),
			Type:        search.TypeSystem,
			Path:        "base64:encode",
			Action: func() {
				clipboard.Write(clipboard.FmtText, []byte(encoded))
			},
		})

		// Try to decode - clean up the input first
		cleanClipText := strings.TrimSpace(clipText)
		slog.Debug("Attempting base64 decode", slog.String("input", cleanClipText[:min(20, len(cleanClipText))]))

		decoded, err := base64.StdEncoding.DecodeString(cleanClipText)
		if err == nil {
			decodedStr := string(decoded)
			slog.Debug("Base64 decode successful",
				slog.String("result", decodedStr[:min(20, len(decodedStr))]))

			results = append(results, search.SearchResult{
				Title:       "Base64 Decode",
				Description: fmt.Sprintf("Result: %s", decodedStr),
				Type:        search.TypeSystem,
				Path:        "base64:decode",
				Action: func() {
					clipboard.Write(clipboard.FmtText, decoded)
				},
			})
		} else {
			slog.Debug("Base64 decode failed", slog.Any("error", err))
			results = append(results, search.SearchResult{
				Title:       "Base64 Decode Failed",
				Description: "Invalid base64 input in clipboard: " + err.Error(),
				Type:        search.TypeSystem,
				Path:        "base64:decode_failed",
			})
		}
	}

	if fuzzy.Match(query, "hex") {
		// Add encode option
		encoded := hex.EncodeToString([]byte(clipText))
		results = append(results, search.SearchResult{
			Title:       "Hex Encode",
			Description: fmt.Sprintf("Result: %s", encoded),
			Type:        search.TypeSystem,
			Path:        "hex:encode",
			Action: func() {
				clipboard.Write(clipboard.FmtText, []byte(encoded))
			},
		})

		// Try to decode - clean up the input first
		cleanClipText := strings.TrimSpace(clipText)
		slog.Debug("Attempting hex decode", slog.String("input", cleanClipText[:min(20, len(cleanClipText))]))

		decoded, err := hex.DecodeString(cleanClipText)
		if err == nil {
			decodedStr := string(decoded)
			slog.Debug("Hex decode successful",
				slog.String("result", decodedStr[:min(20, len(decodedStr))]))

			results = append(results, search.SearchResult{
				Title:       "Hex Decode",
				Description: fmt.Sprintf("Result: %s", decodedStr),
				Type:        search.TypeSystem,
				Path:        "hex:decode",
				Action: func() {
					clipboard.Write(clipboard.FmtText, decoded)
				},
			})
		} else {
			slog.Debug("Hex decode failed", slog.Any("error", err))
			results = append(results, search.SearchResult{
				Title:       "Hex Decode Failed",
				Description: "Invalid hex input in clipboard: " + err.Error(),
				Type:        search.TypeSystem,
				Path:        "hex:decode_failed",
			})
		}
	}

	slog.Debug("Encode/Decode results", slog.Int("count", len(results)))
	return results, nil
}

// Type implements search.Provider.
func (e *EncoderDecoder) Type() search.ProviderType {
	return search.TypeSystem
}
