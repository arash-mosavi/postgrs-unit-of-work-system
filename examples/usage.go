package examples

import (
	"context"
	"fmt"
	"log"

	"github.com/arash-mosavi/postgrs-unit-of-work-system/pkg/persistence"
	"github.com/arash-mosavi/postgrs-unit-of-work-system/pkg/postgres"
)

// UserService demonstrates enterprise-level service implementation following proper architectural flow:
// Service -> Repository -> BaseRepository -> Unit of Work -> Database
type UserService struct {
	uowFactory  persistence.IUnitOfWorkFactory[*User]
	postFactory persistence.IUnitOfWorkFactory[*Post]
}

// NewUserService creates a new user service with dependency injection
func NewUserService(
	userFactory persistence.IUnitOfWorkFactory[*User],
	postFactory persistence.IUnitOfWorkFactory[*Post],
) *UserService {
	return &UserService{
		uowFactory:  userFactory,
		postFactory: postFactory,
	}
}

// CreateUserWithPosts demonstrates complex transaction with multiple entities
// Following the architectural flow: Service -> Repository -> Unit of Work -> Database
func (s *UserService) CreateUserWithPosts(ctx context.Context, user *User, posts []*Post) error {
	// Create Unit of Work instances for both entities
	userUow := s.uowFactory.CreateWithContext(ctx)
	postUow := s.postFactory.CreateWithContext(ctx)

	// Create repositories using Unit of Work instances
	userRepo := NewUserRepository(userUow)
	postRepo := NewPostRepository(postUow)

	// Begin transaction on both UoWs (this would ideally be coordinated)
	if err := userUow.BeginTransaction(ctx); err != nil {
		return fmt.Errorf("failed to begin user transaction: %w", err)
	}

	defer func() {
		if r := recover(); r != nil {
			userUow.RollbackTransaction(ctx)
			panic(r)
		}
	}()

	// Service -> Repository -> Unit of Work -> Database
	createdUser, err := userRepo.Create(ctx, user)
	if err != nil {
		userUow.RollbackTransaction(ctx)
		return fmt.Errorf("failed to create user: %w", err)
	}

	// Begin post transaction
	if err := postUow.BeginTransaction(ctx); err != nil {
		userUow.RollbackTransaction(ctx)
		return fmt.Errorf("failed to begin post transaction: %w", err)
	}

	defer func() {
		if r := recover(); r != nil {
			postUow.RollbackTransaction(ctx)
			userUow.RollbackTransaction(ctx)
			panic(r)
		}
	}()

	// Create associated posts through repository layer
	for _, post := range posts {
		post.UserID = createdUser.ID // Set foreign key
		if _, err := postRepo.Create(ctx, post); err != nil {
			postUow.RollbackTransaction(ctx)
			userUow.RollbackTransaction(ctx)
			return fmt.Errorf("failed to create post: %w", err)
		}
	}

	// Commit both transactions
	if err := postUow.CommitTransaction(ctx); err != nil {
		userUow.RollbackTransaction(ctx)
		return fmt.Errorf("failed to commit post transaction: %w", err)
	}

	if err := userUow.CommitTransaction(ctx); err != nil {
		return fmt.Errorf("failed to commit user transaction: %w", err)
	}

	return nil
}

// ListUsers demonstrates querying with pagination through repository layer
func (s *UserService) ListUsers(ctx context.Context, page, pageSize int) ([]*User, uint, error) {
	// Service -> Repository -> Unit of Work -> Database
	uow := s.uowFactory.CreateWithContext(ctx)
	userRepo := NewUserRepository(uow)

	return userRepo.FindWithPagination(ctx, page, pageSize)
}

// FindUserByEmail demonstrates identifier-based queries through repository layer
func (s *UserService) FindUserByEmail(ctx context.Context, email string) (*User, error) {
	// Service -> Repository -> Unit of Work -> Database
	uow := s.uowFactory.CreateWithContext(ctx)
	userRepo := NewUserRepository(uow)

	return userRepo.GetByEmail(ctx, email)
}

