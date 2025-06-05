package examples

import (
	"context"
	"fmt"
	"log"

	"github.com/arash-mosavi/postgrs-unit-of-work-system/pkg/persistence"
	"github.com/arash-mosavi/postgrs-unit-of-work-system/pkg/postgres"
)

type UserService struct {
	uowFactory  persistence.IUnitOfWorkFactory[*User]
	postFactory persistence.IUnitOfWorkFactory[*Post]
}

func NewUserService(
	userFactory persistence.IUnitOfWorkFactory[*User],
	postFactory persistence.IUnitOfWorkFactory[*Post],
) *UserService {
	return &UserService{
		uowFactory:  userFactory,
		postFactory: postFactory,
	}
}

func (s *UserService) CreateUserWithPosts(ctx context.Context, user *User, posts []*Post) error {

	userUow := s.uowFactory.CreateWithContext(ctx)
	postUow := s.postFactory.CreateWithContext(ctx)

	userRepo := NewUserRepository(userUow)
	postRepo := NewPostRepository(postUow)

	if err := userUow.BeginTransaction(ctx); err != nil {
		return fmt.Errorf("failed to begin user transaction: %w", err)
	}

	defer func() {
		if r := recover(); r != nil {
			userUow.RollbackTransaction(ctx)
			panic(r)
		}
	}()

	createdUser, err := userRepo.Create(ctx, user)
	if err != nil {
		userUow.RollbackTransaction(ctx)
		return fmt.Errorf("failed to create user: %w", err)
	}

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

	for _, post := range posts {
		post.UserID = createdUser.ID
		if _, err := postRepo.Create(ctx, post); err != nil {
			postUow.RollbackTransaction(ctx)
			userUow.RollbackTransaction(ctx)
			return fmt.Errorf("failed to create post: %w", err)
		}
	}

	if err := postUow.CommitTransaction(ctx); err != nil {
		userUow.RollbackTransaction(ctx)
		return fmt.Errorf("failed to commit post transaction: %w", err)
	}

	if err := userUow.CommitTransaction(ctx); err != nil {
		return fmt.Errorf("failed to commit user transaction: %w", err)
	}

	return nil
}

func (s *UserService) ListUsers(ctx context.Context, page, pageSize int) ([]*User, uint, error) {

	uow := s.uowFactory.CreateWithContext(ctx)
	userRepo := NewUserRepository(uow)

	return userRepo.FindWithPagination(ctx, page, pageSize)
}

func (s *UserService) FindUserByEmail(ctx context.Context, email string) (*User, error) {

	uow := s.uowFactory.CreateWithContext(ctx)
	userRepo := NewUserRepository(uow)

	return userRepo.GetByEmail(ctx, email)
}

func (s *UserService) SearchUsers(ctx context.Context, name, email string, activeOnly bool) ([]*User, error) {

	uow := s.uowFactory.CreateWithContext(ctx)
	userRepo := NewUserRepository(uow)

	filter := &User{}
	if name != "" {
		filter.Name = name
	}
	if email != "" {
		filter.Email = email
	}

	return userRepo.Search(ctx, filter, 50)
}

func (s *UserService) BatchCreateUsers(ctx context.Context, users []*User) ([]*User, error) {

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

func (s *UserService) SoftDeleteUser(ctx context.Context, userID int) (*User, error) {

	uow := s.uowFactory.CreateWithContext(ctx)
	userRepo := NewUserRepository(uow)

	return userRepo.SoftDelete(ctx, userID)
}

func (s *UserService) GetTrashedUsers(ctx context.Context) ([]*User, error) {

	uow := s.uowFactory.CreateWithContext(ctx)
	userRepo := NewUserRepository(uow)

	return userRepo.GetTrashed(ctx)
}

func (s *UserService) RestoreUser(ctx context.Context, userID int) (*User, error) {

	uow := s.uowFactory.CreateWithContext(ctx)
	userRepo := NewUserRepository(uow)

	return userRepo.Restore(ctx, userID)
}

type PostService struct {
	uowFactory persistence.IUnitOfWorkFactory[*Post]
}

func NewPostService(factory persistence.IUnitOfWorkFactory[*Post]) *PostService {
	return &PostService{
		uowFactory: factory,
	}
}

func (s *PostService) GetUserPosts(ctx context.Context, userID int) ([]*Post, error) {

	uow := s.uowFactory.CreateWithContext(ctx)
	postRepo := NewPostRepository(uow)

	return postRepo.GetByUserID(ctx, userID)
}

func (s *PostService) BatchCreatePosts(ctx context.Context, posts []*Post) ([]*Post, error) {

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

func Example() {

	config := postgres.NewConfig()
	config.Host = "localhost"
	config.Port = 5432
	config.User = "postgres"
	config.Password = "password"
	config.Database = "testdb"
	config.SSLMode = "disable"

	userFactory := postgres.NewUnitOfWorkFactory[*User](config)
	postFactory := postgres.NewUnitOfWorkFactory[*Post](config)

	userService := NewUserService(userFactory, postFactory)
	postService := NewPostService(postFactory)

	ctx := context.Background()

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

	users, total, err := userService.ListUsers(ctx, 1, 10)
	if err != nil {
		log.Printf("Failed to list users: %v", err)
		return
	}

	log.Printf("Found %d users (total: %d)", len(users), total)

	searchResults, err := userService.SearchUsers(ctx, "John", "", true)
	if err != nil {
		log.Printf("Failed to search users: %v", err)
		return
	}

	log.Printf("Search returned %d users", len(searchResults))

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

	if len(createdUsers) > 0 {
		userToDelete := createdUsers[0]

		deletedUser, err := userService.SoftDeleteUser(ctx, userToDelete.ID)
		if err != nil {
			log.Printf("Failed to soft delete user: %v", err)
			return
		}
		log.Printf("Soft deleted user: %s", deletedUser.Name)

		trashedUsers, err := userService.GetTrashedUsers(ctx)
		if err != nil {
			log.Printf("Failed to get trashed users: %v", err)
			return
		}
		log.Printf("Found %d trashed users", len(trashedUsers))

		restoredUser, err := userService.RestoreUser(ctx, deletedUser.ID)
		if err != nil {
			log.Printf("Failed to restore user: %v", err)
			return
		}
		log.Printf("Restored user: %s", restoredUser.Name)
	}

	foundUser, err := userService.FindUserByEmail(ctx, "john@example.com")
	if err != nil {
		log.Printf("Failed to find user by email: %v", err)
		return
	}
	log.Printf("Found user by email: %s", foundUser.Name)

	if len(createdUsers) > 0 {
		userPosts, err := postService.GetUserPosts(ctx, createdUsers[0].ID)
		if err != nil {
			log.Printf("Failed to get user posts: %v", err)
			return
		}
		log.Printf("Found %d posts for user", len(userPosts))

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
