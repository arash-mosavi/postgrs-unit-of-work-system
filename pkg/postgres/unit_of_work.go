package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"reflect"
	"sync"

	"unit-of-work/pkg/domain"
	"unit-of-work/pkg/identifier"
	"unit-of-work/pkg/persistence"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// UnitOfWork implements IUnitOfWork for PostgreSQL with generics
type UnitOfWork[T domain.BaseModel] struct {
	db           *gorm.DB
	tx           *gorm.DB
	ctx          context.Context
	repositories map[string]interface{}
	mu           sync.RWMutex
	inTx         bool
}

// NewUnitOfWork creates a new PostgreSQL unit of work
func NewUnitOfWork[T domain.BaseModel](config *Config) (*UnitOfWork[T], error) {
	db, err := gorm.Open(postgres.Open(config.DSN()), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	return &UnitOfWork[T]{
		db:           db,
		ctx:          context.Background(),
		repositories: make(map[string]interface{}),
	}, nil
}

// BeginTransaction starts a new database transaction
func (uow *UnitOfWork[T]) BeginTransaction(ctx context.Context) error {
	if uow.inTx {
		return fmt.Errorf("transaction already in progress")
	}

	tx := uow.db.WithContext(ctx).Begin(&sql.TxOptions{
		Isolation: sql.LevelReadCommitted,
		ReadOnly:  false,
	})

	if tx.Error != nil {
		return fmt.Errorf("failed to begin transaction: %w", tx.Error)
	}

	uow.tx = tx
	uow.ctx = ctx
	uow.inTx = true
	return nil
}

// CommitTransaction commits the current transaction
func (uow *UnitOfWork[T]) CommitTransaction(ctx context.Context) error {
	if !uow.inTx {
		return fmt.Errorf("no active transaction to commit")
	}

	if err := uow.tx.Commit().Error; err != nil {
		uow.RollbackTransaction(ctx)
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	uow.tx = nil
	uow.inTx = false
	return nil
}

// RollbackTransaction rolls back the current transaction
func (uow *UnitOfWork[T]) RollbackTransaction(ctx context.Context) {
	if !uow.inTx || uow.tx == nil {
		return
	}

	uow.tx.Rollback()
	uow.tx = nil
	uow.inTx = false
}

// FindAll retrieves all entities of type T
func (uow *UnitOfWork[T]) FindAll(ctx context.Context) ([]T, error) {
	var entities []T
	db := uow.getActiveDB()

	if err := db.Find(&entities).Error; err != nil {
		return nil, fmt.Errorf("failed to find all entities: %w", err)
	}

	return entities, nil
}

// FindAllWithPagination retrieves entities with pagination
func (uow *UnitOfWork[T]) FindAllWithPagination(ctx context.Context, query domain.QueryParams[T]) ([]T, uint, error) {
	var entities []T
	var total int64

	db := uow.getActiveDB()

	// Apply filters if provided
	if !reflect.ValueOf(query.Filter).IsZero() {
		db = db.Where(query.Filter)
	}

	// Count total records
	if err := db.Model(new(T)).Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count entities: %w", err)
	}

	// Apply sorting
	if query.Sort != nil {
		for field, direction := range query.Sort {
			db = db.Order(fmt.Sprintf("%s %s", field, direction))
		}
	}

	// Apply pagination
	if query.Limit > 0 {
		db = db.Limit(query.Limit)
	}
	if query.Offset > 0 {
		db = db.Offset(query.Offset)
	}

	// Apply includes (preloading)
	for _, include := range query.Include {
		db = db.Preload(include)
	}

	if err := db.Find(&entities).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to find entities with pagination: %w", err)
	}

	return entities, uint(total), nil
}

// FindOne retrieves a single entity by filter
func (uow *UnitOfWork[T]) FindOne(ctx context.Context, filter T) (T, error) {
	var entity T
	db := uow.getActiveDB()

	if err := db.Where(filter).First(&entity).Error; err != nil {
		return entity, fmt.Errorf("failed to find entity: %w", err)
	}

	return entity, nil
}

// FindOneById retrieves a single entity by ID
func (uow *UnitOfWork[T]) FindOneById(ctx context.Context, id int) (T, error) {
	var entity T
	db := uow.getActiveDB()

	if err := db.First(&entity, id).Error; err != nil {
		return entity, fmt.Errorf("failed to find entity by id: %w", err)
	}

	return entity, nil
}

// FindOneByIdentifier retrieves a single entity by identifier
func (uow *UnitOfWork[T]) FindOneByIdentifier(ctx context.Context, identifier identifier.IIdentifier) (T, error) {
	var entity T
	db := uow.getActiveDB()

	queryMap := identifier.ToMap()
	if err := db.Where(queryMap).First(&entity).Error; err != nil {
		return entity, fmt.Errorf("failed to find entity by identifier: %w", err)
	}

	return entity, nil
}

