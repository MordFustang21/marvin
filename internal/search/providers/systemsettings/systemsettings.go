package systemsettings

import (
	"fmt"
	"os/exec"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
	"github.com/MordFustang21/marvin-go/internal/search"
	"github.com/MordFustang21/marvin-go/internal/ui/icons"
)

// SystemSetting represents a macOS system setting panel
type SystemSetting struct {
	Name        string   // Display name
	Description string   // Description of what this setting controls
	URLScheme   string   // URL scheme to open the setting
	Keywords    []string // Additional keywords for search
	Category    string   // Category (General, Privacy, etc.)
}

// Provider is a search provider for macOS System Settings
type Provider struct {
	priority           int
	settings           []SystemSetting
	systemSettingsIcon fyne.Resource
}

// NewProvider creates a new System Settings provider
func NewProvider(priority int) *Provider {
	provider := &Provider{
		priority: priority,
		settings: getSystemSettings(),
	}

	// Try to get the actual System Settings app icon
	provider.systemSettingsIcon = provider.getSystemSettingsIcon()

	return provider
}

// Name returns the provider's name
func (p *Provider) Name() string {
	return "System Settings"
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
	if len(query) < 2 {
		return false
	}

	query = strings.ToLower(query)

	// Only check setting names for faster performance
	for _, setting := range p.settings {
		settingName := strings.ToLower(setting.Name)
		if strings.Contains(settingName, query) {
			return true
		}
	}

	return false
}

// Search performs a search for system settings
func (p *Provider) Search(query string) ([]search.SearchResult, error) {
	if query == "" {
		return []search.SearchResult{}, nil
	}

	query = strings.ToLower(query)
	results := []search.SearchResult{}

	// Find matching settings - only check names for speed, limit results
	for _, setting := range p.settings {
		// Limit results for better performance
		if len(results) >= 5 {
			break
		}

		settingName := strings.ToLower(setting.Name)

		// Use contains matching for better discoverability
		if strings.Contains(settingName, query) {
			// Create a copy for the closure
			settingCopy := setting

			results = append(results, search.SearchResult{
				Title:       setting.Name,
				Description: fmt.Sprintf("%s • %s", setting.Category, setting.Description),
				Path:        setting.URLScheme,
				Icon:        p.systemSettingsIcon,
				Type:        search.TypeSystem,
				Action: func() {
					p.openSystemSetting(settingCopy.URLScheme)
				},
			})
		}
	}

	return results, nil
}

// Execute triggers an action for the given result
func (p *Provider) Execute(result search.SearchResult) error {
	if result.Type != search.TypeSystem {
		return fmt.Errorf("not a system settings result")
	}

	if result.Action != nil {
		result.Action()
	}

	return nil
}

// openSystemSetting opens a system setting using its URL scheme
func (p *Provider) openSystemSetting(urlScheme string) error {
	cmd := exec.Command("open", urlScheme)
	return cmd.Run()
}

// getSystemSettingsIcon tries to get the actual System Settings app icon
func (p *Provider) getSystemSettingsIcon() fyne.Resource {
	// Try to get the System Settings app icon (macOS Ventura+)
	systemSettingsPath := "/System/Applications/System Settings.app"
	if icon := icons.GetAppIcon(systemSettingsPath); icon != nil && icon != theme.ComputerIcon() {
		return icon
	}

	// Fallback to System Preferences app icon (macOS Monterey and earlier)
	systemPrefsPath := "/System/Applications/System Preferences.app"
	if icon := icons.GetAppIcon(systemPrefsPath); icon != nil && icon != theme.ComputerIcon() {
		return icon
	}

	// Final fallback to the theme settings icon
	return theme.SettingsIcon()
}

