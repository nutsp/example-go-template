package repository

import (
	"context"
	"fmt"
	"time"

	"example-api-template/internal/domain"

	"gorm.io/gorm"
)

// PostgreSQLExampleRepository implements ExampleRepository using PostgreSQL
type PostgreSQLExampleRepository struct {
	db *gorm.DB
}

// NewPostgreSQLExampleRepository creates a new PostgreSQL repository
func NewPostgreSQLExampleRepository(db *gorm.DB) *PostgreSQLExampleRepository {
	return &PostgreSQLExampleRepository{
		db: db,
	}
}

// AutoMigrate creates or updates the database schema
func (r *PostgreSQLExampleRepository) AutoMigrate() error {
	return r.db.AutoMigrate(&domain.Example{})
}

// Create creates a new example in the database
func (r *PostgreSQLExampleRepository) Create(ctx context.Context, example *domain.Example) error {
	result := r.db.WithContext(ctx).Create(example)
	if result.Error != nil {
		// Check for unique constraint violation
		if isDuplicateKeyError(result.Error) {
			return ErrExampleAlreadyExists
		}
		return fmt.Errorf("failed to create example: %w", result.Error)
	}

	return nil
}

// GetByID retrieves an example by ID
func (r *PostgreSQLExampleRepository) GetByID(ctx context.Context, id string) (*domain.Example, error) {
	if id == "" {
		return nil, fmt.Errorf("id cannot be empty")
	}

	var example domain.Example
	result := r.db.WithContext(ctx).First(&example, "id = ?", id)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, ErrExampleNotFound
		}
		return nil, fmt.Errorf("failed to get example by ID: %w", result.Error)
	}

	return &example, nil
}

// GetByEmail retrieves an example by email
func (r *PostgreSQLExampleRepository) GetByEmail(ctx context.Context, email string) (*domain.Example, error) {
	if email == "" {
		return nil, fmt.Errorf("email cannot be empty")
	}

	var example domain.Example
	result := r.db.WithContext(ctx).First(&example, "email = ?", email)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, ErrExampleNotFound
		}
		return nil, fmt.Errorf("failed to get example by email: %w", result.Error)
	}

	return &example, nil
}

// Update updates an existing example
func (r *PostgreSQLExampleRepository) Update(ctx context.Context, example *domain.Example) error {
	// Update the UpdatedAt timestamp
	example.UpdatedAt = time.Now()

	result := r.db.WithContext(ctx).Model(&domain.Example{}).
		Where("id = ?", example.ID).
		Updates(example)

	if result.Error != nil {
		// Check for unique constraint violation
		if isDuplicateKeyError(result.Error) {
			return ErrExampleAlreadyExists
		}
		return fmt.Errorf("failed to update example: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return ErrExampleNotFound
	}

	return nil
}

// Delete deletes an example by ID
func (r *PostgreSQLExampleRepository) Delete(ctx context.Context, id string) error {
	if id == "" {
		return fmt.Errorf("id cannot be empty")
	}

	result := r.db.WithContext(ctx).Delete(&domain.Example{}, "id = ?", id)
	if result.Error != nil {
		return fmt.Errorf("failed to delete example: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return ErrExampleNotFound
	}

	return nil
}

// List retrieves a list of examples with pagination
func (r *PostgreSQLExampleRepository) List(ctx context.Context, limit, offset int) ([]*domain.Example, error) {
	var examples []domain.Example

	query := r.db.WithContext(ctx).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset)

	result := query.Find(&examples)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to list examples: %w", result.Error)
	}

	// Convert to slice of pointers
	result_examples := make([]*domain.Example, len(examples))
	for i := range examples {
		result_examples[i] = &examples[i]
	}

	return result_examples, nil
}

// Count returns the total number of examples
func (r *PostgreSQLExampleRepository) Count(ctx context.Context) (int, error) {
	var count int64
	result := r.db.WithContext(ctx).Model(&domain.Example{}).Count(&count)
	if result.Error != nil {
		return 0, fmt.Errorf("failed to count examples: %w", result.Error)
	}
	return int(count), nil
}

// ListByAge retrieves examples filtered by age range
func (r *PostgreSQLExampleRepository) ListByAge(ctx context.Context, minAge, maxAge, limit, offset int) ([]*domain.Example, error) {
	var examples []domain.Example

	query := r.db.WithContext(ctx).
		Where("age >= ? AND age <= ?", minAge, maxAge).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset)

	result := query.Find(&examples)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to list examples by age: %w", result.Error)
	}

	// Convert to slice of pointers
	result_examples := make([]*domain.Example, len(examples))
	for i := range examples {
		result_examples[i] = &examples[i]
	}

	return result_examples, nil
}

