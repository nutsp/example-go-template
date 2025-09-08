package repository

import (
	"context"
	"fmt"
	"sync"

	"example-api-template/internal/domain"
)

// Error message templates
const (
	ErrTemplateID    = "%w: id %s"
	ErrTemplateEmail = "%w: email %s"
)

// ExampleRepository defines the interface for example data access
type ExampleRepository interface {
	Create(ctx context.Context, example *domain.Example) error
	GetByID(ctx context.Context, id string) (*domain.Example, error)
	GetByEmail(ctx context.Context, email string) (*domain.Example, error)
	Update(ctx context.Context, example *domain.Example) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, limit, offset int) ([]*domain.Example, error)
	Count(ctx context.Context) (int, error)
}

// InMemoryExampleRepository is an in-memory implementation of ExampleRepository
type InMemoryExampleRepository struct {
	data  map[string]*domain.Example
	mutex sync.RWMutex
}

// NewInMemoryExampleRepository creates a new in-memory example repository
func NewInMemoryExampleRepository() *InMemoryExampleRepository {
	return &InMemoryExampleRepository{
		data: make(map[string]*domain.Example),
	}
}

// Create stores a new example in memory
func (r *InMemoryExampleRepository) Create(ctx context.Context, example *domain.Example) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	// Check if example with same ID already exists
	if _, exists := r.data[example.ID]; exists {
		return fmt.Errorf("%w: id %s", ErrExampleAlreadyExists, example.ID)
	}

	// Check if example with same email already exists
	for _, existing := range r.data {
		if existing.Email == example.Email {
			return fmt.Errorf(ErrTemplateEmail, ErrExampleAlreadyExists, example.Email)
		}
	}

	// Create a copy to avoid external modifications
	exampleCopy := *example
	r.data[example.ID] = &exampleCopy
	return nil
}

// GetByID retrieves an example by ID
func (r *InMemoryExampleRepository) GetByID(ctx context.Context, id string) (*domain.Example, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	example, exists := r.data[id]
	if !exists {
		return nil, fmt.Errorf("%w: id %s", ErrExampleNotFound, id)
	}

	// Return a copy to avoid external modifications
	exampleCopy := *example
	return &exampleCopy, nil
}

// GetByEmail retrieves an example by email
func (r *InMemoryExampleRepository) GetByEmail(ctx context.Context, email string) (*domain.Example, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	for _, example := range r.data {
		if example.Email == email {
			// Return a copy to avoid external modifications
			exampleCopy := *example
			return &exampleCopy, nil
		}
	}

	return nil, fmt.Errorf(ErrTemplateEmail, ErrExampleNotFound, email)
}

// Update updates an existing example
func (r *InMemoryExampleRepository) Update(ctx context.Context, example *domain.Example) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	// Check if example exists
	existing, exists := r.data[example.ID]
	if !exists {
		return fmt.Errorf("%w: id %s", ErrExampleNotFound, example.ID)
	}

	// Check if email is being changed and conflicts with another example
	if existing.Email != example.Email {
		for id, other := range r.data {
			if id != example.ID && other.Email == example.Email {
				return fmt.Errorf(ErrTemplateEmail, ErrExampleAlreadyExists, example.Email)
			}
		}
	}

	// Create a copy to avoid external modifications
	exampleCopy := *example
	r.data[example.ID] = &exampleCopy
	return nil
}

// Delete removes an example by ID
func (r *InMemoryExampleRepository) Delete(ctx context.Context, id string) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	if _, exists := r.data[id]; !exists {
		return fmt.Errorf(ErrTemplateID, ErrExampleNotFound, id)
	}

	delete(r.data, id)
	return nil
}

// List retrieves a paginated list of examples
func (r *InMemoryExampleRepository) List(ctx context.Context, limit, offset int) ([]*domain.Example, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	// Convert map to slice for pagination
	examples := make([]*domain.Example, 0, len(r.data))
	for _, example := range r.data {
		exampleCopy := *example
		examples = append(examples, &exampleCopy)
	}

	// Apply pagination
	start := offset
	if start > len(examples) {
		start = len(examples)
	}

	end := start + limit
	if end > len(examples) {
		end = len(examples)
	}

	if start >= end {
		return []*domain.Example{}, nil
	}

	return examples[start:end], nil
}

// Count returns the total number of examples
func (r *InMemoryExampleRepository) Count(ctx context.Context) (int, error) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	return len(r.data), nil
}
