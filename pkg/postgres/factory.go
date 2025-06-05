package postgres

import (
	"context"
	"unit-of-work/pkg/domain"
	"unit-of-work/pkg/persistence"
)

// UnitOfWorkFactory implements IUnitOfWorkFactory for PostgreSQL with generics
type UnitOfWorkFactory[T domain.BaseModel] struct {
	Config *Config
}

// NewUnitOfWorkFactory creates a new PostgreSQL unit of work factory
func NewUnitOfWorkFactory[T domain.BaseModel](config *Config) *UnitOfWorkFactory[T] {
	return &UnitOfWorkFactory[T]{
		Config: config,
	}
}

// Create creates a new unit of work instance
func (f *UnitOfWorkFactory[T]) Create() persistence.IUnitOfWork[T] {
	uow, err := NewUnitOfWork[T](f.Config)
	if err != nil {
		// In a production environment, you might want to handle this differently
		panic(err)
	}
	return uow
}

// CreateWithContext creates a new unit of work instance with context
func (f *UnitOfWorkFactory[T]) CreateWithContext(ctx context.Context) persistence.IUnitOfWork[T] {
	uow, err := NewUnitOfWork[T](f.Config)
	if err != nil {
		// In a production environment, you might want to handle this differently
		panic(err)
	}
	uow.ctx = ctx
	return uow
}