// getSystemSettings returns the comprehensive list of macOS system settings
// Ordered by most commonly searched items first for better performance
func getSystemSettings() []SystemSetting {
	return []SystemSetting{
		// Most commonly searched settings first
		{
			Name:        "Wi-Fi",
			Description: "Wireless network settings",
			URLScheme:   "x-apple.systempreferences:com.apple.preference.network?Wi-Fi",
			Keywords:    []string{"wifi", "wireless", "network", "internet"},
			Category:    "Network",
		},
		{
			Name:        "Bluetooth",
			Description: "Bluetooth device connections",
			URLScheme:   "x-apple.systempreferences:com.apple.preference.bluetooth",
			Keywords:    []string{"bluetooth", "wireless", "devices", "pairing"},
			Category:    "Network",
		},
		{
			Name:        "Sound",
			Description: "Audio input and output settings",
			URLScheme:   "x-apple.systempreferences:com.apple.preference.sound",
			Keywords:    []string{"sound", "audio", "volume", "input", "output", "speakers", "headphones"},
			Category:    "Hardware",
		},
		{
			Name:        "Displays",
			Description: "Display resolution and arrangement",
			URLScheme:   "x-apple.systempreferences:com.apple.preference.displays",
			Keywords:    []string{"display", "monitor", "resolution", "arrangement", "brightness"},
			Category:    "Hardware",
		},
		{
			Name:        "Keyboard",
			Description: "Keyboard settings and shortcuts",
			URLScheme:   "x-apple.systempreferences:com.apple.preference.keyboard",
			Keywords:    []string{"keyboard", "shortcuts", "text replacement", "input sources"},
			Category:    "Input",
		},
		{
			Name:        "Trackpad",
			Description: "Trackpad gestures and settings",
			URLScheme:   "x-apple.systempreferences:com.apple.preference.trackpad",
			Keywords:    []string{"trackpad", "gestures", "tap", "scroll", "zoom"},
			Category:    "Input",
		},
		{
			Name:        "Mouse",
			Description: "Mouse settings and gestures",
			URLScheme:   "x-apple.systempreferences:com.apple.preference.mouse",
			Keywords:    []string{"mouse", "pointer", "scrolling", "clicking"},
			Category:    "Input",
		},
		{
			Name:        "Privacy & Security",
			Description: "Privacy and security settings",
			URLScheme:   "x-apple.systempreferences:com.apple.preference.security",
			Keywords:    []string{"privacy", "security", "firewall", "filevault", "gatekeeper"},
			Category:    "Privacy & Security",
		},
		{
			Name:        "Desktop & Dock",
			Description: "Desktop background and Dock settings",
			URLScheme:   "x-apple.systempreferences:com.apple.preference.dock",
			Keywords:    []string{"desktop", "wallpaper", "background", "dock", "magnification"},
			Category:    "Appearance",
		},
		{
			Name:        "General",
			Description: "General system settings and appearance",
			URLScheme:   "x-apple.systempreferences:com.apple.preference.general",
			Keywords:    []string{"appearance", "accent", "highlight", "sidebar", "menubar"},
			Category:    "General",
		},

		// Network
		{
			Name:        "Network",
			Description: "Network and internet settings",
			URLScheme:   "x-apple.systempreferences:com.apple.preference.network",
			Keywords:    []string{"network", "ethernet", "vpn", "proxy"},
			Category:    "Network",
		},

		// Privacy & Security detailed settings
		{
			Name:        "Location Services",
			Description: "Location access for apps and services",
			URLScheme:   "x-apple.systempreferences:com.apple.preference.security?Privacy_LocationServices",
			Keywords:    []string{"location", "gps", "maps"},
			Category:    "Privacy & Security",
		},
		{
			Name:        "Camera",
			Description: "Camera access permissions",
			URLScheme:   "x-apple.systempreferences:com.apple.preference.security?Privacy_Camera",
			Keywords:    []string{"camera", "webcam", "video"},
			Category:    "Privacy & Security",
		},
		{
			Name:        "Microphone",
			Description: "Microphone access permissions",
			URLScheme:   "x-apple.systempreferences:com.apple.preference.security?Privacy_Microphone",
			Keywords:    []string{"microphone", "mic", "audio", "recording"},
			Category:    "Privacy & Security",
		},
		{
			Name:        "Full Disk Access",
			Description: "Full disk access permissions",
			URLScheme:   "x-apple.systempreferences:com.apple.preference.security?Privacy_AllFiles",
			Keywords:    []string{"full disk access", "files", "documents"},
			Category:    "Privacy & Security",
		},

		// Other commonly accessed settings
		{
			Name:        "Focus",
			Description: "Do Not Disturb and Focus modes",
			URLScheme:   "x-apple.systempreferences:com.apple.preference.notifications",
			Keywords:    []string{"focus", "do not disturb", "dnd", "notifications"},
			Category:    "Focus",
		},
		{
			Name:        "Mission Control",
			Description: "Spaces and window management",
			URLScheme:   "x-apple.systempreferences:com.apple.preference.expose",
			Keywords:    []string{"spaces", "expose", "dashboard", "hot corners"},
			Category:    "Desktop",
		},
		{
			Name:        "Control Center",
			Description: "Control Center and menu bar settings",
			URLScheme:   "x-apple.systempreferences:com.apple.preference.controlcenter",
			Keywords:    []string{"control center", "menu bar", "widgets"},
			Category:    "Control Center",
		},
		{
			Name:        "Night Shift",
			Description: "Blue light filter settings",
			URLScheme:   "x-apple.systempreferences:com.apple.preference.displays?Night%20Shift",
			Keywords:    []string{"night shift", "blue light", "warm", "schedule"},
			Category:    "Hardware",
		},
		{
			Name:        "Text Input",
			Description: "Text input and language settings",
			URLScheme:   "x-apple.systempreferences:com.apple.preference.keyboard?Text",
			Keywords:    []string{"text input", "autocorrect", "spelling", "predictive"},
			Category:    "Input",
		},

		// Less commonly accessed settings
		{
			Name:        "Screen Time",
			Description: "App usage and screen time controls",
			URLScheme:   "x-apple.systempreferences:com.apple.preference.screentime",
			Keywords:    []string{"screen time", "app limits", "downtime", "usage"},
			Category:    "Screen Time",
		},
		{
			Name:        "General Storage",
			Description: "Storage usage and optimization",
			URLScheme:   "x-apple.systempreferences:com.apple.preference.storage",
			Keywords:    []string{"storage", "disk", "space", "cleanup", "optimize"},
			Category:    "Storage",
		},
		{
			Name:        "Apple ID",
			Description: "Apple ID and iCloud settings",
			URLScheme:   "x-apple.systempreferences:com.apple.preferences.AppleIDPrefPane",
			Keywords:    []string{"apple id", "icloud", "account", "sync"},
			Category:    "Apple ID",
		},
		{
			Name:        "Users & Groups",
			Description: "User accounts and login settings",
			URLScheme:   "x-apple.systempreferences:com.apple.preferences.users",
			Keywords:    []string{"users", "accounts", "login", "password", "groups"},
			Category:    "Users & Groups",
		},
		{
			Name:        "Accessibility",
			Description: "Accessibility features and settings",
			URLScheme:   "x-apple.systempreferences:com.apple.preference.universalaccess",
			Keywords:    []string{"accessibility", "voiceover", "zoom", "contrast", "assistive"},
			Category:    "Accessibility",
		},
		{
			Name:        "Siri & Spotlight",
			Description: "Siri and Spotlight search settings",
			URLScheme:   "x-apple.systempreferences:com.apple.preference.speech",
			Keywords:    []string{"siri", "spotlight", "search", "voice", "assistant"},
			Category:    "Siri & Spotlight",
		},
		{
			Name:        "Language & Region",
			Description: "Language, region, and date formats",
			URLScheme:   "x-apple.systempreferences:com.apple.Localization",
			Keywords:    []string{"language", "region", "locale", "date", "time", "currency"},
			Category:    "Language & Region",
		},
		{
			Name:        "Date & Time",
			Description: "Date, time, and time zone settings",
			URLScheme:   "x-apple.systempreferences:com.apple.preference.datetime",
			Keywords:    []string{"date", "time", "timezone", "clock", "calendar"},
			Category:    "Date & Time",
		},
		{
			Name:        "Sharing",
			Description: "File sharing and network services",
			URLScheme:   "x-apple.systempreferences:com.apple.preferences.sharing",
			Keywords:    []string{"sharing", "file sharing", "screen sharing", "remote", "airdrop"},
			Category:    "Sharing",
		},
		{
			Name:        "Time Machine",
			Description: "Backup settings and schedule",
			URLScheme:   "x-apple.systempreferences:com.apple.prefs.backup",
			Keywords:    []string{"time machine", "backup", "restore", "history"},
			Category:    "Backup",
		},
		{
			Name:        "Software Update",
			Description: "System and app updates",
			URLScheme:   "x-apple.systempreferences:com.apple.preferences.softwareupdate",
			Keywords:    []string{"software update", "updates", "upgrade", "patch"},
			Category:    "Software Update",
		},

		// Advanced/rarely accessed settings
		{
			Name:        "VoiceOver",
			Description: "Screen reader settings",
			URLScheme:   "x-apple.systempreferences:com.apple.preference.universalaccess?Seeing_VoiceOver",
			Keywords:    []string{"voiceover", "screen reader", "speech", "blind"},
			Category:    "Accessibility",
		},
		{
			Name:        "Zoom",
			Description: "Screen magnification settings",
			URLScheme:   "x-apple.systempreferences:com.apple.preference.universalaccess?Seeing_Zoom",
			Keywords:    []string{"zoom", "magnify", "magnification", "vision"},
			Category:    "Accessibility",
		},
		{
			Name:        "Startup Disk",
			Description: "Choose startup disk",
			URLScheme:   "x-apple.systempreferences:com.apple.preference.startupdisk",
			Keywords:    []string{"startup disk", "boot", "startup", "disk"},
			Category:    "Startup Disk",
		},
		{
			Name:        "Profiles",
			Description: "Configuration profiles",
			URLScheme:   "x-apple.systempreferences:com.apple.preferences.configurationprofiles",
			Keywords:    []string{"profiles", "configuration", "mdm", "management"},
			Category:    "Profiles",
		},
	}
}
