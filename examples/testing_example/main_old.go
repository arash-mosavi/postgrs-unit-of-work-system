// Testing Example: How to write tests using the PostgreSQL Unit of Work SDK
//
// This example demonstrates:
// 1. Setting up test database
// 2. Writing unit tests for repositories
// 3. Testing transaction scenarios
// 4. Mocking and test isolation
//
// Run these tests:
//   go test examples/testing_example/

package main

import (
	"context"
	"testing"

	"github.com/arash-mosavi/postgrs-unit-of-work-system/pkg/persistence"
	"github.com/arash-mosavi/postgrs-unit-of-work-system/pkg/postgres"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// Test models
type TestUser struct {
	ID    uint   `gorm:"primarykey"`
	Name  string `gorm:"not null"`
	Email string `gorm:"uniqueIndex;not null"`
}

type TestUserRepository struct {
	*postgres.BaseRepository[*TestUser]
}

func NewTestUserRepository(uow persistence.IUnitOfWork[*TestUser]) *TestUserRepository {
	baseRepo := postgres.NewBaseRepository[*TestUser](uow)
	return &TestUserRepository{BaseRepository: baseRepo}
}

func (r *TestUserRepository) FindByEmail(ctx context.Context, email string) (*TestUser, error) {
	var user TestUser
	err := r.GetDB().WithContext(ctx).Where("email = ?", email).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// Test helper functions
func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err, "Failed to open test database")

	err = db.AutoMigrate(&TestUser{})
	require.NoError(t, err, "Failed to migrate test schema")

	return db
}

func createTestUser() *TestUser {
	return &TestUser{
		Name:  "Test User",
		Email: "test@example.com",
	}
}

// Basic CRUD tests
func TestUserRepository_CRUD(t *testing.T) {
	db := setupTestDB(t)
	factory := postgres.NewUnitOfWorkFactory[*TestUser](db)
	ctx := context.Background()

	t.Run("Create_User", func(t *testing.T) {
		uow := factory.CreateWithContext(ctx)
		repo := NewTestUserRepository(uow)

		user := createTestUser()
		err := repo.Create(ctx, user)

		assert.NoError(t, err)
		assert.NotZero(t, user.ID)
	})

	t.Run("Get_User_By_ID", func(t *testing.T) {
		uow := factory.CreateWithContext(ctx)
		repo := NewTestUserRepository(uow)

		// Create user first
		user := createTestUser()
		err := repo.Create(ctx, user)
		require.NoError(t, err)

		// Get user by ID
		foundUser, err := repo.GetByID(ctx, user.ID)

		assert.NoError(t, err)
		assert.Equal(t, user.Name, foundUser.Name)
		assert.Equal(t, user.Email, foundUser.Email)
	})

	t.Run("Update_User", func(t *testing.T) {
		uow := factory.CreateWithContext(ctx)
		repo := NewTestUserRepository(uow)

		// Create user first
		user := createTestUser()
		err := repo.Create(ctx, user)
		require.NoError(t, err)

		// Update user
		user.Name = "Updated Name"
		err = repo.Update(ctx, user)
		assert.NoError(t, err)

		// Verify update
		updatedUser, err := repo.GetByID(ctx, user.ID)
		require.NoError(t, err)
		assert.Equal(t, "Updated Name", updatedUser.Name)
	})

	t.Run("Delete_User", func(t *testing.T) {
		uow := factory.CreateWithContext(ctx)
		repo := NewTestUserRepository(uow)

		// Create user first
		user := createTestUser()
		err := repo.Create(ctx, user)
		require.NoError(t, err)

		// Delete user
		err = repo.Delete(ctx, user.ID)
		assert.NoError(t, err)

		// Verify deletion
		_, err = repo.GetByID(ctx, user.ID)
		assert.Error(t, err)
	})

	t.Run("Find_By_Email", func(t *testing.T) {
		uow := factory.CreateWithContext(ctx)
		repo := NewTestUserRepository(uow)

		// Create user first
		user := createTestUser()
		user.Email = "unique@example.com"
		err := repo.Create(ctx, user)
		require.NoError(t, err)

		// Find by email
		foundUser, err := repo.FindByEmail(ctx, "unique@example.com")

		assert.NoError(t, err)
		assert.Equal(t, user.ID, foundUser.ID)
		assert.Equal(t, user.Name, foundUser.Name)
	})
}

