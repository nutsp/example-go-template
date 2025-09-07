package repository

import (
	"context"
	"errors"
	"fmt"
	"time"
)

var (
	ErrExternalAPIUnavailable = errors.New("external API unavailable")
	ErrExternalAPITimeout     = errors.New("external API timeout")
	ErrInvalidExternalData    = errors.New("invalid external data")
)

// ExternalExampleData represents data from external API
type ExternalExampleData struct {
	ExternalID   string            `json:"external_id"`
	Metadata     map[string]string `json:"metadata"`
	Score        float64           `json:"score"`
	LastModified time.Time         `json:"last_modified"`
}

// ExternalExampleAPI defines the interface for external API interactions
type ExternalExampleAPI interface {
	// GetExampleData fetches additional data for an example from external source
	GetExampleData(ctx context.Context, exampleID string) (*ExternalExampleData, error)

	// ValidateExample validates an example against external rules
	ValidateExample(ctx context.Context, name, email string, age int) (bool, error)

	// EnrichExample enriches example data with external information
	EnrichExample(ctx context.Context, exampleID string) (map[string]interface{}, error)

	// NotifyExampleCreated sends notification about new example creation
	NotifyExampleCreated(ctx context.Context, exampleID, email string) error
}

// MockExternalExampleAPI is a mock implementation for testing and development
type MockExternalExampleAPI struct {
	shouldFail bool
	delay      time.Duration
}

// NewMockExternalExampleAPI creates a new mock external API
func NewMockExternalExampleAPI(shouldFail bool, delay time.Duration) *MockExternalExampleAPI {
	return &MockExternalExampleAPI{
		shouldFail: shouldFail,
		delay:      delay,
	}
}

// GetExampleData returns mock external data
func (m *MockExternalExampleAPI) GetExampleData(ctx context.Context, exampleID string) (*ExternalExampleData, error) {
	// Simulate delay
	if m.delay > 0 {
		select {
		case <-time.After(m.delay):
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}

	if m.shouldFail {
		return nil, ErrExternalAPIUnavailable
	}

	return &ExternalExampleData{
		ExternalID: fmt.Sprintf("ext_%s", exampleID),
		Metadata: map[string]string{
			"source":    "mock_api",
			"version":   "1.0",
			"processed": time.Now().Format(time.RFC3339),
		},
		Score:        0.85,
		LastModified: time.Now(),
	}, nil
}

// ValidateExample validates example data against mock rules
func (m *MockExternalExampleAPI) ValidateExample(ctx context.Context, name, email string, age int) (bool, error) {
	// Simulate delay
	if m.delay > 0 {
		select {
		case <-time.After(m.delay):
		case <-ctx.Done():
			return false, ctx.Err()
		}
	}

	if m.shouldFail {
		return false, ErrExternalAPIUnavailable
	}

	// Mock validation rules
	if name == "invalid" {
		return false, nil
	}
	if email == "blocked@example.com" {
		return false, nil
	}
	if age < 13 {
		return false, nil // COPPA compliance
	}

	return true, nil
}

// EnrichExample returns mock enrichment data
func (m *MockExternalExampleAPI) EnrichExample(ctx context.Context, exampleID string) (map[string]interface{}, error) {
	// Simulate delay
	if m.delay > 0 {
		select {
		case <-time.After(m.delay):
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}

	if m.shouldFail {
		return nil, ErrExternalAPIUnavailable
	}

	return map[string]interface{}{
		"external_id":  fmt.Sprintf("ext_%s", exampleID),
		"risk_score":   0.1,
		"verification": "pending",
		"location_data": map[string]string{
			"country": "US",
			"region":  "CA",
		},
		"preferences": map[string]bool{
			"marketing_emails": true,
			"notifications":    false,
		},
	}, nil
}

// NotifyExampleCreated sends mock notification
func (m *MockExternalExampleAPI) NotifyExampleCreated(ctx context.Context, exampleID, email string) error {
	// Simulate delay
	if m.delay > 0 {
		select {
		case <-time.After(m.delay):
		case <-ctx.Done():
			return ctx.Err()
		}
	}

	if m.shouldFail {
		return ErrExternalAPIUnavailable
	}

	// Mock notification logic - in real implementation this would call external service
	return nil
}

// SetShouldFail configures the mock to simulate failures
func (m *MockExternalExampleAPI) SetShouldFail(shouldFail bool) {
	m.shouldFail = shouldFail
}

// SetDelay configures the mock to simulate network delays
func (m *MockExternalExampleAPI) SetDelay(delay time.Duration) {
	m.delay = delay
}
