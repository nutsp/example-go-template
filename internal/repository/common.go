package repository

import (
	"errors"
	"fmt"

	"gorm.io/gorm"
)

var (
	ErrExampleNotFound      = errors.New("example not found")
	ErrExampleAlreadyExists = errors.New("example already exists")
	ErrDatabaseConnection   = errors.New("database connection error")
	ErrQueryTimeout         = errors.New("query timeout")
	ErrInvalidQuery         = errors.New("invalid query")
	ErrTransactionFailed    = errors.New("transaction failed")
)

func handleError(err error) error {
	if err == nil {
		return nil
	}

	if isRecordNotFoundError(err) {
		return ErrExampleNotFound
	}

	if isDuplicateKeyError(err) {
		return ErrExampleAlreadyExists
	}

	if isConnectionError(err) {
		return ErrDatabaseConnection
	}

	if isTimeoutError(err) {
		return ErrQueryTimeout
	}

	return err
}

// handleErrorWithContext provides error handling with operation context
func handleErrorWithContext(err error, operation string, resourceID string) error {
	if err == nil {
		return nil
	}

	if isRecordNotFoundError(err) {
		return ErrExampleNotFound
	}

	if isDuplicateKeyError(err) {
		return ErrExampleAlreadyExists
	}

	if isConnectionError(err) {
		return ErrDatabaseConnection
	}

	if isTimeoutError(err) {
		return ErrQueryTimeout
	}

	// Add context to the error for better debugging
	return fmt.Errorf("%s failed for resource %s: %w", operation, resourceID, err)
}

func isRecordNotFoundError(err error) bool {
	return err == gorm.ErrRecordNotFound
}

func isDuplicateKeyError(err error) bool {
	errStr := err.Error()
	return contains(errStr, "duplicate key value violates unique constraint") ||
		contains(errStr, "UNIQUE constraint failed") ||
		contains(errStr, "pq: duplicate key value")
}

func isConnectionError(err error) bool {
	errStr := err.Error()
	return contains(errStr, "connection refused") ||
		contains(errStr, "connection reset") ||
		contains(errStr, "no such host") ||
		contains(errStr, "network is unreachable") ||
		contains(errStr, "connection timeout")
}

func isTimeoutError(err error) bool {
	errStr := err.Error()
	return contains(errStr, "context deadline exceeded") ||
		contains(errStr, "timeout") ||
		contains(errStr, "deadline exceeded")
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
