package ui

import (
	"bytes"
	"io"
	"net/http"
)

// MockResponse implements the OpenSearch response interface for testing
type MockResponse struct {
	body       io.ReadCloser
	statusCode int
	isError    bool
}

// NewMockResponse creates a new mock response with the given data and status code
func NewMockResponse(data []byte, statusCode int) *MockResponse {
	return &MockResponse{
		body:       io.NopCloser(bytes.NewReader(data)),
		statusCode: statusCode,
		isError:    statusCode >= 400,
	}
}

// NewMockErrorResponse creates a new mock response that represents an error
func NewMockErrorResponse(statusCode int) *MockResponse {
	return &MockResponse{
		body:       io.NopCloser(bytes.NewReader([]byte("{}"))),
		statusCode: statusCode,
		isError:    true,
	}
}

// Body returns the response body
func (m *MockResponse) Body() io.ReadCloser {
	return m.body
}

// Status returns the HTTP status code as a string
func (m *MockResponse) Status() string {
	return http.StatusText(m.statusCode)
}

// StatusCode returns the HTTP status code
func (m *MockResponse) StatusCode() int {
	return m.statusCode
}

// IsError returns whether the response represents an error
func (m *MockResponse) IsError() bool {
	return m.isError
}

// Close closes the response body
func (m *MockResponse) Close() error {
	return m.body.Close()
}
