package commands

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/theme"
	"github.com/MordFustang21/marvin-go/internal/search"
	"github.com/lithammer/fuzzysearch/fuzzy"
)

// CommandActionType defines the type of action a command can perform
type CommandActionType string

const (
	// ActionTypeShell indicates a command that executes a shell command
	ActionTypeShell CommandActionType = "shell"
	// ActionTypeURL indicates a command that opens a URL
	ActionTypeURL CommandActionType = "url"
	// ActionTypeApplication indicates a command that opens an application
	ActionTypeApplication CommandActionType = "application"
)

// CommandAction defines the action to be performed when a command is executed
type CommandAction struct {
	Type    CommandActionType `json:"type"`
	Command string            `json:"command"`
	Path    string            `json:"path,omitempty"` // For ActionTypeApplication
	URL     string            `json:"url,omitempty"`  // For ActionTypeURL
}

// Command represents a single custom command
type Command struct {
	Name        string        `json:"name"`
	Trigger     string        `json:"trigger"`
	Description string        `json:"description"`
	Action      CommandAction `json:"action"`
	Icon        string        `json:"icon,omitempty"` // Optional path to an icon file
}

// CommandProvider represents a collection of related commands
type CommandProvider struct {
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Icon        string    `json:"icon,omitempty"` // Path to an icon file
	Commands    []Command `json:"commands"`
}

// Provider is a search provider that handles custom user-defined commands
type Provider struct {
	priority         int
	commandFiles     []string                 // Paths to command definition files
	commandProviders []CommandProvider        // Parsed command providers
	commands         map[string][]Command     // All commands indexed by trigger keyword
	icons            map[string]fyne.Resource // Cached icons
	configDir        string                   // Directory containing command definition files
}

// NewProvider creates a new commands provider
func NewProvider(priority int, configDir string) *Provider {
	if configDir == "" {
		// Default to ~/.config/marvin/commands/
		homeDir, err := os.UserHomeDir()
		if err != nil {
			slog.Error("Failed to get home directory", slog.Any("error", err))
			homeDir = "."
		}

		configDir = filepath.Join(homeDir, ".config", "marvin", "commands")
	}

	// Create directory if it doesn't exist
	if err := os.MkdirAll(configDir, 0755); err != nil {
		slog.Error("Failed to create commands directory", slog.String("path", configDir), slog.Any("error", err))
	}

	provider := &Provider{
		priority:  priority,
		commands:  make(map[string][]Command),
		icons:     make(map[string]fyne.Resource),
		configDir: configDir,
	}

	// Load command definitions
	provider.loadCommands()

	return provider
}

// Name returns the provider's name
func (p *Provider) Name() string {
	return "Commands"
}

// Type returns the provider type
func (p *Provider) Type() search.ProviderType {
	return search.TypeSystem
}

// Priority returns the provider's priority
func (p *Provider) Priority() int {
	return p.priority
}

// CanHandle returns whether the provider can handle the given query
func (p *Provider) CanHandle(query string) bool {
	query = strings.ToLower(query)

	// Check if query matches any command trigger
	for trigger := range p.commands {
		if strings.HasPrefix(query, trigger) || strings.HasPrefix(trigger, query) {
			return true
		}
	}

	return false
}

// Search returns custom commands matching the query
func (p *Provider) Search(query string) ([]search.SearchResult, error) {
	query = strings.ToLower(query)
	results := []search.SearchResult{}

	// Find matching commands
	for trigger, cmds := range p.commands {
		if fuzzy.Match(query, trigger) || fuzzy.Match(trigger, query) {
			for _, cmd := range cmds {
				// Create a copy to use in the closure
				command := cmd

				icon := p.getCommandIcon(command)

				results = append(results, search.SearchResult{
					Title:       command.Name,
					Description: command.Description,
					Path:        cmd.Name,
					Icon:        icon,
					Type:        search.TypeSystem,
					Action: func() {
						p.executeCommand(command)
					},
				})
			}
		}
	}

	return results, nil
}

// Execute executes a command
func (p *Provider) Execute(result search.SearchResult) error {
	if result.Type != search.TypeSystem {
		return fmt.Errorf("not a command result")
	}

	if result.Action != nil {
		result.Action()
	}

	return nil
}

// loadCommands loads all command definitions from the config directory
func (p *Provider) loadCommands() {
	// Clear existing commands
	p.commandProviders = []CommandProvider{}
	p.commands = make(map[string][]Command)

	// Find all JSON files in the config directory
	err := filepath.WalkDir(p.configDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if !d.IsDir() && strings.HasSuffix(strings.ToLower(path), ".json") {
			p.commandFiles = append(p.commandFiles, path)

			// Load the command provider from the file
			if err := p.loadCommandProvider(path); err != nil {
				slog.Error("failed to load commands", slog.String("path", path))
			}
		}

		return nil
	})

	if err != nil {
		slog.Error("failed to walk commands directory", slog.String("path", p.configDir), slog.Any("error", err))
	}

	slog.Debug("Loaded command providers", slog.Int("numProviders", len(p.commandProviders)), slog.Int("commands", p.countTotalCommands()))
}

