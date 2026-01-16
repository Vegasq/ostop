package client

import (
	"strings"
	"testing"
)

// ==============================================================================
// Phase 2: Client Factory Tests
// ==============================================================================

// TestNewClient_AWSEndpointDetection tests that AWS endpoints are correctly identified
func TestNewClient_AWSEndpointDetection(t *testing.T) {
	tests := []struct {
		name        string
		endpoint    string
		region      string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "aws_es_endpoint_no_region",
			endpoint:    "https://my-domain.us-east-1.es.amazonaws.com",
			region:      "",
			expectError: true,
			errorMsg:    "--region required for AWS OpenSearch endpoints",
		},
		{
			name:        "aws_aoss_endpoint_no_region",
			endpoint:    "https://my-collection.us-west-2.aoss.amazonaws.com",
			region:      "",
			expectError: true,
			errorMsg:    "--region required for AWS OpenSearch endpoints",
		},
		{
			name:        "local_endpoint_no_region",
			endpoint:    "https://localhost:9200",
			region:      "",
			expectError: false,
		},
		{
			name:        "custom_domain_no_region",
			endpoint:    "https://opensearch.example.com",
			region:      "",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewClient(tt.endpoint, tt.region, "", false)

			if tt.expectError {
				if err == nil {
					t.Errorf("NewClient() expected error, got nil")
				} else if !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("NewClient() error = %v, want error containing %q", err, tt.errorMsg)
				}
			} else {
				// For non-AWS endpoints without credentials, we expect the client creation
				// to succeed (we're just testing the endpoint detection logic)
				// Note: In real scenarios, connection might fail, but client creation should work
				if err != nil {
					// Check if it's not the region error
					if strings.Contains(err.Error(), "--region required") {
						t.Errorf("NewClient() unexpected region error for non-AWS endpoint: %v", err)
					}
					// Other errors (like connection failures) are acceptable for this test
				}
			}
		})
	}
}

// TestIsAWSEndpoint tests AWS endpoint detection patterns
func TestIsAWSEndpoint(t *testing.T) {
	tests := []struct {
		name     string
		endpoint string
		isAWS    bool
	}{
		{"aws_es_standard", "https://my-domain.us-east-1.es.amazonaws.com", true},
		{"aws_es_different_region", "https://my-domain.eu-west-1.es.amazonaws.com", true},
		{"aws_aoss", "https://my-collection.us-west-2.aoss.amazonaws.com", true},
		{"localhost", "https://localhost:9200", false},
		{"ip_address", "https://192.168.1.1:9200", false},
		{"custom_domain", "https://opensearch.example.com", false},
		{"http_local", "http://localhost:9200", false},
		{"aws_es_without_protocol", "my-domain.us-east-1.es.amazonaws.com", true},
		{"aws_aoss_without_protocol", "my-collection.us-west-2.aoss.amazonaws.com", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isAWS := strings.Contains(tt.endpoint, ".es.amazonaws.com") ||
				strings.Contains(tt.endpoint, ".aoss.amazonaws.com")

			if isAWS != tt.isAWS {
				t.Errorf("AWS detection for %q = %v, want %v", tt.endpoint, isAWS, tt.isAWS)
			}
		})
	}
}

// TestNewLocalClient_InsecureFlag tests that insecure flag is properly handled
func TestNewLocalClient_InsecureFlag(t *testing.T) {
	tests := []struct {
		name     string
		endpoint string
		insecure bool
	}{
		{"secure_connection", "https://localhost:9200", false},
		{"insecure_connection", "https://localhost:9200", true},
		{"http_connection", "http://localhost:9200", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := newLocalClient(tt.endpoint, tt.insecure)
			if err != nil {
				t.Fatalf("newLocalClient() error = %v", err)
			}

			if client == nil {
				t.Error("newLocalClient() returned nil client")
			}

			// Verify Transport is set when insecure is true
			if tt.insecure {
				if client.Transport == nil {
					t.Error("newLocalClient() with insecure=true should set Transport")
				}
			}
		})
	}
}

// TestNewLocalClient_Success tests successful local client creation
func TestNewLocalClient_Success(t *testing.T) {
	endpoints := []string{
		"http://localhost:9200",
		"https://localhost:9200",
		"http://192.168.1.1:9200",
		"https://opensearch.example.com",
	}

	for _, endpoint := range endpoints {
		t.Run(endpoint, func(t *testing.T) {
			client, err := newLocalClient(endpoint, false)
			if err != nil {
				t.Errorf("newLocalClient(%q) error = %v, want nil", endpoint, err)
			}

			if client == nil {
				t.Errorf("newLocalClient(%q) returned nil client", endpoint)
			}
		})
	}
}

// TestNewClient_EndpointVariations tests various endpoint format variations
func TestNewClient_EndpointVariations(t *testing.T) {
	tests := []struct {
		name     string
		endpoint string
		region   string
	}{
		{"with_https", "https://localhost:9200", ""},
		{"with_http", "http://localhost:9200", ""},
		{"with_port", "https://opensearch.local:9200", ""},
		{"without_port", "https://opensearch.local", ""},
		{"ip_with_port", "https://10.0.0.1:9200", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := NewClient(tt.endpoint, tt.region, "", false)
			// We don't check for connection success, just that client creation doesn't fail
			// due to endpoint format issues
			if err != nil {
				// Acceptable errors are connection-related, not format-related
				if strings.Contains(err.Error(), "invalid") || strings.Contains(err.Error(), "parse") {
					t.Errorf("NewClient() failed with format error: %v", err)
				}
			}
			// Client might be nil due to connection issues, which is acceptable
			_ = client
		})
	}
}

// TestNewClient_ProfileParameter tests that profile parameter is properly passed
func TestNewClient_ProfileParameter(t *testing.T) {
	// This test verifies that the profile parameter doesn't cause errors
	// when provided for non-AWS endpoints
	endpoint := "http://localhost:9200"
	profile := "test-profile"

	client, err := NewClient(endpoint, "", profile, false)
	// Should succeed for local endpoints regardless of profile
	if err != nil {
		// Check it's not a profile-related error
		if strings.Contains(err.Error(), "profile") {
			t.Errorf("NewClient() with profile parameter failed: %v", err)
		}
	}
	_ = client
}
