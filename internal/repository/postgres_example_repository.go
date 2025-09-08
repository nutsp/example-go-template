package repository

import (
	"context"
	"time"

	"example-api-template/internal/domain"

	"gorm.io/gorm"
)

// RepositoryStats holds statistics about the repository
type RepositoryStats struct {
	TotalCount      int64            `json:"total_count"`
	AverageAge      float64          `json:"average_age"`
	AgeDistribution map[string]int64 `json:"age_distribution"`
	RecentActivity  int64            `json:"recent_activity"`
}

// Constants for database queries
const (
	QueryByID        = "id = ?"
	QueryByEmail     = "email = ?"
	OrderByCreatedAt = "created_at DESC"
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

	return handleErrorWithContext(result.Error, "create example", example.ID)
}

// GetByID retrieves an example by ID
func (r *PostgreSQLExampleRepository) GetByID(ctx context.Context, id string) (*domain.Example, error) {
	var example domain.Example
	result := r.db.WithContext(ctx).First(&example, QueryByID, id)
	return &example, handleErrorWithContext(result.Error, "get example by ID", id)
}

// GetByEmail retrieves an example by email
func (r *PostgreSQLExampleRepository) GetByEmail(ctx context.Context, email string) (*domain.Example, error) {
	var example domain.Example
	result := r.db.WithContext(ctx).First(&example, QueryByEmail, email)
	return &example, handleErrorWithContext(result.Error, "get example by email", email)
}

// Update updates an existing example
func (r *PostgreSQLExampleRepository) Update(ctx context.Context, example *domain.Example) error {
	example.UpdatedAt = time.Now()

	result := r.db.WithContext(ctx).Model(&domain.Example{}).
		Where(QueryByID, example.ID).
		Updates(example)

	return handleErrorWithContext(result.Error, "update example", example.ID)
}

// Delete deletes an example by ID
func (r *PostgreSQLExampleRepository) Delete(ctx context.Context, id string) error {
	result := r.db.WithContext(ctx).Delete(&domain.Example{}, QueryByID, id)
	return handleErrorWithContext(result.Error, "delete example", id)
}

// List retrieves a list of examples with pagination
func (r *PostgreSQLExampleRepository) List(ctx context.Context, limit, offset int) ([]*domain.Example, error) {
	var examples []domain.Example

	query := r.db.WithContext(ctx).
		Order(OrderByCreatedAt).
		Limit(limit).
		Offset(offset)

	result := query.Find(&examples)
	if err := handleError(result.Error); err != nil {
		return nil, err
	}

	// Convert to slice of pointers
	resultExamples := make([]*domain.Example, len(examples))
	for i := range examples {
		resultExamples[i] = &examples[i]
	}

	return resultExamples, nil
}

// Count returns the total number of examples
func (r *PostgreSQLExampleRepository) Count(ctx context.Context) (int, error) {
	var count int64
	result := r.db.WithContext(ctx).Model(&domain.Example{}).Count(&count)
	if err := handleError(result.Error); err != nil {
		return 0, err
	}
	return int(count), nil
}

// ListByAge retrieves examples filtered by age range
func (r *PostgreSQLExampleRepository) ListByAge(ctx context.Context, minAge, maxAge, limit, offset int) ([]*domain.Example, error) {
	var examples []domain.Example

	query := r.db.WithContext(ctx).
		Where("age >= ? AND age <= ?", minAge, maxAge).
		Order(OrderByCreatedAt).
		Limit(limit).
		Offset(offset)

	result := query.Find(&examples)
	if err := handleError(result.Error); err != nil {
		return nil, err
	}

	// Convert to slice of pointers
	resultExamples := make([]*domain.Example, len(examples))
	for i := range examples {
		resultExamples[i] = &examples[i]
	}

	return resultExamples, nil
}

// Search searches for examples by name (case-insensitive partial match)
func (r *PostgreSQLExampleRepository) Search(ctx context.Context, query string, limit, offset int) ([]*domain.Example, error) {
	var examples []domain.Example

	searchQuery := r.db.WithContext(ctx).
		Where("LOWER(name) LIKE LOWER(?)", "%"+query+"%").
		Order(OrderByCreatedAt).
		Limit(limit).
		Offset(offset)

	result := searchQuery.Find(&examples)
	if err := handleError(result.Error); err != nil {
		return nil, err
	}

	// Convert to slice of pointers
	resultExamples := make([]*domain.Example, len(examples))
	for i := range examples {
		resultExamples[i] = &examples[i]
	}

	return resultExamples, nil
}

// GetStats returns statistics about examples
func (r *PostgreSQLExampleRepository) GetStats(ctx context.Context) (*RepositoryStats, error) {
	var stats RepositoryStats

	// Get total count
	var totalCount int64
	err := r.db.WithContext(ctx).Model(&domain.Example{}).Count(&totalCount).Error
	if err := handleError(err); err != nil {
		return nil, err
	}

	stats.TotalCount = totalCount

	// Get average age
	var avgAge *float64
	err = r.db.WithContext(ctx).Model(&domain.Example{}).Select("AVG(age)").Scan(&avgAge).Error
	if err := handleError(err); err != nil {
		return nil, err
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
	err = r.db.WithContext(ctx).Model(&domain.Example{}).
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
	if err := handleError(err); err != nil {
		return nil, err
	}

	stats.AgeDistribution = make(map[string]int64)
	for _, group := range ageGroups {
		stats.AgeDistribution[group.AgeRange] = group.Count
	}

	// Get recent activity (examples created in last 24 hours)
	var recentCount int64
	yesterday := time.Now().Add(-24 * time.Hour)
	err = r.db.WithContext(ctx).Model(&domain.Example{}).
		Where("created_at > ?", yesterday).
		Count(&recentCount).Error
	if err := handleError(err); err != nil {
		return nil, err
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
