// Basic Example: Simple CRUD operations using the PostgreSQL Unit of Work SDK
//
// This example demonstrates:
// 1. Setting up database connection
// 2. Creating services with Unit of Work pattern
// 3. Basic CRUD operations (Create, Read, Update, Delete)
// 4. Transaction handling and error management
//
// Run this example:
//   go run examples/basic_example/main.go

package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// User represents a simple user model that implements BaseModel interface
type User struct {
	ID        int            `gorm:"primarykey"`
	Name      string         `gorm:"not null"`
	Email     string         `gorm:"uniqueIndex;not null"`
	Slug      string         `gorm:"uniqueIndex;not null"`
	CreatedAt time.Time      `gorm:"autoCreateTime"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime"`
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

// Implement BaseModel interface methods
func (u *User) GetID() int                    { return u.ID }
func (u *User) GetSlug() string               { return u.Slug }
func (u *User) SetSlug(slug string)           { u.Slug = slug }
func (u *User) GetCreatedAt() time.Time       { return u.CreatedAt }
func (u *User) GetUpdatedAt() time.Time       { return u.UpdatedAt }
func (u *User) GetArchivedAt() gorm.DeletedAt { return u.DeletedAt }
func (u *User) GetName() string               { return u.Name }

// UserService demonstrates Unit of Work pattern with direct GORM operations
// This approach maintains transaction boundaries and proper error handling
type UserService struct {
	db *gorm.DB
}

func NewUserService(db *gorm.DB) *UserService {
	return &UserService{db: db}
}

// Create demonstrates creating a user within a transaction boundary
func (s *UserService) Create(ctx context.Context, user *User) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		return tx.WithContext(ctx).Create(user).Error
	})
}

// GetByID demonstrates reading a user by ID
func (s *UserService) GetByID(ctx context.Context, id int) (*User, error) {
	var user User
	err := s.db.WithContext(ctx).First(&user, id).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// Update demonstrates updating a user within a transaction boundary
func (s *UserService) Update(ctx context.Context, user *User) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		return tx.WithContext(ctx).Save(user).Error
	})
}

// Delete demonstrates soft deletion within a transaction boundary
func (s *UserService) Delete(ctx context.Context, id int) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		return tx.WithContext(ctx).Delete(&User{}, id).Error
	})
}

// GetAll demonstrates retrieving all non-deleted users
func (s *UserService) GetAll(ctx context.Context) ([]*User, error) {
	var users []*User
	err := s.db.WithContext(ctx).Find(&users).Error
	return users, err
}

// FindByEmail demonstrates finding a user by email with proper error handling
func (s *UserService) FindByEmail(ctx context.Context, email string) (*User, error) {
	var user User
	err := s.db.WithContext(ctx).Where("email = ?", email).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func main() {
	// Setup database connection with SQLite for easy demonstration
	ctx := context.Background()

	// Create SQLite database for demo
	db, err := setupDatabase()
	if err != nil {
		log.Fatal("Failed to setup database:", err)
	}

	// Create UserService - demonstrates Unit of Work pattern with transaction boundaries
	userService := NewUserService(db)

	// Run basic CRUD examples
	fmt.Println("=== Basic CRUD Example ===")
	fmt.Println("Note: This example uses SQLite for demonstration.")
	fmt.Println("In production, use PostgreSQL with proper configuration.")

	// 1. Create a user
	user := &User{
		Name:  "John Doe",
		Email: "john.doe@example.com",
		Slug:  "john-doe",
	}

	fmt.Println("\n1. Creating user...")
	if err := userService.Create(ctx, user); err != nil {
		log.Fatal("Failed to create user:", err)
	}
	fmt.Printf("   Created user with ID: %d\n", user.ID)

	// 2. Read the user by ID
	fmt.Println("2. Reading user by ID...")
	foundUser, err := userService.GetByID(ctx, user.ID)
	if err != nil {
		log.Fatal("Failed to get user:", err)
	}
	fmt.Printf("   Found user: %+v\n", foundUser)

	// 3. Read user by email (custom method)
	fmt.Println("3. Reading user by email...")
	userByEmail, err := userService.FindByEmail(ctx, "john.doe@example.com")
	if err != nil {
		log.Fatal("Failed to get user by email:", err)
	}
	fmt.Printf("   Found user by email: %+v\n", userByEmail)

	// 4. Update the user
	fmt.Println("4. Updating user...")
	foundUser.Name = "John Smith"
	if err := userService.Update(ctx, foundUser); err != nil {
		log.Fatal("Failed to update user:", err)
	}
	fmt.Printf("   Updated user: %+v\n", foundUser)

	// 5. List all users
	fmt.Println("5. Listing all users...")
	users, err := userService.GetAll(ctx)
	if err != nil {
		log.Fatal("Failed to get all users:", err)
	}
	fmt.Printf("   Total users: %d\n", len(users))
	for _, u := range users {
		fmt.Printf("   - %+v\n", u)
	}

	// 6. Delete the user
	fmt.Println("6. Deleting user...")
	if err := userService.Delete(ctx, foundUser.ID); err != nil {
		log.Fatal("Failed to delete user:", err)
	}
	fmt.Println("   User deleted successfully")

	// 7. Verify deletion (should return not found error)
	fmt.Println("7. Verifying deletion...")
	_, err = userService.GetByID(ctx, foundUser.ID)
	if err != nil {
		fmt.Println("   User successfully deleted (not found)")
	} else {
		fmt.Println("   Warning: User still exists")
	}

	fmt.Println("\n=== Basic CRUD Example Completed ===")
	fmt.Println("\nUnit of Work Pattern Benefits Demonstrated:")
	fmt.Println("- Transaction boundaries for each operation")
	fmt.Println("- Automatic rollback on errors")
	fmt.Println("- Consistent error handling")
	fmt.Println("- Clear separation of concerns")
	fmt.Println("\nTo use with PostgreSQL:")
	fmt.Println("1. Setup PostgreSQL database")
	fmt.Println("2. Update connection string")
	fmt.Println("3. Replace SQLite driver with PostgreSQL driver")
}

// setupDatabase creates a database connection for demonstration
// In this example, we use SQLite for simplicity
func setupDatabase() (*gorm.DB, error) {
	// For demonstration purposes, we'll use SQLite
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Auto-migrate the schema
	if err := db.AutoMigrate(&User{}); err != nil {
		return nil, fmt.Errorf("failed to migrate schema: %w", err)
	}

	return db, nil
}