// ResolveIDByUniqueField resolves an ID by a unique field
func (uow *UnitOfWork[T]) ResolveIDByUniqueField(ctx context.Context, model domain.BaseModel, field string, value interface{}) (int, error) {
	var entity T
	db := uow.getActiveDB()

	if err := db.Where(field+" = ?", value).First(&entity).Error; err != nil {
		return 0, fmt.Errorf("failed to resolve ID by unique field: %w", err)
	}

	return entity.GetID(), nil
}

// Insert creates a new entity
func (uow *UnitOfWork[T]) Insert(ctx context.Context, entity T) (T, error) {
	db := uow.getActiveDB()

	if err := db.Create(&entity).Error; err != nil {
		return entity, fmt.Errorf("failed to insert entity: %w", err)
	}

	return entity, nil
}

// Update updates an existing entity
func (uow *UnitOfWork[T]) Update(ctx context.Context, identifier identifier.IIdentifier, entity T) (T, error) {
	db := uow.getActiveDB()

	queryMap := identifier.ToMap()
	if err := db.Where(queryMap).Updates(&entity).Error; err != nil {
		return entity, fmt.Errorf("failed to update entity: %w", err)
	}

	// Retrieve the updated entity
	var updatedEntity T
	if err := db.Where(queryMap).First(&updatedEntity).Error; err != nil {
		return entity, fmt.Errorf("failed to retrieve updated entity: %w", err)
	}

	return updatedEntity, nil
}

// Delete removes an entity (hard delete)
func (uow *UnitOfWork[T]) Delete(ctx context.Context, identifier identifier.IIdentifier) error {
	db := uow.getActiveDB()

	queryMap := identifier.ToMap()
	if err := db.Unscoped().Where(queryMap).Delete(new(T)).Error; err != nil {
		return fmt.Errorf("failed to delete entity: %w", err)
	}

	return nil
}

// SoftDelete performs a soft delete on an entity
func (uow *UnitOfWork[T]) SoftDelete(ctx context.Context, identifier identifier.IIdentifier) (T, error) {
	var entity T
	db := uow.getActiveDB()

	queryMap := identifier.ToMap()

	// First find the entity
	if err := db.Where(queryMap).First(&entity).Error; err != nil {
		return entity, fmt.Errorf("failed to find entity for soft delete: %w", err)
	}

	// Perform soft delete
	if err := db.Where(queryMap).Delete(&entity).Error; err != nil {
		return entity, fmt.Errorf("failed to soft delete entity: %w", err)
	}

	return entity, nil
}

// HardDelete performs a hard delete on an entity
func (uow *UnitOfWork[T]) HardDelete(ctx context.Context, identifier identifier.IIdentifier) (T, error) {
	var entity T
	db := uow.getActiveDB()

	queryMap := identifier.ToMap()

	// First find the entity
	if err := db.Where(queryMap).First(&entity).Error; err != nil {
		return entity, fmt.Errorf("failed to find entity for hard delete: %w", err)
	}

	// Perform hard delete
	if err := db.Unscoped().Where(queryMap).Delete(&entity).Error; err != nil {
		return entity, fmt.Errorf("failed to hard delete entity: %w", err)
	}

	return entity, nil
}

// BulkInsert creates multiple entities
func (uow *UnitOfWork[T]) BulkInsert(ctx context.Context, entities []T) ([]T, error) {
	db := uow.getActiveDB()

	if err := db.CreateInBatches(&entities, 100).Error; err != nil {
		return nil, fmt.Errorf("failed to bulk insert entities: %w", err)
	}

	return entities, nil
}

// BulkUpdate updates multiple entities
func (uow *UnitOfWork[T]) BulkUpdate(ctx context.Context, entities []T) ([]T, error) {
	db := uow.getActiveDB()

	for i := range entities {
		if err := db.Save(&entities[i]).Error; err != nil {
			return nil, fmt.Errorf("failed to bulk update entity at index %d: %w", i, err)
		}
	}

	return entities, nil
}

// BulkSoftDelete performs soft delete on multiple entities
func (uow *UnitOfWork[T]) BulkSoftDelete(ctx context.Context, identifiers []identifier.IIdentifier) error {
	db := uow.getActiveDB()

	for _, id := range identifiers {
		queryMap := id.ToMap()
		if err := db.Where(queryMap).Delete(new(T)).Error; err != nil {
			return fmt.Errorf("failed to bulk soft delete entity: %w", err)
		}
	}

	return nil
}

// BulkHardDelete performs hard delete on multiple entities
func (uow *UnitOfWork[T]) BulkHardDelete(ctx context.Context, identifiers []identifier.IIdentifier) error {
	db := uow.getActiveDB()

	for _, id := range identifiers {
		queryMap := id.ToMap()
		if err := db.Unscoped().Where(queryMap).Delete(new(T)).Error; err != nil {
			return fmt.Errorf("failed to bulk hard delete entity: %w", err)
		}
	}

	return nil
}

