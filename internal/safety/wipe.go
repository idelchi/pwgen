// Package safety provides secure memory handling and data wiping functionality.
package safety

import (
	"crypto/rand"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

// SecureString represents a string that can be securely wiped from memory.
type SecureString struct {
	data []byte
	mu   sync.RWMutex
}

// NewSecureString creates a new secure string.
func NewSecureString(s string) *SecureString {
	ss := &SecureString{
		data: make([]byte, len(s)),
	}
	copy(ss.data, []byte(s))

	return ss
}

// String returns the string value (read-only access).
func (ss *SecureString) String() string {
	ss.mu.RLock()
	defer ss.mu.RUnlock()

	if ss.data == nil {
		return ""
	}

	// Return a copy to prevent external modification
	return string(ss.data)
}

// Wipe securely clears the string from memory.
func (ss *SecureString) Wipe() {
	ss.mu.Lock()
	defer ss.mu.Unlock()

	if ss.data != nil {
		WipeBytes(ss.data)

		ss.data = nil
	}
}

// Len returns the length of the string.
func (ss *SecureString) Len() int {
	ss.mu.RLock()
	defer ss.mu.RUnlock()

	if ss.data == nil {
		return 0
	}

	return len(ss.data)
}

// IsWiped returns true if the string has been wiped.
func (ss *SecureString) IsWiped() bool {
	ss.mu.RLock()
	defer ss.mu.RUnlock()

	return ss.data == nil
}

// WipeBytes securely overwrites a byte slice with random data, then zeros.
func WipeBytes(data []byte) {
	if len(data) == 0 {
		return
	}

	// First pass: fill with random data
	_, _ = rand.Read(data)

	// Second pass: fill with zeros
	for i := range data {
		data[i] = 0
	}

	// Third pass: fill with 0xFF
	for i := range data {
		data[i] = 0xFF
	}

	// Final pass: zeros again
	for i := range data {
		data[i] = 0
	}
}

// WipeString securely overwrites a string by converting to bytes and wiping.
// Note: This may not be completely effective due to Go's string immutability
// and potential compiler optimizations. Use SecureString for better protection.
func WipeString(str *string) {
	if str == nil || *str == "" {
		return
	}

	// Convert to byte slice and wipe (this is a best-effort approach)
	data := []byte(*str)
	WipeBytes(data)

	// Clear the string pointer
	*str = ""
}

// CleanupManager manages cleanup of sensitive data on program exit.
type CleanupManager struct {
	items []func()
	mu    sync.Mutex
	once  sync.Once
}

// globalCleanupManager is the default cleanup manager instance.
//
//nolint:gochecknoglobals // Global cleanup manager is necessary for signal handling
var globalCleanupManager = &CleanupManager{}

// RegisterCleanup adds a cleanup function to be called on program exit.
func RegisterCleanup(cleanup func()) {
	globalCleanupManager.Register(cleanup)
}

// Register adds a cleanup function to this manager.
func (cm *CleanupManager) Register(cleanup func()) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	cm.items = append(cm.items, cleanup)
}

// Cleanup runs all registered cleanup functions.
func (cm *CleanupManager) Cleanup() {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	for _, cleanup := range cm.items {
		if cleanup != nil {
			cleanup()
		}
	}

	cm.items = nil
}

// InstallSignalHandlers installs signal handlers to ensure cleanup on exit.
func InstallSignalHandlers() {
	globalCleanupManager.once.Do(func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan,
			syscall.SIGINT,
			syscall.SIGTERM,
			syscall.SIGQUIT,
		)

		go func() {
			<-sigChan
			globalCleanupManager.Cleanup()
			os.Exit(0) //nolint:forbidigo // Required for signal handler cleanup
		}()
	})
}

// Cleanup runs the global cleanup manager.
func Cleanup() {
	globalCleanupManager.Cleanup()
}

// SecureBuffer provides a buffer that can be securely wiped.
type SecureBuffer struct {
	data []byte
	mu   sync.RWMutex
}

// NewSecureBuffer creates a new secure buffer with the specified capacity.
func NewSecureBuffer(capacity int) *SecureBuffer {
	return &SecureBuffer{
		data: make([]byte, 0, capacity),
	}
}

// Write appends data to the buffer.
func (sb *SecureBuffer) Write(data []byte) {
	sb.mu.Lock()
	defer sb.mu.Unlock()

	sb.data = append(sb.data, data...)
}

// WriteString appends a string to the buffer.
func (sb *SecureBuffer) WriteString(s string) {
	sb.Write([]byte(s))
}

// String returns the buffer contents as a string.
func (sb *SecureBuffer) String() string {
	sb.mu.RLock()
	defer sb.mu.RUnlock()

	return string(sb.data)
}

// Bytes returns a copy of the buffer contents.
func (sb *SecureBuffer) Bytes() []byte {
	sb.mu.RLock()
	defer sb.mu.RUnlock()

	result := make([]byte, len(sb.data))
	copy(result, sb.data)

	return result
}

// Len returns the length of the buffer.
func (sb *SecureBuffer) Len() int {
	sb.mu.RLock()
	defer sb.mu.RUnlock()

	return len(sb.data)
}

// Reset clears the buffer.
func (sb *SecureBuffer) Reset() {
	sb.mu.Lock()
	defer sb.mu.Unlock()

	sb.data = sb.data[:0]
}

// Wipe securely clears the buffer.
func (sb *SecureBuffer) Wipe() {
	sb.mu.Lock()
	defer sb.mu.Unlock()

	if len(sb.data) > 0 {
		WipeBytes(sb.data)

		sb.data = sb.data[:0]
	}
}

// MaskString creates a masked version of a string for display purposes.
func MaskString(s string, maskChar rune) string {
	if s == "" {
		return ""
	}

	result := make([]rune, len([]rune(s)))
	for i := range result {
		result[i] = maskChar
	}

	return string(result)
}

// MaskStringPartial creates a partially masked version showing first and last characters.
func MaskStringPartial(s string, maskChar rune, showCount int) string {
	runes := []rune(s)
	if len(runes) <= showCount*2 {
		return MaskString(s, maskChar)
	}

	result := make([]rune, len(runes))

	// Copy first characters
	for i := 0; i < showCount && i < len(runes); i++ {
		result[i] = runes[i]
	}

	// Fill middle with mask characters
	for i := showCount; i < len(runes)-showCount; i++ {
		result[i] = maskChar
	}

	// Copy last characters
	for i := len(runes) - showCount; i < len(runes); i++ {
		result[i] = runes[i]
	}

	return string(result)
}