// Transaction tests
func TestUserRepository_Transactions(t *testing.T) {
	db := setupTestDB(t)
	factory := postgres.NewUnitOfWorkFactory[*TestUser](db)
	ctx := context.Background()

	t.Run("Successful_Transaction", func(t *testing.T) {
		uow := factory.CreateWithContext(ctx)
		repo := NewTestUserRepository(uow)

		// Begin transaction
		err := uow.BeginTransaction(ctx)
		require.NoError(t, err)

		// Create user within transaction
		user := createTestUser()
		user.Email = "transaction@example.com"
		err = repo.Create(ctx, user)
		require.NoError(t, err)

		// Commit transaction
		err = uow.CommitTransaction(ctx)
		require.NoError(t, err)

		// Verify user exists after commit
		foundUser, err := repo.FindByEmail(ctx, "transaction@example.com")
		assert.NoError(t, err)
		assert.Equal(t, user.ID, foundUser.ID)
	})

	t.Run("Rollback_Transaction", func(t *testing.T) {
		uow := factory.CreateWithContext(ctx)
		repo := NewTestUserRepository(uow)

		// Begin transaction
		err := uow.BeginTransaction(ctx)
		require.NoError(t, err)

		// Create user within transaction
		user := createTestUser()
		user.Email = "rollback@example.com"
		err = repo.Create(ctx, user)
		require.NoError(t, err)

		// Rollback transaction
		err = uow.RollbackTransaction(ctx)
		require.NoError(t, err)

		// Verify user doesn't exist after rollback
		_, err = repo.FindByEmail(ctx, "rollback@example.com")
		assert.Error(t, err) // Should not find the user
	})

	t.Run("Multiple_Operations_In_Transaction", func(t *testing.T) {
		uow := factory.CreateWithContext(ctx)
		repo := NewTestUserRepository(uow)

		// Begin transaction
		err := uow.BeginTransaction(ctx)
		require.NoError(t, err)

		// Create multiple users
		user1 := &TestUser{Name: "User 1", Email: "user1@example.com"}
		user2 := &TestUser{Name: "User 2", Email: "user2@example.com"}

		err = repo.Create(ctx, user1)
		require.NoError(t, err)

		err = repo.Create(ctx, user2)
		require.NoError(t, err)

		// Update first user
		user1.Name = "Updated User 1"
		err = repo.Update(ctx, user1)
		require.NoError(t, err)

		// Commit all operations
		err = uow.CommitTransaction(ctx)
		require.NoError(t, err)

		// Verify all operations were successful
		foundUser1, err := repo.FindByEmail(ctx, "user1@example.com")
		require.NoError(t, err)
		assert.Equal(t, "Updated User 1", foundUser1.Name)

		foundUser2, err := repo.FindByEmail(ctx, "user2@example.com")
		require.NoError(t, err)
		assert.Equal(t, "User 2", foundUser2.Name)
	})
}

