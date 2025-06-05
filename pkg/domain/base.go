package domain

import (
	"time"

	"gorm.io/gorm"
)

// BaseModel defines the contract for all domain entities
// Optimized for both relational and document-based querying
type BaseModel interface {
	GetID() int
	GetSlug() string
	SetSlug(slug string)
	GetCreatedAt() time.Time
	GetUpdatedAt() time.Time
	GetArchivedAt() gorm.DeletedAt
	GetName() string
}

// SortDirection represents sorting order
type SortDirection string

const (
	SortAsc  SortDirection = "asc"
	SortDesc SortDirection = "desc"
)

// SortMap defines field-level sorting configuration
// Using map for O(1) lookup complexity during query building
type SortMap map[string]SortDirection

// QueryParams provides type-safe query configuration with generics
// Designed for efficient query construction and caching
type QueryParams[E BaseModel] struct {
	Filter  E        `json:"filter,omitempty"`
	Sort    SortMap  `json:"sort,omitempty"`
	Include []string `json:"include,omitempty"` // Eager loading relationships
	Limit   int      `json:"limit,omitempty"`   // Pagination size (max 1000 for performance)
	Offset  int      `json:"offset,omitempty"`  // Pagination offset
}

// Validate ensures query parameters are within acceptable bounds
// Prevents potential DoS through excessive limit values
func (q *QueryParams[E]) Validate() error {
	if q.Limit < 0 {
		q.Limit = 10 // Default page size
	}
	if q.Limit > 1000 {
		q.Limit = 1000 // Maximum page size for performance
	}
	if q.Offset < 0 {
		q.Offset = 0
	}
	return nil
}

// GetPageInfo calculates pagination metadata
func (q *QueryParams[E]) GetPageInfo() (page int, size int) {
	size = q.Limit
	if size == 0 {
		size = 10
	}
	page = (q.Offset / size) + 1
	return page, size
}
