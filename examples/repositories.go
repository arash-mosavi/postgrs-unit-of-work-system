package examples

import (
	"context"

	"github.com/arash-mosavi/postgrs-unit-of-work-system/pkg/domain"
	"github.com/arash-mosavi/postgrs-unit-of-work-system/pkg/identifier"
	"github.com/arash-mosavi/postgrs-unit-of-work-system/pkg/persistence"
)

// IUserRepository defines user-specific repository operations
type IUserRepository interface {
	// Basic CRUD
	Create(ctx context.Context, user *User) (*User, error)
	GetByID(ctx context.Context, id int) (*User, error)
	GetByEmail(ctx context.Context, email string) (*User, error)
	Update(ctx context.Context, user *User) (*User, error)
	Delete(ctx context.Context, id int) error

	// Queries
	FindAll(ctx context.Context) ([]*User, error)
	FindWithPagination(ctx context.Context, page, pageSize int) ([]*User, uint, error)
	Search(ctx context.Context, filter *User, limit int) ([]*User, error)

	// Batch operations
	BatchCreate(ctx context.Context, users []*User) ([]*User, error)
	BatchUpdate(ctx context.Context, users []*User) ([]*User, error)

	// Soft delete operations
	SoftDelete(ctx context.Context, id int) (*User, error)
	GetTrashed(ctx context.Context) ([]*User, error)
	Restore(ctx context.Context, id int) (*User, error)
}

// UserRepository implements IUserRepository using Unit of Work
type UserRepository struct {
	uow persistence.IUnitOfWork[*User]
}

// NewUserRepository creates a new user repository
func NewUserRepository(uow persistence.IUnitOfWork[*User]) IUserRepository {
	return &UserRepository{
		uow: uow,
	}
}

// Create inserts a new user
func (r *UserRepository) Create(ctx context.Context, user *User) (*User, error) {
	return r.uow.Insert(ctx, user)
}

// GetByID retrieves a user by ID
func (r *UserRepository) GetByID(ctx context.Context, id int) (*User, error) {
	return r.uow.FindOneById(ctx, id)
}

// GetByEmail retrieves a user by email address
func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*User, error) {
	emailIdentifier := identifier.NewIdentifier().Equal("email", email)
	return r.uow.FindOneByIdentifier(ctx, emailIdentifier)
}

// Update modifies an existing user
func (r *UserRepository) Update(ctx context.Context, user *User) (*User, error) {
	idIdentifier := identifier.NewIdentifier().Equal("id", user.ID)
	return r.uow.Update(ctx, idIdentifier, user)
}

// Delete removes a user (hard delete)
func (r *UserRepository) Delete(ctx context.Context, id int) error {
	idIdentifier := identifier.NewIdentifier().Equal("id", id)
	return r.uow.Delete(ctx, idIdentifier)
}

// FindAll retrieves all users
func (r *UserRepository) FindAll(ctx context.Context) ([]*User, error) {
	return r.uow.FindAll(ctx)
}

// FindWithPagination retrieves users with pagination
func (r *UserRepository) FindWithPagination(ctx context.Context, page, pageSize int) ([]*User, uint, error) {
	params := domain.QueryParams[*User]{
		Sort: domain.SortMap{
			"created_at": domain.SortDesc,
			"name":       domain.SortAsc,
		},
		Limit:  pageSize,
		Offset: (page - 1) * pageSize,
	}

	return r.uow.FindAllWithPagination(ctx, params)
}

// Search finds users matching the given criteria
func (r *UserRepository) Search(ctx context.Context, filter *User, limit int) ([]*User, error) {
	params := domain.QueryParams[*User]{
		Filter: filter,
		Sort: domain.SortMap{
			"name":       domain.SortAsc,
			"created_at": domain.SortDesc,
		},
		Limit: limit,
	}

	users, _, err := r.uow.FindAllWithPagination(ctx, params)
	return users, err
}

// BatchCreate performs bulk user creation
func (r *UserRepository) BatchCreate(ctx context.Context, users []*User) ([]*User, error) {
	return r.uow.BulkInsert(ctx, users)
}

// BatchUpdate performs bulk user updates
func (r *UserRepository) BatchUpdate(ctx context.Context, users []*User) ([]*User, error) {
	return r.uow.BulkUpdate(ctx, users)
}

// SoftDelete soft deletes a user
func (r *UserRepository) SoftDelete(ctx context.Context, id int) (*User, error) {
	idIdentifier := identifier.NewIdentifier().Equal("id", id)
	return r.uow.SoftDelete(ctx, idIdentifier)
}

// GetTrashed retrieves all soft-deleted users
func (r *UserRepository) GetTrashed(ctx context.Context) ([]*User, error) {
	return r.uow.GetTrashed(ctx)
}

// Restore restores a soft-deleted user
func (r *UserRepository) Restore(ctx context.Context, id int) (*User, error) {
	idIdentifier := identifier.NewIdentifier().Equal("id", id)
	return r.uow.Restore(ctx, idIdentifier)
}

// IPostRepository defines post-specific repository operations
type IPostRepository interface {
	Create(ctx context.Context, post *Post) (*Post, error)
	GetByID(ctx context.Context, id int) (*Post, error)
	GetByUserID(ctx context.Context, userID int) ([]*Post, error)
	Update(ctx context.Context, post *Post) (*Post, error)
	Delete(ctx context.Context, id int) error
	BatchCreate(ctx context.Context, posts []*Post) ([]*Post, error)
}

// PostRepository implements IPostRepository using Unit of Work
type PostRepository struct {
	uow persistence.IUnitOfWork[*Post]
}

// NewPostRepository creates a new post repository
func NewPostRepository(uow persistence.IUnitOfWork[*Post]) IPostRepository {
	return &PostRepository{
		uow: uow,
	}
}

// Create inserts a new post
func (r *PostRepository) Create(ctx context.Context, post *Post) (*Post, error) {
	return r.uow.Insert(ctx, post)
}

// GetByID retrieves a post by ID
func (r *PostRepository) GetByID(ctx context.Context, id int) (*Post, error) {
	return r.uow.FindOneById(ctx, id)
}

// GetByUserID retrieves all posts for a specific user
func (r *PostRepository) GetByUserID(ctx context.Context, userID int) ([]*Post, error) {
	filter := &Post{UserID: userID}
	params := domain.QueryParams[*Post]{
		Filter: filter,
		Sort: domain.SortMap{
			"created_at": domain.SortDesc,
		},
	}

	posts, _, err := r.uow.FindAllWithPagination(ctx, params)
	return posts, err
}

// Update modifies an existing post
func (r *PostRepository) Update(ctx context.Context, post *Post) (*Post, error) {
	idIdentifier := identifier.NewIdentifier().Equal("id", post.ID)
	return r.uow.Update(ctx, idIdentifier, post)
}

// Delete removes a post
func (r *PostRepository) Delete(ctx context.Context, id int) error {
	idIdentifier := identifier.NewIdentifier().Equal("id", id)
	return r.uow.Delete(ctx, idIdentifier)
}

// BatchCreate performs bulk post creation
func (r *PostRepository) BatchCreate(ctx context.Context, posts []*Post) ([]*Post, error) {
	return r.uow.BulkInsert(ctx, posts)
}