// Performance and edge case tests
func TestUserRepository_EdgeCases(t *testing.T) {
	db := setupTestDB(t)
	factory := postgres.NewUnitOfWorkFactory[*TestUser](db)
	ctx := context.Background()

	t.Run("Get_Nonexistent_User", func(t *testing.T) {
		uow := factory.CreateWithContext(ctx)
		repo := NewTestUserRepository(uow)

		_, err := repo.GetByID(ctx, 99999)
		assert.Error(t, err)
	})

	t.Run("Delete_Nonexistent_User", func(t *testing.T) {
		uow := factory.CreateWithContext(ctx)
		repo := NewTestUserRepository(uow)

		err := repo.Delete(ctx, 99999)
		// Delete should not error for nonexistent records in this implementation
		assert.NoError(t, err)
	})

	t.Run("Unique_Constraint_Violation", func(t *testing.T) {
		uow := factory.CreateWithContext(ctx)
		repo := NewTestUserRepository(uow)

		// Create first user
		user1 := &TestUser{Name: "User 1", Email: "duplicate@example.com"}
		err := repo.Create(ctx, user1)
		require.NoError(t, err)

		// Try to create second user with same email
		user2 := &TestUser{Name: "User 2", Email: "duplicate@example.com"}
		err = repo.Create(ctx, user2)
		assert.Error(t, err) // Should fail due to unique constraint
	})

	t.Run("Concurrent_Access", func(t *testing.T) {
		// This test demonstrates that each UoW instance is independent
		uow1 := factory.CreateWithContext(ctx)
		uow2 := factory.CreateWithContext(ctx)

		repo1 := NewTestUserRepository(uow1)
		repo2 := NewTestUserRepository(uow2)

		// Create users with different repositories
		user1 := &TestUser{Name: "Concurrent User 1", Email: "concurrent1@example.com"}
		user2 := &TestUser{Name: "Concurrent User 2", Email: "concurrent2@example.com"}

		err1 := repo1.Create(ctx, user1)
		err2 := repo2.Create(ctx, user2)

		assert.NoError(t, err1)
		assert.NoError(t, err2)

		// Both should be able to read each other's data
		foundUser1, err := repo2.FindByEmail(ctx, "concurrent1@example.com")
		assert.NoError(t, err)
		assert.Equal(t, user1.Name, foundUser1.Name)
	})
}

// Benchmark tests
func BenchmarkUserRepository_Create(b *testing.B) {
	require := require.New(b)
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(err)
	err = db.AutoMigrate(&TestUser{})
	require.NoError(err)
	
	factory := postgres.NewUnitOfWorkFactory[*TestUser](db)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		uow := factory.CreateWithContext(ctx)
		repo := NewTestUserRepository(uow)

		user := &TestUser{
			Name:  "Benchmark User",
			Email: fmt.Sprintf("bench%d@example.com", i),
		}

		_ = repo.Create(ctx, user)
	}
}

func BenchmarkUserRepository_GetByID(b *testing.B) {
	require := require.New(b)
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(err)
	err = db.AutoMigrate(&TestUser{})
	require.NoError(err)
	
	factory := postgres.NewUnitOfWorkFactory[*TestUser](db)
	ctx := context.Background()

	// Create a user for benchmarking
	uow := factory.CreateWithContext(ctx)
	repo := NewTestUserRepository(uow)
	user := createTestUser()
	_ = repo.Create(ctx, user)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = repo.GetByID(ctx, user.ID)
	}
}

// Example test with table-driven tests
func TestUserRepository_TableDriven(t *testing.T) {
	db := setupTestDB(t)
	factory := postgres.NewUnitOfWorkFactory[*TestUser](db)
	ctx := context.Background()

	tests := []struct {
		name    string
		user    *TestUser
		wantErr bool
	}{
		{
			name:    "Valid user",
			user:    &TestUser{Name: "Valid User", Email: "valid@example.com"},
			wantErr: false,
		},
		{
			name:    "Empty name",
			user:    &TestUser{Name: "", Email: "empty@example.com"},
			wantErr: false, // GORM allows empty strings by default
		},
		{
			name:    "Empty email",
			user:    &TestUser{Name: "No Email", Email: ""},
			wantErr: true, // Email is required and unique
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uow := factory.CreateWithContext(ctx)
			repo := NewTestUserRepository(uow)

			err := repo.Create(ctx, tt.user)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotZero(t, tt.user.ID)
			}
		})
	}
}

// Main function for running as a standalone program
func main() {
	// This is just for demonstration - normally you'd run: go test
	testing.Main(
		func(string, string) (bool, error) { return true, nil },
		[]testing.InternalTest{
			{"TestUserRepository_CRUD", TestUserRepository_CRUD},
			{"TestUserRepository_Transactions", TestUserRepository_Transactions},
			{"TestUserRepository_EdgeCases", TestUserRepository_EdgeCases},
		},
		[]testing.InternalBenchmark{
			{"BenchmarkUserRepository_Create", BenchmarkUserRepository_Create},
			{"BenchmarkUserRepository_GetByID", BenchmarkUserRepository_GetByID},
		},
		[]testing.InternalExample{},
	)
}