// SearchUsers demonstrates complex querying with multiple conditions through repository layer
func (s *UserService) SearchUsers(ctx context.Context, name, email string, activeOnly bool) ([]*User, error) {
	// Service -> Repository -> Unit of Work -> Database
	uow := s.uowFactory.CreateWithContext(ctx)
	userRepo := NewUserRepository(uow)

	// Build filter based on criteria
	filter := &User{}
	if name != "" {
		filter.Name = name
	}
	if email != "" {
		filter.Email = email
	}

	return userRepo.Search(ctx, filter, 50)
}

// BatchCreateUsers demonstrates bulk operations for performance through repository layer
func (s *UserService) BatchCreateUsers(ctx context.Context, users []*User) ([]*User, error) {
	// Service -> Repository -> Unit of Work -> Database
	uow := s.uowFactory.CreateWithContext(ctx)
	userRepo := NewUserRepository(uow)

	if err := uow.BeginTransaction(ctx); err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer func() {
		if r := recover(); r != nil {
			uow.RollbackTransaction(ctx)
			panic(r)
		}
	}()

	// Use bulk insert through repository for better performance
	createdUsers, err := userRepo.BatchCreate(ctx, users)
	if err != nil {
		uow.RollbackTransaction(ctx)
		return nil, fmt.Errorf("failed to batch create users: %w", err)
	}

	if err := uow.CommitTransaction(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return createdUsers, nil
}

// SoftDeleteUser demonstrates soft delete functionality through repository layer
func (s *UserService) SoftDeleteUser(ctx context.Context, userID int) (*User, error) {
	// Service -> Repository -> Unit of Work -> Database
	uow := s.uowFactory.CreateWithContext(ctx)
	userRepo := NewUserRepository(uow)

	return userRepo.SoftDelete(ctx, userID)
}

// GetTrashedUsers demonstrates retrieving soft-deleted data through repository layer
func (s *UserService) GetTrashedUsers(ctx context.Context) ([]*User, error) {
	// Service -> Repository -> Unit of Work -> Database
	uow := s.uowFactory.CreateWithContext(ctx)
	userRepo := NewUserRepository(uow)

	return userRepo.GetTrashed(ctx)
}

// RestoreUser demonstrates data restoration functionality through repository layer
func (s *UserService) RestoreUser(ctx context.Context, userID int) (*User, error) {
	// Service -> Repository -> Unit of Work -> Database
	uow := s.uowFactory.CreateWithContext(ctx)
	userRepo := NewUserRepository(uow)

	return userRepo.Restore(ctx, userID)
}

// PostService demonstrates multi-entity service patterns following proper architectural flow
type PostService struct {
	uowFactory persistence.IUnitOfWorkFactory[*Post]
}

// NewPostService creates a new post service
func NewPostService(factory persistence.IUnitOfWorkFactory[*Post]) *PostService {
	return &PostService{
		uowFactory: factory,
	}
}

// GetUserPosts demonstrates querying posts by user ID through repository layer
func (s *PostService) GetUserPosts(ctx context.Context, userID int) ([]*Post, error) {
	// Service -> Repository -> Unit of Work -> Database
	uow := s.uowFactory.CreateWithContext(ctx)
	postRepo := NewPostRepository(uow)

	return postRepo.GetByUserID(ctx, userID)
}

// BatchCreatePosts demonstrates bulk post creation through repository layer
func (s *PostService) BatchCreatePosts(ctx context.Context, posts []*Post) ([]*Post, error) {
	// Service -> Repository -> Unit of Work -> Database
	uow := s.uowFactory.CreateWithContext(ctx)
	postRepo := NewPostRepository(uow)

	if err := uow.BeginTransaction(ctx); err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer func() {
		if r := recover(); r != nil {
			uow.RollbackTransaction(ctx)
			panic(r)
		}
	}()

	createdPosts, err := postRepo.BatchCreate(ctx, posts)
	if err != nil {
		uow.RollbackTransaction(ctx)
		return nil, fmt.Errorf("failed to batch create posts: %w", err)
	}

	if err := uow.CommitTransaction(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return createdPosts, nil
}

// Example demonstrates complete usage of the Unit of Work pattern with proper architectural flow
func Example() {
	// Initialize PostgreSQL configuration
	config := postgres.NewConfig()
	config.Host = "localhost"
	config.Port = 5432
	config.User = "postgres"
	config.Password = "password"
	config.Database = "testdb"
	config.SSLMode = "disable"

	// Create Unit of Work factories with explicit type parameters
	userFactory := postgres.NewUnitOfWorkFactory[*User](config)
	postFactory := postgres.NewUnitOfWorkFactory[*Post](config)

	// Create services with dependency injection
	userService := NewUserService(userFactory, postFactory)
	postService := NewPostService(postFactory)

	ctx := context.Background()

	// Example 1: Complex transaction with multiple entities
	user := &User{
		Name:  "John Doe",
		Email: "john@example.com",
		Slug:  "john-doe",
	}

	posts := []*Post{
		{Name: "First Post", Content: "Hello World", Slug: "first-post"},
		{Name: "Second Post", Content: "Learning Go", Slug: "second-post"},
	}

	if err := userService.CreateUserWithPosts(ctx, user, posts); err != nil {
		log.Printf("Failed to create user with posts: %v", err)
		return
	}

	log.Printf("Created user %s with %d posts", user.Name, len(posts))

	// Example 2: Query with pagination through repository layer
	users, total, err := userService.ListUsers(ctx, 1, 10)
	if err != nil {
		log.Printf("Failed to list users: %v", err)
		return
	}

	log.Printf("Found %d users (total: %d)", len(users), total)

	// Example 3: Search with complex conditions through repository layer
	searchResults, err := userService.SearchUsers(ctx, "John", "", true)
	if err != nil {
		log.Printf("Failed to search users: %v", err)
		return
	}

	log.Printf("Search returned %d users", len(searchResults))

	// Example 4: Batch operations through repository layer
	batchUsers := []*User{
		{Name: "Alice", Email: "alice@example.com", Slug: "alice"},
		{Name: "Bob", Email: "bob@example.com", Slug: "bob"},
		{Name: "Charlie", Email: "charlie@example.com", Slug: "charlie"},
	}

	createdUsers, err := userService.BatchCreateUsers(ctx, batchUsers)
	if err != nil {
		log.Printf("Failed to batch create users: %v", err)
		return
	}

	log.Printf("Successfully created %d users in batch", len(createdUsers))

	// Example 5: Soft delete and restore through repository layer
	if len(createdUsers) > 0 {
		userToDelete := createdUsers[0]

		// Soft delete
		deletedUser, err := userService.SoftDeleteUser(ctx, userToDelete.ID)
		if err != nil {
			log.Printf("Failed to soft delete user: %v", err)
			return
		}
		log.Printf("Soft deleted user: %s", deletedUser.Name)

		// Get trashed users
		trashedUsers, err := userService.GetTrashedUsers(ctx)
		if err != nil {
			log.Printf("Failed to get trashed users: %v", err)
			return
		}
		log.Printf("Found %d trashed users", len(trashedUsers))

		// Restore user
		restoredUser, err := userService.RestoreUser(ctx, deletedUser.ID)
		if err != nil {
			log.Printf("Failed to restore user: %v", err)
			return
		}
		log.Printf("Restored user: %s", restoredUser.Name)
	}

	// Example 6: Finding user by email through repository layer
	foundUser, err := userService.FindUserByEmail(ctx, "john@example.com")
	if err != nil {
		log.Printf("Failed to find user by email: %v", err)
		return
	}
	log.Printf("Found user by email: %s", foundUser.Name)

	// Example 7: Working with posts through repository layer
	if len(createdUsers) > 0 {
		userPosts, err := postService.GetUserPosts(ctx, createdUsers[0].ID)
		if err != nil {
			log.Printf("Failed to get user posts: %v", err)
			return
		}
		log.Printf("Found %d posts for user", len(userPosts))

		// Batch create additional posts
		additionalPosts := []*Post{
			{Name: "Post 1", Content: "Content 1", Slug: "post-1", UserID: createdUsers[0].ID},
			{Name: "Post 2", Content: "Content 2", Slug: "post-2", UserID: createdUsers[0].ID},
		}

		createdPosts, err := postService.BatchCreatePosts(ctx, additionalPosts)
		if err != nil {
			log.Printf("Failed to batch create posts: %v", err)
			return
		}
		log.Printf("Successfully created %d posts in batch", len(createdPosts))
	}

	log.Println("All examples completed successfully following the architectural flow:")
	log.Println("Service -> Repository -> BaseRepository -> Unit of Work -> Database")
}
