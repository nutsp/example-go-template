package domain

import (
	"errors"
	"fmt"
	"regexp"
	"time"
)

// Example represents the core business entity
type Example struct {
	ID        string    `json:"id" gorm:"primaryKey;size:255"`
	Name      string    `json:"name" gorm:"size:255;not null;index"`
	Email     string    `json:"email" gorm:"size:255;not null;unique;index"`
	Age       int       `json:"age" gorm:"not null"`
	CreatedAt time.Time `json:"created_at" gorm:"not null"`
	UpdatedAt time.Time `json:"updated_at" gorm:"not null"`
}

// NewExample creates a new Example entity with validation
func NewExample(id, name, email string, age int) (*Example, error) {
	if err := validateExample(name, email, age); err != nil {
		return nil, err
	}

	now := time.Now()
	return &Example{
		ID:        id,
		Name:      name,
		Email:     email,
		Age:       age,
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}

// TableName returns the table name for GORM
func (Example) TableName() string {
	return "examples"
}

// Update updates the example entity with validation
func (e *Example) Update(name, email string, age int) error {
	if err := validateExample(name, email, age); err != nil {
		return err
	}

	e.Name = name
	e.Email = email
	e.Age = age
	e.UpdatedAt = time.Now()
	return nil
}

// validateExample validates the example fields
func validateExample(name, email string, age int) error {
	if name == "" {
		return errors.New("name cannot be empty")
	}
	if len(name) > 100 {
		return errors.New("name cannot exceed 100 characters")
	}

	if email == "" {
		return errors.New("email cannot be empty")
	}
	if !isValidEmail(email) {
		return errors.New("invalid email format")
	}

	if age < 0 {
		return errors.New("age cannot be negative")
	}
	if age > 150 {
		return errors.New("age cannot exceed 150")
	}

	return nil
}

// isValidEmail validates email format using regex
func isValidEmail(email string) bool {
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	return emailRegex.MatchString(email)
}

// String returns a string representation of the Example
func (e *Example) String() string {
	return fmt.Sprintf("Example{ID: %s, Name: %s, Email: %s, Age: %d}", e.ID, e.Name, e.Email, e.Age)
}
