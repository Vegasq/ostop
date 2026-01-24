package ui

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	opensearch "github.com/opensearch-project/opensearch-go/v2"
)

// LoadFixture loads a test fixture from the testdata directory
func LoadFixture(filename string) ([]byte, error) {
	path := filepath.Join("testdata", filename)
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read fixture %s: %w", filename, err)
	}
	return data, nil
}

// NewMockClient creates an OpenSearch client with a mock transport
func NewMockClient(transport *MockTransport) (*opensearch.Client, error) {
	cfg := opensearch.Config{
		Addresses: []string{"http://localhost:9200"},
		Transport: transport,
	}
	return opensearch.NewClient(cfg)
}

// NewMockClientWithFixtures creates a mock client with all fixtures loaded
func NewMockClientWithFixtures() (*opensearch.Client, error) {
	transport := NewMockTransport()
	if err := transport.LoadAllFixtures(); err != nil {
		return nil, err
	}
	return NewMockClient(transport)
}

// NewMockClientWithError creates a mock client that returns an error for a specific endpoint
func NewMockClientWithError(endpoint string) (*opensearch.Client, error) {
	transport := NewMockTransport()
	if err := transport.LoadAllFixtures(); err != nil {
		return nil, err
	}
	transport.SetError(endpoint, fmt.Errorf("mock error for %s", endpoint))
	return NewMockClient(transport)
}

// NewTestApp creates a new App instance for testing with a mock client
func NewTestApp(client *opensearch.Client, endpoint string) *App {
	return NewApp(client, endpoint)
}

// ExecuteCommand executes a Bubble Tea command synchronously and returns the message
// This is useful for testing async operations
func ExecuteCommand(cmd tea.Cmd) tea.Msg {
	if cmd == nil {
		return nil
	}

	// Create a channel to receive the message
	msgChan := make(chan tea.Msg, 1)

	// Execute command in a goroutine
	go func() {
		msg := cmd()
		msgChan <- msg
	}()

	// Wait for the message with a timeout
	select {
	case msg := <-msgChan:
		return msg
	case <-time.After(5 * time.Second):
		return nil // Timeout
	}
}

// ExecuteCommandWithContext executes a command with a custom context
func ExecuteCommandWithContext(ctx context.Context, cmd tea.Cmd) tea.Msg {
	if cmd == nil {
		return nil
	}

	msgChan := make(chan tea.Msg, 1)

	go func() {
		msg := cmd()
		msgChan <- msg
	}()

	select {
	case msg := <-msgChan:
		return msg
	case <-ctx.Done():
		return nil
	}
}

// SetupTestApp creates and initializes a test app with fixtures loaded
func SetupTestApp() (*App, error) {
	client, err := NewMockClientWithFixtures()
	if err != nil {
		return nil, err
	}
	app := NewTestApp(client, "http://localhost:9200")
	return app, nil
}

// InitializeTestApp creates, initializes, and refreshes a test app
func InitializeTestApp() (*App, error) {
	app, err := SetupTestApp()
	if err != nil {
		return nil, err
	}

	// Execute initial refresh
	cmd := app.Init()
	msg := ExecuteCommand(cmd)
	app.Update(msg)

	return app, nil
}

// SendWindowSize sends a window size message to the app to initialize viewport
func SendWindowSize(app *App, width, height int) *App {
	msg := tea.WindowSizeMsg{Width: width, Height: height}
	app.Update(msg)
	return app
}

// SendKey sends a key message to the app
func SendKey(app *App, key string) (*App, tea.Cmd) {
	var msg tea.KeyMsg

	switch key {
	case "up", "down", "left", "right", "enter", "esc", "tab", "backspace":
		msg = tea.KeyMsg{Type: tea.KeyType(keyTypeMap[key])}
	case "pgup":
		msg = tea.KeyMsg{Type: tea.KeyPgUp}
	case "pgdown":
		msg = tea.KeyMsg{Type: tea.KeyPgDown}
	case "home":
		msg = tea.KeyMsg{Type: tea.KeyHome}
	case "end":
		msg = tea.KeyMsg{Type: tea.KeyEnd}
	default:
		// Single character keys
		msg = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(key)}
	}

	_, cmd := app.Update(msg)
	return app, cmd
}

// keyTypeMap maps string keys to tea.KeyType
var keyTypeMap = map[string]tea.KeyType{
	"up":        tea.KeyUp,
	"down":      tea.KeyDown,
	"left":      tea.KeyLeft,
	"right":     tea.KeyRight,
	"enter":     tea.KeyEnter,
	"esc":       tea.KeyEsc,
	"tab":       tea.KeyTab,
	"backspace": tea.KeyBackspace,
}

// WaitForRefresh waits for a refresh operation to complete
func WaitForRefresh(app *App, maxWait time.Duration) bool {
	start := time.Now()
	for app.loading {
		if time.Since(start) > maxWait {
			return false
		}
		time.Sleep(10 * time.Millisecond)
	}
	return true
}

// AssertNoError is a helper to check for errors in tests
func AssertNoError(t interface{ Errorf(string, ...interface{}) }, err error, message string) {
	if err != nil {
		t.Errorf("%s: %v", message, err)
	}
}

// AssertEqual is a helper to check equality in tests
func AssertEqual(t interface{ Errorf(string, ...interface{}) }, got, want interface{}, message string) {
	if got != want {
		t.Errorf("%s: got %v, want %v", message, got, want)
	}
}

// AssertNotNil is a helper to check for nil values in tests
func AssertNotNil(t interface{ Errorf(string, ...interface{}) }, value interface{}, message string) {
	if value == nil {
		t.Errorf("%s: value is nil", message)
	}
}
