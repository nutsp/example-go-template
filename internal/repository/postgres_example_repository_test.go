package repository

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"example-api-template/internal/config"
	"example-api-template/internal/domain"
	"example-api-template/pkg/database"
	"example-api-template/pkg/logger"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// PostgreSQLRepositoryTestSuite defines the test suite for PostgreSQL repository
type PostgreSQLRepositoryTestSuite struct {
	suite.Suite
	db         *gorm.DB
	repository *PostgreSQLExampleRepository
	ctx        context.Context
}

// SetupSuite runs once before all tests in the suite
func (suite *PostgreSQLRepositoryTestSuite) SetupSuite() {
	suite.ctx = context.Background()

	// Use SQLite in-memory database for testing (compatible with GORM)
	// This avoids the need for a real PostgreSQL instance during testing
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		NowFunc: func() time.Time {
			return time.Now().UTC()
		},
	})
	require.NoError(suite.T(), err)

	suite.db = db
	suite.repository = NewPostgreSQLExampleRepository(db)

	// Run migrations
	err = suite.repository.AutoMigrate()
	require.NoError(suite.T(), err)
}

// TearDownSuite runs once after all tests in the suite
func (suite *PostgreSQLRepositoryTestSuite) TearDownSuite() {
	if suite.db != nil {
		sqlDB, _ := suite.db.DB()
		if sqlDB != nil {
			sqlDB.Close()
		}
	}
}

// SetupTest runs before each test
func (suite *PostgreSQLRepositoryTestSuite) SetupTest() {
	// Clean up the database before each test
	suite.db.Exec("DELETE FROM examples")
}

// TestCreate tests the Create method
func (suite *PostgreSQLRepositoryTestSuite) TestCreate() {
	example := suite.createValidExample()

	err := suite.repository.Create(suite.ctx, example)
	assert.NoError(suite.T(), err)

	// Verify the example was created
	retrieved, err := suite.repository.GetByID(suite.ctx, example.ID)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), example.ID, retrieved.ID)
	assert.Equal(suite.T(), example.Name, retrieved.Name)
	assert.Equal(suite.T(), example.Email, retrieved.Email)
	assert.Equal(suite.T(), example.Age, retrieved.Age)
}

// TestCreateDuplicateEmail tests creating an example with duplicate email
func (suite *PostgreSQLRepositoryTestSuite) TestCreateDuplicateEmail() {
	example1 := suite.createValidExample()
	err := suite.repository.Create(suite.ctx, example1)
	assert.NoError(suite.T(), err)

	// Try to create another example with the same email
	example2 := suite.createValidExample()
	example2.ID = uuid.New().String()
	example2.Email = example1.Email // Same email

	err = suite.repository.Create(suite.ctx, example2)
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), ErrExampleAlreadyExists, err)
}

// TestGetByID tests the GetByID method
func (suite *PostgreSQLRepositoryTestSuite) TestGetByID() {
	// Test getting non-existent example
	_, err := suite.repository.GetByID(suite.ctx, "non-existent-id")
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), ErrExampleNotFound, err)

	// Test getting existing example
	example := suite.createValidExample()
	err = suite.repository.Create(suite.ctx, example)
	assert.NoError(suite.T(), err)

	retrieved, err := suite.repository.GetByID(suite.ctx, example.ID)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), example.ID, retrieved.ID)
	assert.Equal(suite.T(), example.Name, retrieved.Name)
	assert.Equal(suite.T(), example.Email, retrieved.Email)
	assert.Equal(suite.T(), example.Age, retrieved.Age)

	// Test empty ID
	_, err = suite.repository.GetByID(suite.ctx, "")
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "id cannot be empty")
}

// TestGetByEmail tests the GetByEmail method
func (suite *PostgreSQLRepositoryTestSuite) TestGetByEmail() {
	// Test getting non-existent example
	_, err := suite.repository.GetByEmail(suite.ctx, "nonexistent@example.com")
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), ErrExampleNotFound, err)

	// Test getting existing example
	example := suite.createValidExample()
	err = suite.repository.Create(suite.ctx, example)
	assert.NoError(suite.T(), err)

	retrieved, err := suite.repository.GetByEmail(suite.ctx, example.Email)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), example.ID, retrieved.ID)
	assert.Equal(suite.T(), example.Name, retrieved.Name)
	assert.Equal(suite.T(), example.Email, retrieved.Email)
	assert.Equal(suite.T(), example.Age, retrieved.Age)

	// Test empty email
	_, err = suite.repository.GetByEmail(suite.ctx, "")
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "email cannot be empty")
}

