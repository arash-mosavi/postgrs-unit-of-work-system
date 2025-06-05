// Testing Example: Unit testing with PostgreSQL Unit of Work SDK
//
// This example demonstrates:
// 1. Setting up test database
// 2. Testing CRUD operations with Unit of Work pattern
// 3. Testing transaction rollback scenarios
// 4. Performance testing
// 5. Error handling patterns
//
// Run this example:
//   go run examples/testing_example/main.go

package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// TestUser represents a test user model that implements BaseModel interface
type TestUser struct {
	ID        int            `gorm:"primarykey"`
	Name      string         `gorm:"not null"`
	Email     string         `gorm:"uniqueIndex;not null"`
	Slug      string         `gorm:"uniqueIndex;not null"`
	CreatedAt time.Time      `gorm:"autoCreateTime"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime"`
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

// Implement BaseModel interface methods
func (u *TestUser) GetID() int                    { return u.ID }
func (u *TestUser) GetSlug() string               { return u.Slug }
func (u *TestUser) SetSlug(slug string)           { u.Slug = slug }
func (u *TestUser) GetCreatedAt() time.Time       { return u.CreatedAt }
func (u *TestUser) GetUpdatedAt() time.Time       { return u.UpdatedAt }
func (u *TestUser) GetArchivedAt() gorm.DeletedAt { return u.DeletedAt }
func (u *TestUser) GetName() string               { return u.Name }

// TestUserService demonstrates Unit of Work pattern for testing
type TestUserService struct {
	db *gorm.DB
}

func NewTestUserService(db *gorm.DB) *TestUserService {
	return &TestUserService{db: db}
}

// CRUD operations with transaction boundaries
func (s *TestUserService) Create(ctx context.Context, user *TestUser) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		return tx.WithContext(ctx).Create(user).Error
	})
}

func (s *TestUserService) GetByID(ctx context.Context, id int) (*TestUser, error) {
	var user TestUser
	err := s.db.WithContext(ctx).First(&user, id).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (s *TestUserService) Update(ctx context.Context, user *TestUser) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		return tx.WithContext(ctx).Save(user).Error
	})
}

func (s *TestUserService) Delete(ctx context.Context, id int) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		return tx.WithContext(ctx).Delete(&TestUser{}, id).Error
	})
}

func (s *TestUserService) FindByEmail(ctx context.Context, email string) (*TestUser, error) {
	var user TestUser
	err := s.db.WithContext(ctx).Where("email = ?", email).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// CreateMultipleUsers demonstrates complex transaction with rollback capability
func (s *TestUserService) CreateMultipleUsers(ctx context.Context, users []*TestUser) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		for _, user := range users {
			if err := tx.WithContext(ctx).Create(user).Error; err != nil {
				// Transaction will automatically rollback on error
				return fmt.Errorf("failed to create user %s: %w", user.Email, err)
			}
		}
		return nil
	})
}

func main() {
	ctx := context.Background()

	fmt.Println("=== Testing Example: Unit of Work Pattern ===")
	fmt.Println("This example demonstrates testing patterns with transaction boundaries")

	// Setup test database
	db, err := setupTestDatabase()
	if err != nil {
		log.Fatal("Failed to setup test database:", err)
	}

	service := NewTestUserService(db)

	// Run various test scenarios
	testBasicCRUD(ctx, service)
	testTransactionRollback(ctx, service)
	testErrorHandling(ctx, service)
	testConcurrentOperations(ctx, service)
	testPerformance(ctx, service)

	fmt.Println("\n=== Testing Example Completed ===")
	fmt.Println("\nTesting Best Practices Demonstrated:")
	fmt.Println("- Transaction boundaries for data integrity")
	fmt.Println("- Proper error handling and rollback")
	fmt.Println("- Isolation of test scenarios")
	fmt.Println("- Performance testing patterns")
	fmt.Println("- Edge case handling")
}

func testBasicCRUD(ctx context.Context, service *TestUserService) {
	fmt.Println("\n--- Test: Basic CRUD Operations ---")

	// Test Create
	user := &TestUser{
		Name:  "Test User",
		Email: "test@example.com",
		Slug:  "test-user",
	}

	if err := service.Create(ctx, user); err != nil {
		fmt.Printf("❌ Create failed: %v\n", err)
		return
	}
	fmt.Printf("✅ Created user with ID: %d\n", user.ID)

	// Test Read
	foundUser, err := service.GetByID(ctx, user.ID)
	if err != nil {
		fmt.Printf("❌ Read failed: %v\n", err)
		return
	}
	fmt.Printf("✅ Read user: %s (%s)\n", foundUser.Name, foundUser.Email)

	// Test Update
	foundUser.Name = "Updated Test User"
	if err := service.Update(ctx, foundUser); err != nil {
		fmt.Printf("❌ Update failed: %v\n", err)
		return
	}
	fmt.Printf("✅ Updated user name to: %s\n", foundUser.Name)

	// Test Delete
	if err := service.Delete(ctx, foundUser.ID); err != nil {
		fmt.Printf("❌ Delete failed: %v\n", err)
		return
	}
	fmt.Printf("✅ Deleted user successfully\n")

	// Verify deletion
	_, err = service.GetByID(ctx, foundUser.ID)
	if err != nil {
		fmt.Printf("✅ Deletion verified (user not found)\n")
	} else {
		fmt.Printf("❌ User still exists after deletion\n")
	}
}

