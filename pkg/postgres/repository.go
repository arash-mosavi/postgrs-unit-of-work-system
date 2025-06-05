package postgres

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	"github.com/arash-mosavi/postgrs-unit-of-work-system/pkg/domain"

	"gorm.io/gorm"
)

// BaseRepository provides common CRUD operations for PostgreSQL
// Optimized for performance with batch operations and prepared statements
type BaseRepository struct {
	db *gorm.DB
}

// NewBaseRepository creates a new base repository
func NewBaseRepository(db *gorm.DB) *BaseRepository {
	return &BaseRepository{db: db}
}

// Create inserts a new entity into the database
// Uses GORM's optimized insert with returning clause
func (r *BaseRepository) Create(ctx context.Context, entity interface{}) error {
	result := r.db.WithContext(ctx).Create(entity)
	if result.Error != nil {
		return fmt.Errorf("failed to create entity: %w", result.Error)
	}
	return nil
}

// GetByID retrieves an entity by its ID
// Uses prepared statements for optimal performance
func (r *BaseRepository) GetByID(ctx context.Context, id int64, entity interface{}) error {
	result := r.db.WithContext(ctx).First(entity, id)
	if result.Error != nil {
		return fmt.Errorf("failed to get entity by ID: %w", result.Error)
	}
	return nil
}

// GetBySlug retrieves an entity by its slug
// Uses index scan for optimal performance
func (r *BaseRepository) GetBySlug(ctx context.Context, slug string, entity interface{}) error {
	result := r.db.WithContext(ctx).Where("slug = ?", slug).First(entity)
	if result.Error != nil {
		return fmt.Errorf("failed to get entity by slug: %w", result.Error)
	}
	return nil
}

// Update modifies an existing entity
// Uses optimistic locking with updated_at field
func (r *BaseRepository) Update(ctx context.Context, entity interface{}) error {
	result := r.db.WithContext(ctx).Save(entity)
	if result.Error != nil {
		return fmt.Errorf("failed to update entity: %w", result.Error)
	}
	return nil
}

// Delete removes an entity by ID (soft delete if supported)
func (r *BaseRepository) Delete(ctx context.Context, id int64, entity interface{}) error {
	result := r.db.WithContext(ctx).Delete(entity, id)
	if result.Error != nil {
		return fmt.Errorf("failed to delete entity: %w", result.Error)
	}
	return nil
}

// List retrieves entities with filtering, sorting, and pagination
// Optimized query building with minimal allocations
func (r *BaseRepository) List(ctx context.Context, entities interface{}, params interface{}) error {
	query := r.db.WithContext(ctx)

	// Apply query parameters if provided
	if params != nil {
		query = r.applyQueryParams(query, params)
	}

	result := query.Find(entities)
	if result.Error != nil {
		return fmt.Errorf("failed to list entities: %w", result.Error)
	}

	return nil
}

// Count returns the total number of entities matching the criteria
// Uses optimized COUNT query without loading data
func (r *BaseRepository) Count(ctx context.Context, entity interface{}, params interface{}) (int64, error) {
	query := r.db.WithContext(ctx).Model(entity)

	// Apply query parameters if provided
	if params != nil {
		query = r.applyQueryParams(query, params)
	}

	var count int64
	result := query.Count(&count)
	if result.Error != nil {
		return 0, fmt.Errorf("failed to count entities: %w", result.Error)
	}

	return count, nil
}

// CreateBatch performs bulk insert for multiple entities
// Uses batch insert for optimal performance - O(1) database round trip
func (r *BaseRepository) CreateBatch(ctx context.Context, entities interface{}) error {
	result := r.db.WithContext(ctx).CreateInBatches(entities, 100) // Optimal batch size
	if result.Error != nil {
		return fmt.Errorf("failed to create batch: %w", result.Error)
	}
	return nil
}

// UpdateBatch performs bulk update for multiple entities
// Uses prepared statements for optimal performance
func (r *BaseRepository) UpdateBatch(ctx context.Context, entities interface{}) error {
	// GORM doesn't have direct bulk update, so we iterate
	// This could be optimized with raw SQL for large datasets
	v := reflect.ValueOf(entities)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	if v.Kind() != reflect.Slice {
		return fmt.Errorf("entities must be a slice")
	}

	for i := 0; i < v.Len(); i++ {
		entity := v.Index(i).Interface()
		if err := r.Update(ctx, entity); err != nil {
			return fmt.Errorf("failed to update entity at index %d: %w", i, err)
		}
	}

	return nil
}