// TestUpdate tests the Update method
func (suite *PostgreSQLRepositoryTestSuite) TestUpdate() {
	// Create an example
	example := suite.createValidExample()
	err := suite.repository.Create(suite.ctx, example)
	assert.NoError(suite.T(), err)

	// Update the example
	updatedExample, _ := domain.NewExample(example.ID, "Updated Name", "updated@example.com", 35)
	err = suite.repository.Update(suite.ctx, updatedExample)
	assert.NoError(suite.T(), err)

	// Verify the update
	retrieved, err := suite.repository.GetByID(suite.ctx, example.ID)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "Updated Name", retrieved.Name)
	assert.Equal(suite.T(), "updated@example.com", retrieved.Email)
	assert.Equal(suite.T(), 35, retrieved.Age)

	// Test updating non-existent example
	nonExistentExample, _ := domain.NewExample("non-existent-id", "Test", "test@example.com", 25)
	err = suite.repository.Update(suite.ctx, nonExistentExample)
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), ErrExampleNotFound, err)
}

// TestUpdateDuplicateEmail tests updating with duplicate email
func (suite *PostgreSQLRepositoryTestSuite) TestUpdateDuplicateEmail() {
	// Create two examples
	example1 := suite.createValidExample()
	err := suite.repository.Create(suite.ctx, example1)
	assert.NoError(suite.T(), err)

	example2 := suite.createValidExample()
	example2.ID = uuid.New().String()
	example2.Email = "different@example.com"
	err = suite.repository.Create(suite.ctx, example2)
	assert.NoError(suite.T(), err)

	// Try to update example2 with example1's email
	updatedExample, _ := domain.NewExample(example2.ID, "Updated", example1.Email, 30)
	err = suite.repository.Update(suite.ctx, updatedExample)
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), ErrExampleAlreadyExists, err)
}

// TestDelete tests the Delete method
func (suite *PostgreSQLRepositoryTestSuite) TestDelete() {
	// Create an example
	example := suite.createValidExample()
	err := suite.repository.Create(suite.ctx, example)
	assert.NoError(suite.T(), err)

	// Delete the example
	err = suite.repository.Delete(suite.ctx, example.ID)
	assert.NoError(suite.T(), err)

	// Verify deletion
	_, err = suite.repository.GetByID(suite.ctx, example.ID)
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), ErrExampleNotFound, err)

	// Test deleting non-existent example
	err = suite.repository.Delete(suite.ctx, "non-existent-id")
	assert.Error(suite.T(), err)
	assert.Equal(suite.T(), ErrExampleNotFound, err)

	// Test empty ID
	err = suite.repository.Delete(suite.ctx, "")
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "id cannot be empty")
}

// TestList tests the List method
func (suite *PostgreSQLRepositoryTestSuite) TestList() {
	// Test empty list
	examples, err := suite.repository.List(suite.ctx, 10, 0)
	assert.NoError(suite.T(), err)
	assert.Empty(suite.T(), examples)

	// Create multiple examples
	exampleCount := 5
	createdExamples := make([]*domain.Example, exampleCount)
	for i := 0; i < exampleCount; i++ {
		example := suite.createValidExample()
		example.ID = uuid.New().String()
		example.Email = fmt.Sprintf("test%d@example.com", i)
		example.Name = fmt.Sprintf("Test User %d", i)
		createdExamples[i] = example

		err := suite.repository.Create(suite.ctx, example)
		assert.NoError(suite.T(), err)
	}

	// Test listing all examples
	examples, err = suite.repository.List(suite.ctx, 10, 0)
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), examples, exampleCount)

	// Test pagination
	examples, err = suite.repository.List(suite.ctx, 2, 0)
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), examples, 2)

	examples, err = suite.repository.List(suite.ctx, 2, 2)
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), examples, 2)

	examples, err = suite.repository.List(suite.ctx, 2, 4)
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), examples, 1)
}

// TestCount tests the Count method
func (suite *PostgreSQLRepositoryTestSuite) TestCount() {
	// Test empty count
	count, err := suite.repository.Count(suite.ctx)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), 0, count)

	// Create examples
	exampleCount := 3
	for i := 0; i < exampleCount; i++ {
		example := suite.createValidExample()
		example.ID = uuid.New().String()
		example.Email = fmt.Sprintf("test%d@example.com", i)
		err := suite.repository.Create(suite.ctx, example)
		assert.NoError(suite.T(), err)
	}

	// Test count
	count, err = suite.repository.Count(suite.ctx)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), exampleCount, count)
}

// TestListByAge tests the ListByAge method
func (suite *PostgreSQLRepositoryTestSuite) TestListByAge() {
	// Create examples with different ages
	ages := []int{20, 25, 30, 35, 40}
	for i, age := range ages {
		example := suite.createValidExample()
		example.ID = uuid.New().String()
		example.Email = fmt.Sprintf("test%d@example.com", i)
		example.Age = age
		err := suite.repository.Create(suite.ctx, example)
		assert.NoError(suite.T(), err)
	}

	// Test age range filter
	examples, err := suite.repository.ListByAge(suite.ctx, 25, 35, 10, 0)
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), examples, 3) // Ages 25, 30, 35

	for _, example := range examples {
		assert.GreaterOrEqual(suite.T(), example.Age, 25)
		assert.LessOrEqual(suite.T(), example.Age, 35)
	}
}

