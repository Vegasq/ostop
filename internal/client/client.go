package client

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"strings"

	"github.com/aws/aws-sdk-go-v2/config"
	opensearch "github.com/opensearch-project/opensearch-go/v2"
	requestsigner "github.com/opensearch-project/opensearch-go/v2/signer/awsv2"
)

// NewClient creates an OpenSearch client with automatic AWS signing detection
func NewClient(endpoint, region, profile string, insecure bool) (*opensearch.Client, error) {
	// Detect if endpoint is AWS OpenSearch
	isAWS := strings.Contains(endpoint, ".es.amazonaws.com") ||
		strings.Contains(endpoint, ".aoss.amazonaws.com")

	if isAWS {
		if region == "" {
			return nil, fmt.Errorf("--region required for AWS OpenSearch endpoints")
		}
		return newAWSClient(endpoint, region, profile)
	}

	// Local or non-AWS OpenSearch
	return newLocalClient(endpoint, insecure)
}

// newAWSClient creates a client with AWS Signature V4 signing
func newAWSClient(endpoint, region, profile string) (*opensearch.Client, error) {
	ctx := context.Background()

	// Load AWS config with optional profile
	opts := []func(*config.LoadOptions) error{
		config.WithRegion(region),
	}
	if profile != "" {
		opts = append(opts, config.WithSharedConfigProfile(profile))
	}

	cfg, err := config.LoadDefaultConfig(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	// Create AWS SigV4 signer
	signer, err := requestsigner.NewSigner(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create AWS signer: %w", err)
	}

	// Create OpenSearch client with AWS signing
	client, err := opensearch.NewClient(opensearch.Config{
		Addresses: []string{endpoint},
		Signer:    signer,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create OpenSearch client: %w", err)
	}

	return client, nil
}

// newLocalClient creates a client for local/non-AWS OpenSearch
func newLocalClient(endpoint string, insecure bool) (*opensearch.Client, error) {
	cfg := opensearch.Config{
		Addresses: []string{endpoint},
	}

	// Allow insecure TLS for local development
	if insecure {
		cfg.Transport = &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		}
	}

	client, err := opensearch.NewClient(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create OpenSearch client: %w", err)
	}

	return client, nil
}
