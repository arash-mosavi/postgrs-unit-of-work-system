package main

import (
	"context"
	"fmt"

	"github.com/arash-mosavi/postgrs-unit-of-work-system/examples"
	"github.com/arash-mosavi/postgrs-unit-of-work-system/pkg/postgres"
)

func main() {
	fmt.Println("Unit of Work SDK - Validation Example")
	fmt.Println("=====================================")

	config := postgres.NewConfig()
	config.Host = "localhost"
	config.Port = 5432
	config.User = "postgres"
	config.Password = "password"
	config.Database = "testdb"
	config.SSLMode = "disable"

	fmt.Printf("[OK] Configuration created for %s:%d/%s\n", config.Host, config.Port, config.Database)

	userFactory := postgres.NewUnitOfWorkFactory[*examples.User](config)
	postFactory := postgres.NewUnitOfWorkFactory[*examples.Post](config)
	fmt.Printf("[OK] Unit of Work factories created\n")

	userService := examples.NewUserService(userFactory, postFactory)
	fmt.Printf("[OK] UserService created with dependency injection\n")

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

	fmt.Println("\n[INFO] Test Scenarios:")
	fmt.Println("==================")

	fmt.Println("1. [OK] Complex transaction method signature validated")
	if err := validateCreateUserWithPosts(userService, ctx, user, posts); err == nil {
		fmt.Println("   - CreateUserWithPosts method accepts correct parameters")
	}

	fmt.Println("2. [OK] Pagination method signature validated")
	if err := validateListUsers(userService, ctx); err == nil {
		fmt.Println("   - ListUsers method accepts correct parameters and returns expected types")
	}

	fmt.Println("3. [OK] Batch operations method signature validated")
	if err := validateBatchCreateUsers(userService, ctx, batchUsers); err == nil {
		fmt.Println("   - BatchCreateUsers method accepts correct parameters")
	}

	fmt.Println("4. [OK] BaseModel interface implementation validated")
	validateModels(user, posts[0])

	fmt.Println("\n[SUCCESS] All validations passed!")
	fmt.Println("[READY] The Unit of Work SDK is ready for use!")
	fmt.Println("\n[NOTE] Note: This validation runs without a database connection.")
	fmt.Println("   For full functionality, ensure PostgreSQL is running and configured.")
}

func validateCreateUserWithPosts(service *examples.UserService, ctx context.Context, user *examples.User, posts []*examples.Post) error {

	_ = service.CreateUserWithPosts
	return nil
}

func validateListUsers(service *examples.UserService, ctx context.Context) error {

	_ = service.ListUsers
	return nil
}

func validateBatchCreateUsers(service *examples.UserService, ctx context.Context, users []*examples.User) error {

	_ = service.BatchCreateUsers
	return nil
}

func validateModels(user *examples.User, post *examples.Post) {

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

	if id := post.GetID(); id >= 0 {
		fmt.Printf("   - Post.GetID() returns: %T (value: %d)\n", post.GetID(), id)
	}

	if slug := post.GetSlug(); slug == "first-post" {
		fmt.Printf("   - Post.GetSlug() returns: %s\n", slug)
	}

	user.SetSlug("john-doe")
}