// TestSearch tests the Search method
func (suite *PostgreSQLRepositoryTestSuite) TestSearch() {
	// Create examples with different names
	names := []string{"John Doe", "Jane Smith", "John Johnson", "Alice Cooper"}
	for i, name := range names {
		example := suite.createValidExample()
		example.ID = uuid.New().String()
		example.Email = fmt.Sprintf("test%d@example.com", i)
		example.Name = name
		err := suite.repository.Create(suite.ctx, example)
		assert.NoError(suite.T(), err)
	}

	// Test search by partial name
	examples, err := suite.repository.Search(suite.ctx, "john", 10, 0)
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), examples, 2) // John Doe, John Johnson

	// Test case-insensitive search
	examples, err = suite.repository.Search(suite.ctx, "JOHN", 10, 0)
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), examples, 2)

	// Test search with no results
	examples, err = suite.repository.Search(suite.ctx, "nonexistent", 10, 0)
	assert.NoError(suite.T(), err)
	assert.Empty(suite.T(), examples)
}

// TestGetStats tests the GetStats method
func (suite *PostgreSQLRepositoryTestSuite) TestGetStats() {
	// Test empty stats
	stats, err := suite.repository.GetStats(suite.ctx)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), int64(0), stats.TotalCount)
	assert.Equal(suite.T(), float64(0), stats.AverageAge)

	// Create examples with different ages
	ages := []int{17, 25, 35, 55, 70}
	for i, age := range ages {
		example := suite.createValidExample()
		example.ID = uuid.New().String()
		example.Email = fmt.Sprintf("test%d@example.com", i)
		example.Age = age
		err := suite.repository.Create(suite.ctx, example)
		assert.NoError(suite.T(), err)
	}

	// Test stats
	stats, err = suite.repository.GetStats(suite.ctx)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), int64(5), stats.TotalCount)
	assert.Equal(suite.T(), float64(40.4), stats.AverageAge) // (17+25+35+55+70)/5 = 40.4
	assert.NotNil(suite.T(), stats.AgeDistribution)
}

// TestTransaction tests the Transaction method
func (suite *PostgreSQLRepositoryTestSuite) TestTransaction() {
	// Test successful transaction
	err := suite.repository.Transaction(suite.ctx, func(txRepo ExampleRepository) error {
		example1 := suite.createValidExample()
		if err := txRepo.Create(suite.ctx, example1); err != nil {
			return err
		}

		example2 := suite.createValidExample()
		example2.ID = uuid.New().String()
		example2.Email = "different@example.com"
		return txRepo.Create(suite.ctx, example2)
	})
	assert.NoError(suite.T(), err)

	// Verify both examples were created
	count, err := suite.repository.Count(suite.ctx)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), 2, count)

	// Test failed transaction (should rollback)
	err = suite.repository.Transaction(suite.ctx, func(txRepo ExampleRepository) error {
		example3 := suite.createValidExample()
		example3.ID = uuid.New().String()
		example3.Email = "another@example.com"
		if err := txRepo.Create(suite.ctx, example3); err != nil {
			return err
		}

		// This should cause a rollback
		return fmt.Errorf("simulated error")
	})
	assert.Error(suite.T(), err)

	// Verify the transaction was rolled back
	count, err = suite.repository.Count(suite.ctx)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), 2, count) // Still 2, not 3
}

// TestDomainStructWithGORM tests that the domain struct works correctly with GORM
func (suite *PostgreSQLRepositoryTestSuite) TestDomainStructWithGORM() {
	now := time.Now().UTC()
	example := &domain.Example{
		ID:        uuid.New().String(),
		Name:      "Test User",
		Email:     "test@example.com",
		Age:       25,
		CreatedAt: now,
		UpdatedAt: now,
	}

	// Test that the domain struct can be used directly with GORM
	assert.Equal(suite.T(), "examples", example.TableName())
	assert.NotEmpty(suite.T(), example.ID)
	assert.Equal(suite.T(), "Test User", example.Name)
	assert.Equal(suite.T(), "test@example.com", example.Email)
	assert.Equal(suite.T(), 25, example.Age)
	assert.Equal(suite.T(), now, example.CreatedAt)
	assert.Equal(suite.T(), now, example.UpdatedAt)
}