func testTransactionRollback(ctx context.Context, service *TestUserService) {
	fmt.Println("\n--- Test: Transaction Rollback ---")

	// Create users that will cause a conflict
	users := []*TestUser{
		{Name: "User 1", Email: "user1@example.com", Slug: "user-1"},
		{Name: "User 2", Email: "user2@example.com", Slug: "user-2"},
		{Name: "User 3", Email: "user1@example.com", Slug: "user-3"}, // Duplicate email - will fail
	}

	// This should fail and rollback all operations
	err := service.CreateMultipleUsers(ctx, users)
	if err != nil {
		fmt.Printf("✅ Transaction rolled back as expected: %v\n", err)

		// Verify no users were created
		if _, err := service.FindByEmail(ctx, "user1@example.com"); err != nil {
			fmt.Printf("✅ Rollback verified - no users exist\n")
		} else {
			fmt.Printf("❌ Rollback failed - users still exist\n")
		}
	} else {
		fmt.Printf("❌ Transaction should have failed\n")
	}
}

func testErrorHandling(ctx context.Context, service *TestUserService) {
	fmt.Println("\n--- Test: Error Handling ---")

	// Test getting non-existent user
	_, err := service.GetByID(ctx, 99999)
	if err != nil {
		fmt.Printf("✅ Handled non-existent user error: %v\n", err)
	} else {
		fmt.Printf("❌ Should have returned error for non-existent user\n")
	}

	// Test duplicate email constraint
	user1 := &TestUser{Name: "User 1", Email: "duplicate@example.com", Slug: "dup-1"}
	user2 := &TestUser{Name: "User 2", Email: "duplicate@example.com", Slug: "dup-2"}

	if err := service.Create(ctx, user1); err != nil {
		fmt.Printf("❌ Failed to create first user: %v\n", err)
		return
	}
	fmt.Printf("✅ Created first user\n")

	if err := service.Create(ctx, user2); err != nil {
		fmt.Printf("✅ Handled duplicate email error: %v\n", err)
	} else {
		fmt.Printf("❌ Should have failed due to duplicate email\n")
	}
}

func testConcurrentOperations(ctx context.Context, service *TestUserService) {
	fmt.Println("\n--- Test: Concurrent Operations ---")

	// Simulate concurrent user creation
	user1 := &TestUser{Name: "Concurrent User 1", Email: "concurrent1@example.com", Slug: "conc-1"}
	user2 := &TestUser{Name: "Concurrent User 2", Email: "concurrent2@example.com", Slug: "conc-2"}

	// Create users concurrently (simulated)
	err1 := service.Create(ctx, user1)
	err2 := service.Create(ctx, user2)

	if err1 == nil && err2 == nil {
		fmt.Printf("✅ Concurrent operations completed successfully\n")
		fmt.Printf("   User 1 ID: %d, User 2 ID: %d\n", user1.ID, user2.ID)
	} else {
		fmt.Printf("❌ Concurrent operations failed: err1=%v, err2=%v\n", err1, err2)
	}
}

func testPerformance(ctx context.Context, service *TestUserService) {
	fmt.Println("\n--- Test: Performance Testing ---")

	startTime := time.Now()
	userCount := 100

	// Create multiple users to test performance
	for i := 0; i < userCount; i++ {
		user := &TestUser{
			Name:  fmt.Sprintf("Perf User %d", i),
			Email: fmt.Sprintf("perf%d@example.com", i),
			Slug:  fmt.Sprintf("perf-user-%d", i),
		}

		if err := service.Create(ctx, user); err != nil {
			fmt.Printf("❌ Performance test failed at user %d: %v\n", i, err)
			return
		}
	}

	duration := time.Since(startTime)
	fmt.Printf("✅ Created %d users in %v\n", userCount, duration)
	fmt.Printf("   Average: %v per user\n", duration/time.Duration(userCount))
}

func setupTestDatabase() (*gorm.DB, error) {
	// Use SQLite in-memory database for testing
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to test database: %w", err)
	}

	// Auto-migrate the schema
	if err := db.AutoMigrate(&TestUser{}); err != nil {
		return nil, fmt.Errorf("failed to migrate test schema: %w", err)
	}

	return db, nil
}