// Search searches for examples by name (case-insensitive partial match)
func (r *PostgreSQLExampleRepository) Search(ctx context.Context, query string, limit, offset int) ([]*domain.Example, error) {
	var examples []domain.Example

	searchQuery := r.db.WithContext(ctx).
		Where("LOWER(name) LIKE LOWER(?)", "%"+query+"%").
		Order("created_at DESC").
		Limit(limit).
		Offset(offset)

	result := searchQuery.Find(&examples)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to search examples: %w", result.Error)
	}

	// Convert to slice of pointers
	result_examples := make([]*domain.Example, len(examples))
	for i := range examples {
		result_examples[i] = &examples[i]
	}

	return result_examples, nil
}

// GetStats returns statistics about examples
func (r *PostgreSQLExampleRepository) GetStats(ctx context.Context) (*RepositoryStats, error) {
	var stats RepositoryStats

	// Get total count
	var totalCount int64
	if err := r.db.WithContext(ctx).Model(&domain.Example{}).Count(&totalCount).Error; err != nil {
		return nil, fmt.Errorf("failed to get total count: %w", err)
	}
	stats.TotalCount = totalCount

	// Get average age
	var avgAge *float64
	if err := r.db.WithContext(ctx).Model(&domain.Example{}).Select("AVG(age)").Scan(&avgAge).Error; err != nil {
		return nil, fmt.Errorf("failed to get average age: %w", err)
	}
	if avgAge != nil {
		stats.AverageAge = *avgAge
	} else {
		stats.AverageAge = 0
	}

	// Get age distribution
	type AgeGroup struct {
		AgeRange string
		Count    int64
	}

	var ageGroups []AgeGroup
	err := r.db.WithContext(ctx).Model(&domain.Example{}).
		Select(`
			CASE 
				WHEN age < 18 THEN 'under_18'
				WHEN age >= 18 AND age < 30 THEN '18_29'
				WHEN age >= 30 AND age < 50 THEN '30_49'
				WHEN age >= 50 AND age < 65 THEN '50_64'
				ELSE '65_plus'
			END as age_range,
			COUNT(*) as count
		`).
		Group("age_range").
		Scan(&ageGroups).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get age distribution: %w", err)
	}

	stats.AgeDistribution = make(map[string]int64)
	for _, group := range ageGroups {
		stats.AgeDistribution[group.AgeRange] = group.Count
	}

	// Get recent activity (examples created in last 24 hours)
	var recentCount int64
	yesterday := time.Now().Add(-24 * time.Hour)
	if err := r.db.WithContext(ctx).Model(&domain.Example{}).
		Where("created_at > ?", yesterday).
		Count(&recentCount).Error; err != nil {
		return nil, fmt.Errorf("failed to get recent activity: %w", err)
	}
	stats.RecentActivity = recentCount

	return &stats, nil
}

// Transaction executes a function within a database transaction
func (r *PostgreSQLExampleRepository) Transaction(ctx context.Context, fn func(ExampleRepository) error) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		txRepo := &PostgreSQLExampleRepository{db: tx}
		return fn(txRepo)
	})
}

// isDuplicateKeyError checks if the error is a duplicate key constraint violation
func isDuplicateKeyError(err error) bool {
	if err == nil {
		return false
	}

	errStr := err.Error()
	// PostgreSQL unique constraint violation error codes
	return contains(errStr, "duplicate key value violates unique constraint") ||
		contains(errStr, "UNIQUE constraint failed") ||
		contains(errStr, "pq: duplicate key value")
}

// contains checks if a string contains a substring (case-insensitive)
func contains(str, substr string) bool {
	return len(str) >= len(substr) &&
		(str == substr ||
			(len(str) > len(substr) &&
				(str[:len(substr)] == substr ||
					str[len(str)-len(substr):] == substr ||
					containsInMiddle(str, substr))))
}

// containsInMiddle checks if substr is contained in the middle of str (not at start or end)
func containsInMiddle(str, substr string) bool {
	if len(substr) == 0 || len(str) <= len(substr) {
		return false
	}

	// Check if substring is at the beginning or end
	if len(str) >= len(substr) {
		if str[:len(substr)] == substr || str[len(str)-len(substr):] == substr {
			return false // Found at start or end, not in middle
		}
	}

	// Look for substring in the middle
	for i := 1; i <= len(str)-len(substr)-1; i++ {
		if str[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// RepositoryStats holds statistics about the repository
type RepositoryStats struct {
	TotalCount      int64            `json:"total_count"`
	AverageAge      float64          `json:"average_age"`
	AgeDistribution map[string]int64 `json:"age_distribution"`
	RecentActivity  int64            `json:"recent_activity"`
}