// TestIsDuplicateKeyError tests the isDuplicateKeyError function
func (suite *PostgreSQLRepositoryTestSuite) TestIsDuplicateKeyError() {
	assert.False(suite.T(), isDuplicateKeyError(nil))
	assert.False(suite.T(), isDuplicateKeyError(fmt.Errorf("some other error")))
	assert.True(suite.T(), isDuplicateKeyError(fmt.Errorf("duplicate key value violates unique constraint")))
	assert.True(suite.T(), isDuplicateKeyError(fmt.Errorf("UNIQUE constraint failed")))
	assert.True(suite.T(), isDuplicateKeyError(fmt.Errorf("pq: duplicate key value")))
}

// Helper method to create a valid example
func (suite *PostgreSQLRepositoryTestSuite) createValidExample() *domain.Example {
	example, _ := domain.NewExample(
		uuid.New().String(),
		"Test User",
		"test@example.com",
		25,
	)
	return example
}

// TestPostgreSQLRepositoryTestSuite runs the test suite
func TestPostgreSQLRepositoryTestSuite(t *testing.T) {
	suite.Run(t, new(PostgreSQLRepositoryTestSuite))
}

// Integration tests that require a real PostgreSQL database
func TestPostgreSQLIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	// Skip if no PostgreSQL connection available
	dbURL := os.Getenv("TEST_DATABASE_URL")
	if dbURL == "" {
		t.Skip("TEST_DATABASE_URL not set, skipping PostgreSQL integration tests")
	}

	t.Run("RealPostgreSQLConnection", func(t *testing.T) {
		testRealPostgreSQLConnection(t, dbURL)
	})
}

// testRealPostgreSQLConnection tests with a real PostgreSQL database
func testRealPostgreSQLConnection(t *testing.T, dbURL string) {
	// Parse database URL and create config
	cfg := &config.DatabaseConfig{
		Type:            "postgres",
		Host:            "localhost",
		Port:            5432,
		Name:            "test_db",
		Username:        "test_user",
		Password:        "test_password",
		SSLMode:         "disable",
		MaxConnections:  10,
		MaxIdleConns:    2,
		ConnMaxLifetime: 5 * time.Minute,
	}

	logger, err := logger.New(&config.LoggerConfig{
		Level:  "error",
		Format: "console",
	})
	require.NoError(t, err)
	defer logger.Close()

	// Test connection with retry
	conn, err := database.TestConnection(cfg, logger, 3, time.Second)
	if err != nil {
		t.Skipf("Could not connect to PostgreSQL: %v", err)
	}
	defer conn.Close()

	// Create repository
	repo := NewPostgreSQLExampleRepository(conn.DB)

	// Run migrations
	err = repo.AutoMigrate()
	require.NoError(t, err)

	// Clean up
	defer conn.DB.Exec("TRUNCATE TABLE examples")

	ctx := context.Background()

	// Test basic operations
	example, _ := domain.NewExample(
		uuid.New().String(),
		"Integration Test User",
		"integration@example.com",
		30,
	)

	// Create
	err = repo.Create(ctx, example)
	assert.NoError(t, err)

	// Get
	retrieved, err := repo.GetByID(ctx, example.ID)
	assert.NoError(t, err)
	assert.Equal(t, example.ID, retrieved.ID)

	// Update
	retrieved.Name = "Updated User"
	err = repo.Update(ctx, retrieved)
	assert.NoError(t, err)

	// List
	examples, err := repo.List(ctx, 10, 0)
	assert.NoError(t, err)
	assert.Len(t, examples, 1)

	// Delete
	err = repo.Delete(ctx, example.ID)
	assert.NoError(t, err)

	// Verify deletion
	_, err = repo.GetByID(ctx, example.ID)
	assert.Error(t, err)
	assert.Equal(t, ErrExampleNotFound, err)
}

// Benchmark tests
func BenchmarkPostgreSQLRepository(b *testing.B) {
	// Setup in-memory SQLite for benchmarking
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		b.Fatal(err)
	}

	repo := NewPostgreSQLExampleRepository(db)
	repo.AutoMigrate()

	ctx := context.Background()

	b.Run("Create", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			example, _ := domain.NewExample(
				uuid.New().String(),
				fmt.Sprintf("User %d", i),
				fmt.Sprintf("user%d@example.com", i),
				25,
			)
			repo.Create(ctx, example)
		}
	})

	// Create some test data for read benchmarks
	for i := 0; i < 100; i++ {
		example, _ := domain.NewExample(
			uuid.New().String(),
			fmt.Sprintf("User %d", i),
			fmt.Sprintf("user%d@example.com", i),
			25+i%50,
		)
		repo.Create(ctx, example)
	}

	b.Run("GetByID", func(b *testing.B) {
		examples, _ := repo.List(ctx, 100, 0)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			repo.GetByID(ctx, examples[i%len(examples)].ID)
		}
	})

	b.Run("List", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			repo.List(ctx, 10, 0)
		}
	})

	b.Run("Count", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			repo.Count(ctx)
		}
	})
}