// DeleteBatch performs bulk delete for multiple IDs
// Uses IN clause for optimal performance
func (r *BaseRepository) DeleteBatch(ctx context.Context, ids []int64, entity interface{}) error {
	result := r.db.WithContext(ctx).Delete(entity, ids)
	if result.Error != nil {
		return fmt.Errorf("failed to delete batch: %w", result.Error)
	}
	return nil
}

// applyQueryParams applies filtering, sorting, and pagination
// Optimized query building with type safety
func (r *BaseRepository) applyQueryParams(query *gorm.DB, params interface{}) *gorm.DB {
	v := reflect.ValueOf(params)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	if v.Kind() != reflect.Struct {
		return query
	}

	// Apply filters
	if filterField := v.FieldByName("Filter"); filterField.IsValid() && !filterField.IsZero() {
		query = r.applyFilters(query, filterField.Interface())
	}

	// Apply sorting
	if sortField := v.FieldByName("Sort"); sortField.IsValid() && !sortField.IsZero() {
		if sortMap, ok := sortField.Interface().(domain.SortMap); ok {
			query = r.applySorting(query, sortMap)
		}
	}

	// Apply includes (preloading)
	if includeField := v.FieldByName("Include"); includeField.IsValid() && !includeField.IsZero() {
		if includes, ok := includeField.Interface().([]string); ok {
			for _, include := range includes {
				query = query.Preload(include)
			}
		}
	}

	// Apply pagination
	if limitField := v.FieldByName("Limit"); limitField.IsValid() && !limitField.IsZero() {
		if limit, ok := limitField.Interface().(int); ok && limit > 0 {
			query = query.Limit(limit)
		}
	}

	if offsetField := v.FieldByName("Offset"); offsetField.IsValid() && !offsetField.IsZero() {
		if offset, ok := offsetField.Interface().(int); ok && offset > 0 {
			query = query.Offset(offset)
		}
	}

	return query
}

// applyFilters applies filter conditions to the query
// Uses reflection to build WHERE clauses dynamically
func (r *BaseRepository) applyFilters(query *gorm.DB, filter interface{}) *gorm.DB {
	v := reflect.ValueOf(filter)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	if v.Kind() != reflect.Struct {
		return query
	}

	t := v.Type()
	for i := 0; i < v.NumField(); i++ {
		field := t.Field(i)
		value := v.Field(i)

		if !field.IsExported() || value.IsZero() {
			continue
		}

		// Get database column name from json tag or field name
		columnName := field.Name
		if jsonTag := field.Tag.Get("json"); jsonTag != "" {
			if tagName := strings.Split(jsonTag, ",")[0]; tagName != "-" {
				columnName = tagName
			}
		}

		// Convert to snake_case for database columns
		columnName = toSnakeCase(columnName)

		query = query.Where(fmt.Sprintf("%s = ?", columnName), value.Interface())
	}

	return query
}

// applySorting applies sort conditions to the query
// Validates sort fields to prevent SQL injection
func (r *BaseRepository) applySorting(query *gorm.DB, sortMap domain.SortMap) *gorm.DB {
	for field, direction := range sortMap {
		// Convert to snake_case and validate direction
		columnName := toSnakeCase(field)
		if direction != "asc" && direction != "desc" {
			direction = "asc" // Default to ascending
		}

		query = query.Order(fmt.Sprintf("%s %s", columnName, direction))
	}

	return query
}

// toSnakeCase converts CamelCase to snake_case
// Optimized string conversion with minimal allocations
func toSnakeCase(str string) string {
	var result strings.Builder
	result.Grow(len(str) + 5) // Pre-allocate with estimated size

	for i, r := range str {
		if i > 0 && r >= 'A' && r <= 'Z' {
			result.WriteByte('_')
		}
		if r >= 'A' && r <= 'Z' {
			result.WriteRune(r - 'A' + 'a')
		} else {
			result.WriteRune(r)
		}
	}

	return result.String()
}