// GetTrashed retrieves all soft-deleted entities
func (uow *UnitOfWork[T]) GetTrashed(ctx context.Context) ([]T, error) {
	var entities []T
	db := uow.getActiveDB()

	if err := db.Unscoped().Where("deleted_at IS NOT NULL").Find(&entities).Error; err != nil {
		return nil, fmt.Errorf("failed to get trashed entities: %w", err)
	}

	return entities, nil
}

// GetTrashedWithPagination retrieves soft-deleted entities with pagination
func (uow *UnitOfWork[T]) GetTrashedWithPagination(ctx context.Context, query domain.QueryParams[T]) ([]T, uint, error) {
	var entities []T
	var total int64

	db := uow.getActiveDB().Unscoped().Where("deleted_at IS NOT NULL")

	// Apply filters if provided
	if !reflect.ValueOf(query.Filter).IsZero() {
		db = db.Where(query.Filter)
	}

	// Count total records
	if err := db.Model(new(T)).Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count trashed entities: %w", err)
	}

	// Apply sorting
	if query.Sort != nil {
		for field, direction := range query.Sort {
			db = db.Order(fmt.Sprintf("%s %s", field, direction))
		}
	}

	// Apply pagination
	if query.Limit > 0 {
		db = db.Limit(query.Limit)
	}
	if query.Offset > 0 {
		db = db.Offset(query.Offset)
	}

	if err := db.Find(&entities).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to get trashed entities with pagination: %w", err)
	}

	return entities, uint(total), nil
}

// Restore restores a soft-deleted entity
func (uow *UnitOfWork[T]) Restore(ctx context.Context, identifier identifier.IIdentifier) (T, error) {
	var entity T
	db := uow.getActiveDB()

	queryMap := identifier.ToMap()

	// Find the soft-deleted entity
	if err := db.Unscoped().Where(queryMap).Where("deleted_at IS NOT NULL").First(&entity).Error; err != nil {
		return entity, fmt.Errorf("failed to find trashed entity: %w", err)
	}

	// Restore the entity
	if err := db.Unscoped().Model(&entity).Update("deleted_at", nil).Error; err != nil {
		return entity, fmt.Errorf("failed to restore entity: %w", err)
	}

	return entity, nil
}

// RestoreAll restores all soft-deleted entities
func (uow *UnitOfWork[T]) RestoreAll(ctx context.Context) error {
	db := uow.getActiveDB()

	if err := db.Unscoped().Model(new(T)).Where("deleted_at IS NOT NULL").Update("deleted_at", nil).Error; err != nil {
		return fmt.Errorf("failed to restore all entities: %w", err)
	}

	return nil
}

// GetRepository returns a repository for the specified entity type
func (uow *UnitOfWork[T]) GetRepository(entityType string) interface{} {
	uow.mu.RLock()
	repo, exists := uow.repositories[entityType]
	uow.mu.RUnlock()

	if exists {
		return repo
	}

	uow.mu.Lock()
	defer uow.mu.Unlock()

	if repo, exists := uow.repositories[entityType]; exists {
		return repo
	}

	repo = NewBaseRepository(uow.getActiveDB())
	uow.repositories[entityType] = repo
	return repo
}

// RegisterRepository registers a custom repository for a specific entity type
func (uow *UnitOfWork[T]) RegisterRepository(entityType string, repo interface{}) {
	uow.mu.Lock()
	defer uow.mu.Unlock()
	uow.repositories[entityType] = repo
}

// WithContext creates a new unit of work with the specified context
func (uow *UnitOfWork[T]) WithContext(ctx context.Context) persistence.IUnitOfWork[T] {
	newUow := &UnitOfWork[T]{
		db:           uow.db,
		tx:           uow.tx,
		ctx:          ctx,
		repositories: uow.repositories,
		inTx:         uow.inTx,
	}
	return newUow
}

// GetContext returns the current context
func (uow *UnitOfWork[T]) GetContext() context.Context {
	return uow.ctx
}

// IsInTransaction checks if a transaction is currently active
func (uow *UnitOfWork[T]) IsInTransaction() bool {
	return uow.inTx
}

// Close closes the database connection
func (uow *UnitOfWork[T]) Close() error {
	if uow.inTx {
		uow.RollbackTransaction(uow.ctx)
	}

	sqlDB, err := uow.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

// getActiveDB returns the appropriate database connection
func (uow *UnitOfWork[T]) getActiveDB() *gorm.DB {
	if uow.inTx && uow.tx != nil {
		return uow.tx
	}
	return uow.db.WithContext(uow.ctx)
}