// countTotalCommands returns the total number of commands across all providers
func (p *Provider) countTotalCommands() int {
	count := 0
	for _, provider := range p.commandProviders {
		count += len(provider.Commands)
	}
	return count
}

// loadCommandProvider loads a command provider from a JSON file
func (p *Provider) loadCommandProvider(path string) error {
	// Read the file
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read command file: %w", err)
	}

	// Parse the JSON
	var provider CommandProvider
	if err := json.Unmarshal(data, &provider); err != nil {
		return fmt.Errorf("failed to parse command file: %w", err)
	}

	// Add to our list of providers
	p.commandProviders = append(p.commandProviders, provider)

	// Index commands by trigger
	for _, cmd := range provider.Commands {
		trigger := strings.ToLower(cmd.Trigger)
		p.commands[trigger] = append(p.commands[trigger], cmd)

		// Pre-load icon
		if cmd.Icon != "" {
			p.loadIcon(cmd.Icon)
		} else if provider.Icon != "" {
			// Use provider icon as fallback
			p.loadIcon(provider.Icon)
		}
	}

	return nil
}

// executeCommand executes a command based on its action type
func (p *Provider) executeCommand(cmd Command) {
	switch cmd.Action.Type {
	case ActionTypeShell:
		p.executeShellCommand(cmd.Action.Command)
	case ActionTypeURL:
		url := cmd.Action.URL
		if url == "" {
			url = cmd.Action.Command
		}
		p.openURL(url)
	case ActionTypeApplication:
		path := cmd.Action.Path
		if path == "" {
			path = cmd.Action.Command
		}
		p.openApplication(path)
	default:
		slog.Error("unknown command action type", slog.String("type", string(cmd.Action.Type)))
	}
}

// executeShellCommand executes a shell command
func (p *Provider) executeShellCommand(command string) {
	// Execute the command in the user's shell
	cmd := exec.Command("sh", "-c", command)

	// Run the command without waiting for output
	// For commands that might show UI or take a while
	if err := cmd.Start(); err != nil {
		slog.Error("failed to start shell command", slog.String("command", command), slog.Any("error", err))
		return
	}

	// Optionally wait for completion in a goroutine
	go func() {
		if err := cmd.Wait(); err != nil {
			slog.Error("shell command failed", slog.String("command", command), slog.Any("error", err))
		}
	}()
}

// openURL opens a URL in the default browser
func (p *Provider) openURL(url string) {
	cmd := exec.Command("open", url)
	if err := cmd.Run(); err != nil {
		slog.Error("failed to open URL", slog.String("url", url), slog.Any("error", err))
	}
}

// openApplication opens an application
func (p *Provider) openApplication(path string) {
	cmd := exec.Command("open", path)
	if err := cmd.Run(); err != nil {
		slog.Error("failed to open application", slog.String("path", path), slog.Any("error", err))
	}
}

// getCommandIcon returns an icon for a command
func (p *Provider) getCommandIcon(cmd Command) fyne.Resource {
	// Try command-specific icon first
	if cmd.Icon != "" {
		if icon, exists := p.icons[cmd.Icon]; exists {
			return icon
		}

		// Try to load the icon
		if icon := p.loadIcon(cmd.Icon); icon != nil {
			return icon
		}
	}

	// Based on action type, return an appropriate default icon
	switch cmd.Action.Type {
	case ActionTypeShell:
		return theme.ComputerIcon()
	case ActionTypeURL:
		return theme.SearchIcon()
	case ActionTypeApplication:
		return theme.ComputerIcon()
	default:
		return theme.DocumentIcon()
	}
}

// loadIcon loads an icon from a file
func (p *Provider) loadIcon(path string) fyne.Resource {
	// Check if already loaded
	if icon, exists := p.icons[path]; exists {
		return icon
	}

	// If path is not absolute, make it relative to the config dir
	if !filepath.IsAbs(path) {
		path = filepath.Join(p.configDir, path)
	}

	// Check if file exists
	if _, err := os.Stat(path); err != nil {
		slog.Error("icon file does not exist", slog.String("path", path), slog.Any("error", err))
		return nil
	}

	// Load the icon
	uri := storage.NewFileURI(path)
	res, err := storage.LoadResourceFromURI(uri)
	if err != nil {
		slog.Error("failed to load icon resource", slog.String("path", path), slog.Any("error", err))
		return nil
	}

	// Cache and return the icon
	p.icons[path] = res
	return res
}

// ReloadCommands reloads all command definitions
func (p *Provider) ReloadCommands() {
	p.loadCommands()
}

// GetCommandProviders returns all loaded command providers
func (p *Provider) GetCommandProviders() []CommandProvider {
	return p.commandProviders
}
