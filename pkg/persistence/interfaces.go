package persistence

import (
	"context"
	"github.com/arash-mosavi/postgrs-unit-of-work-system/pkg/domain"
	"github.com/arash-mosavi/postgrs-unit-of-work-system/pkg/identifier"
)

// IUnitOfWork defines the comprehensive Unit of Work pattern interface with generics
type IUnitOfWork[T domain.BaseModel] interface {
	// Transaction control
	BeginTransaction(ctx context.Context) error
	CommitTransaction(ctx context.Context) error
	RollbackTransaction(ctx context.Context)

	// Queries
	FindAll(ctx context.Context) ([]T, error)
	FindAllWithPagination(ctx context.Context, query domain.QueryParams[T]) ([]T, uint, error)
	FindOne(ctx context.Context, filter T) (T, error)
	FindOneById(ctx context.Context, id int) (T, error)
	FindOneByIdentifier(ctx context.Context, identifier identifier.IIdentifier) (T, error)
	ResolveIDByUniqueField(ctx context.Context, model domain.BaseModel, field string, value interface{}) (int, error)

	// Mutations
	Insert(ctx context.Context, entity T) (T, error)
	Update(ctx context.Context, identifier identifier.IIdentifier, entity T) (T, error)
	Delete(ctx context.Context, identifier identifier.IIdentifier) error

	// Soft & Hard Delete
	SoftDelete(ctx context.Context, identifier identifier.IIdentifier) (T, error)
	HardDelete(ctx context.Context, identifier identifier.IIdentifier) (T, error)

	// Bulk operations
	BulkInsert(ctx context.Context, entities []T) ([]T, error)
	BulkUpdate(ctx context.Context, entities []T) ([]T, error)
	BulkSoftDelete(ctx context.Context, identifiers []identifier.IIdentifier) error
	BulkHardDelete(ctx context.Context, identifiers []identifier.IIdentifier) error

	// Trashed Data
	GetTrashed(ctx context.Context) ([]T, error)
	GetTrashedWithPagination(ctx context.Context, query domain.QueryParams[T]) ([]T, uint, error)

	// Restore
	Restore(ctx context.Context, identifier identifier.IIdentifier) (T, error)
	RestoreAll(ctx context.Context) error
}

// IUnitOfWorkFactory creates Unit of Work instances with generics
type IUnitOfWorkFactory[T domain.BaseModel] interface {
	Create() IUnitOfWork[T]
	CreateWithContext(ctx context.Context) IUnitOfWork[T]
}
