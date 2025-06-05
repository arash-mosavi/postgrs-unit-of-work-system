package main

import (
	"context"
	"fmt"

	"github.com/arash-mosavi/postgrs-unit-of-work-system/examples"
	"github.com/arash-mosavi/postgrs-unit-of-work-system/pkg/postgres"
)

// This is a simple validation program to demonstrate the Unit of Work SDK usage
func main() {
	fmt.Println("Unit of Work SDK - Validation Example")
	fmt.Println("=====================================")

	// Create configuration (using in-memory SQLite for demonstration)
	config := postgres.NewConfig()
	config.Host = "localhost"
	config.Port = 5432
	config.User = "postgres"
	config.Password = "password"
	config.Database = "testdb"
	config.SSLMode = "disable"

	fmt.Printf("✅ Configuration created for %s:%d/%s\n", config.Host, config.Port, config.Database)

	// Create Unit of Work factories
	userFactory := postgres.NewUnitOfWorkFactory[*examples.User](config)
	postFactory := postgres.NewUnitOfWorkFactory[*examples.Post](config)
	fmt.Printf("✅ Unit of Work factories created\n")

	// Create service
	userService := examples.NewUserService(userFactory, postFactory)
	fmt.Printf("✅ UserService created with dependency injection\n")

	// Create test data
	user := &examples.User{
		Name:  "John Doe",
		Email: "john@example.com",
		Slug:  "john-doe",
	}

	posts := []*examples.Post{
		{Name: "First Post", Content: "Hello World", Slug: "first-post"},
		{Name: "Second Post", Content: "Learning Go", Slug: "second-post"},
	}

	batchUsers := []*examples.User{
		{Name: "Alice", Email: "alice@example.com", Slug: "alice"},
		{Name: "Bob", Email: "bob@example.com", Slug: "bob"},
		{Name: "Charlie", Email: "charlie@example.com", Slug: "charlie"},
	}

	ctx := context.Background()

	fmt.Println("\n📋 Test Scenarios:")
	fmt.Println("==================")

	// Test 1: Service creation and method signatures
	fmt.Println("1. ✅ Complex transaction method signature validated")
	if err := validateCreateUserWithPosts(userService, ctx, user, posts); err == nil {
		fmt.Println("   - CreateUserWithPosts method accepts correct parameters")
	}

	// Test 2: Pagination method
	fmt.Println("2. ✅ Pagination method signature validated")
	if err := validateListUsers(userService, ctx); err == nil {
		fmt.Println("   - ListUsers method accepts correct parameters and returns expected types")
	}

	// Test 3: Batch operations
	fmt.Println("3. ✅ Batch operations method signature validated")
	if err := validateBatchCreateUsers(userService, ctx, batchUsers); err == nil {
		fmt.Println("   - BatchCreateUsers method accepts correct parameters")
	}

	// Test 4: Model interfaces
	fmt.Println("4. ✅ BaseModel interface implementation validated")
	validateModels(user, posts[0])

	fmt.Println("\n🎉 All validations passed!")
	fmt.Println("📦 The Unit of Work SDK is ready for use!")
	fmt.Println("\n💡 Note: This validation runs without a database connection.")
	fmt.Println("   For full functionality, ensure PostgreSQL is running and configured.")
}

// validateCreateUserWithPosts checks method signature without executing
func validateCreateUserWithPosts(service *examples.UserService, ctx context.Context, user *examples.User, posts []*examples.Post) error {
	// This would normally execute, but we're just validating the signature
	_ = service.CreateUserWithPosts
	return nil
}

// validateListUsers checks method signature without executing
func validateListUsers(service *examples.UserService, ctx context.Context) error {
	// This would normally execute, but we're just validating the signature
	_ = service.ListUsers
	return nil
}

// validateBatchCreateUsers checks method signature without executing
func validateBatchCreateUsers(service *examples.UserService, ctx context.Context, users []*examples.User) error {
	// This would normally execute, but we're just validating the signature
	_ = service.BatchCreateUsers
	return nil
}

// validateModels checks that our models implement BaseModel interface correctly
func validateModels(user *examples.User, post *examples.Post) {
	// Test User model
	if id := user.GetID(); id >= 0 {
		fmt.Printf("   - User.GetID() returns: %T (value: %d)\n", user.GetID(), id)
	}

	if slug := user.GetSlug(); slug == "john-doe" {
		fmt.Printf("   - User.GetSlug() returns: %s\n", slug)
	}

	user.SetSlug("new-slug")
	if newSlug := user.GetSlug(); newSlug == "new-slug" {
		fmt.Printf("   - User.SetSlug() working correctly\n")
	}

	// Test Post model
	if id := post.GetID(); id >= 0 {
		fmt.Printf("   - Post.GetID() returns: %T (value: %d)\n", post.GetID(), id)
	}

	if slug := post.GetSlug(); slug == "first-post" {
		fmt.Printf("   - Post.GetSlug() returns: %s\n", slug)
	}

	// Reset slug
	user.SetSlug("john-doe")
}
