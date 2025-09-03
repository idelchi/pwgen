// Package clipboard provides cross-platform clipboard functionality.
package clipboard

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

// Provider represents a clipboard provider.
type Provider interface {
	// Copy copies text to the clipboard.
	Copy(text string) error

	// Name returns the name of this provider.
	Name() string

	// Available returns true if this provider is available on the current system.
	Available() bool
}

// Manager manages clipboard operations with fallback providers.
type Manager struct {
	providers []Provider
}

// NewManager creates a new clipboard manager with default providers.
func NewManager() *Manager {
	manager := &Manager{}

	// Add providers in order of preference
	manager.AddProvider(&OSC52Provider{})

	switch runtime.GOOS {
	case "darwin":
		manager.AddProvider(&PbcopyProvider{})
	case "linux":
		manager.AddProvider(&XclipProvider{})
		manager.AddProvider(&WlCopyProvider{})
	case "windows":
		manager.AddProvider(&WindowsProvider{})
	}

	return manager
}

// AddProvider adds a clipboard provider to the manager.
func (m *Manager) AddProvider(provider Provider) {
	m.providers = append(m.providers, provider)
}

// Copy copies text to the clipboard using the first available provider.
func (m *Manager) Copy(text string) error {
	var lastErr error

	for _, provider := range m.providers {
		if !provider.Available() {
			continue
		}

		err := provider.Copy(text)
		if err == nil {
			return nil
		}

		lastErr = err
	}

	if lastErr != nil {
		return fmt.Errorf("all clipboard providers failed, last error: %w", lastErr)
	}

	return errors.New("no clipboard providers available")
}

// ListProviders returns information about all providers.
func (m *Manager) ListProviders() []string {
	result := make([]string, 0, len(m.providers))

	for _, provider := range m.providers {
		status := "unavailable"

		if provider.Available() {
			status = "available"
		}

		result = append(result, fmt.Sprintf("%s (%s)", provider.Name(), status))
	}

	return result
}

// OSC52Provider uses OSC52 escape sequences for clipboard access.
type OSC52Provider struct{}

// Copy copies text using OSC52 escape sequences.
func (p *OSC52Provider) Copy(text string) error {
	// OSC52 sequence: \e]52;c;<base64_text>\e\\
	// This works in many terminal emulators

	// Encode text to base64
	encoded := encodeBase64(text)

	// Create OSC52 sequence
	sequence := fmt.Sprintf("\x1b]52;c;%s\x1b\\", encoded)

	// Write to stderr (many terminals read OSC sequences from stderr)
	_, err := fmt.Fprint(os.Stderr, sequence)

	return err
}

// Name returns the provider name.
func (p *OSC52Provider) Name() string {
	return "OSC52"
}

// Available returns true if running in a terminal.
func (p *OSC52Provider) Available() bool {
	// OSC52 works if we have a terminal
	return os.Getenv("TERM") != ""
}

// PbcopyProvider uses macOS pbcopy command.
type PbcopyProvider struct{}

// Copy copies text using pbcopy.
func (p *PbcopyProvider) Copy(text string) error {
	cmd := exec.CommandContext(context.Background(), "pbcopy")

	cmd.Stdin = strings.NewReader(text)

	return cmd.Run()
}

// Name returns the provider name.
func (p *PbcopyProvider) Name() string {
	return "pbcopy"
}

// Available returns true if pbcopy is available.
func (p *PbcopyProvider) Available() bool {
	_, err := exec.LookPath("pbcopy")

	return err == nil
}

// XclipProvider uses Linux xclip command.
type XclipProvider struct{}

// Copy copies text using xclip.
func (p *XclipProvider) Copy(text string) error {
	cmd := exec.CommandContext(context.Background(), "xclip", "-selection", "clipboard")

	cmd.Stdin = strings.NewReader(text)

	return cmd.Run()
}

// Name returns the provider name.
func (p *XclipProvider) Name() string {
	return "xclip"
}

// Available returns true if xclip is available and DISPLAY is set.
func (p *XclipProvider) Available() bool {
	if os.Getenv("DISPLAY") == "" {
		return false
	}

	_, err := exec.LookPath("xclip")

	return err == nil
}

// WlCopyProvider uses Wayland wl-copy command.
type WlCopyProvider struct{}

// Copy copies text using wl-copy.
func (p *WlCopyProvider) Copy(text string) error {
	cmd := exec.CommandContext(context.Background(), "wl-copy")

	cmd.Stdin = strings.NewReader(text)

	return cmd.Run()
}

// Name returns the provider name.
func (p *WlCopyProvider) Name() string {
	return "wl-copy"
}

// Available returns true if wl-copy is available and running under Wayland.
func (p *WlCopyProvider) Available() bool {
	if os.Getenv("WAYLAND_DISPLAY") == "" {
		return false
	}

	_, err := exec.LookPath("wl-copy")

	return err == nil
}

// WindowsProvider uses Windows clipboard API.
type WindowsProvider struct{}

// Copy copies text using Windows clipboard.
func (p *WindowsProvider) Copy(text string) error {
	// Use PowerShell to set clipboard
	cmd := exec.CommandContext(context.Background(), "powershell", "-command", "Set-Clipboard", "-Value", text)

	return cmd.Run()
}

// Name returns the provider name.
func (p *WindowsProvider) Name() string {
	return "Windows Clipboard"
}

// Available returns true if running on Windows.
func (p *WindowsProvider) Available() bool {
	return runtime.GOOS == "windows"
}

// encodeBase64 encodes text to base64 for OSC52.
func encodeBase64(text string) string {
	// Simple base64 encoding without imports
	const chars = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/"

	data := []byte(text)

	var result strings.Builder

	// base64 processes 3 bytes at a time, i is standard loop var
	for i := 0; i < len(data); i += 3 { //nolint:varnamelen // i is standard loop variable for base64 processing
		b := uint32(data[i]) << 16 //nolint:mnd,varnamelen // base64 bit shift positions, b is temp var
		if i+1 < len(data) {
			b |= uint32(data[i+1]) << 8 //nolint:mnd // base64 bit shift positions
		}

		if i+2 < len(data) {
			b |= uint32(data[i+2])
		}

		result.WriteByte(chars[(b>>18)&63]) //nolint:mnd // base64 bit shift positions
		result.WriteByte(chars[(b>>12)&63]) //nolint:mnd // base64 bit shift positions

		if i+1 < len(data) {
			result.WriteByte(chars[(b>>6)&63]) //nolint:mnd // base64 bit shift positions
		} else {
			result.WriteByte('=')
		}

		if i+2 < len(data) {
			result.WriteByte(chars[b&63])
		} else {
			result.WriteByte('=')
		}
	}

	return result.String()
}

// Copy is a convenience function that copies text using the default manager.
func Copy(text string) error {
	manager := NewManager()

	return manager.Copy(text)
}
