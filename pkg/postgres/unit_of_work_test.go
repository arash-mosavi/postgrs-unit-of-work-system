package postgres

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/arash-mosavi/postgrs-unit-of-work-system/pkg/domain"
	"github.com/arash-mosavi/postgrs-unit-of-work-system/pkg/identifier"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// TestUser implements BaseModel for testing
type TestUser struct {
	ID        int            `gorm:"primaryKey;autoIncrement" json:"id"`
	Slug      string         `gorm:"uniqueIndex;size:100;not null" json:"slug"`
	Name      string         `gorm:"size:255;not null" json:"name"`
	Email     string         `gorm:"uniqueIndex;size:255;not null" json:"email"`
	Active    bool           `gorm:"default:true" json:"active"`
	CreatedAt time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

// Implement BaseModel interface
func (u *TestUser) GetID() int                    { return u.ID }
func (u *TestUser) GetSlug() string               { return u.Slug }
func (u *TestUser) SetSlug(slug string)           { u.Slug = slug }
func (u *TestUser) GetCreatedAt() time.Time       { return u.CreatedAt }
func (u *TestUser) GetUpdatedAt() time.Time       { return u.UpdatedAt }
func (u *TestUser) GetArchivedAt() gorm.DeletedAt { return u.DeletedAt }
func (u *TestUser) GetName() string               { return u.Name }

func (TestUser) TableName() string { return "test_users" }

// setupTestDB creates an in-memory SQLite database for testing
func setupTestDB(t *testing.T) *UnitOfWork[*TestUser] {
	// Use SQLite in-memory database for testing
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	// Auto-migrate the test user table
	err = db.AutoMigrate(&TestUser{})
	require.NoError(t, err)

	return &UnitOfWork[*TestUser]{
		db:           db,
		ctx:          context.Background(),
		repositories: make(map[string]interface{}),
	}
}

func TestUnitOfWork_BeginTransaction(t *testing.T) {
	uow := setupTestDB(t)
	ctx := context.Background()

	// Test begin transaction
	err := uow.BeginTransaction(ctx)
	assert.NoError(t, err)

	// Test double begin should fail
	err = uow.BeginTransaction(ctx)
	assert.Error(t, err)

	// Test rollback
	uow.RollbackTransaction(ctx)
}

func TestUnitOfWork_Insert(t *testing.T) {
	uow := setupTestDB(t)
	ctx := context.Background()

	user := &TestUser{
		Name:  "John Doe",
		Email: "john@example.com",
		Slug:  "john-doe",
	}

	// Test Insert
	insertedUser, err := uow.Insert(ctx, user)
	assert.NoError(t, err)
	assert.NotNil(t, insertedUser)
	assert.NotZero(t, insertedUser.GetID())
	assert.Equal(t, "John Doe", insertedUser.GetName())
}

func TestUnitOfWork_FindOneById(t *testing.T) {
	uow := setupTestDB(t)
	ctx := context.Background()

	// First create a user
	user := &TestUser{
		Name:  "Jane Doe",
		Email: "jane@example.com",
		Slug:  "jane-doe",
	}

	insertedUser, err := uow.Insert(ctx, user)
	require.NoError(t, err)

	// Test FindOneById
	foundUser, err := uow.FindOneById(ctx, insertedUser.GetID())
	assert.NoError(t, err)
	assert.NotNil(t, foundUser)
	assert.Equal(t, insertedUser.GetID(), foundUser.GetID())
	assert.Equal(t, "Jane Doe", foundUser.GetName())
}

func TestUnitOfWork_FindOneByIdentifier(t *testing.T) {
	uow := setupTestDB(t)
	ctx := context.Background()

	// First create a user
	user := &TestUser{
		Name:  "Bob Smith",
		Email: "bob@example.com",
		Slug:  "bob-smith",
	}

	_, err := uow.Insert(ctx, user)
	require.NoError(t, err)

	// Test FindOneByIdentifier with email
	emailIdentifier := identifier.NewIdentifier().Equal("email", "bob@example.com")
	foundUser, err := uow.FindOneByIdentifier(ctx, emailIdentifier)
	assert.NoError(t, err)
	assert.NotNil(t, foundUser)
	assert.Equal(t, "Bob Smith", foundUser.GetName())
}

func TestUnitOfWork_FindAll(t *testing.T) {
	uow := setupTestDB(t)
	ctx := context.Background()

	// Create multiple users
	users := []*TestUser{
		{Name: "User 1", Email: "user1@example.com", Slug: "user-1"},
		{Name: "User 2", Email: "user2@example.com", Slug: "user-2"},
		{Name: "User 3", Email: "user3@example.com", Slug: "user-3"},
	}

	for _, user := range users {
		_, err := uow.Insert(ctx, user)
		require.NoError(t, err)
	}

	// Test FindAll
	allUsers, err := uow.FindAll(ctx)
	assert.NoError(t, err)
	assert.Len(t, allUsers, 3)
}

func TestUnitOfWork_FindAllWithPagination(t *testing.T) {
	uow := setupTestDB(t)
	ctx := context.Background()

	// Create multiple users
	for i := 1; i <= 5; i++ {
		user := &TestUser{
			Name:  fmt.Sprintf("User %d", i),
			Email: fmt.Sprintf("user%d@example.com", i),
			Slug:  fmt.Sprintf("user-%d", i),
		}
		_, err := uow.Insert(ctx, user)
		require.NoError(t, err)
	}

	// Test pagination
	params := domain.QueryParams[*TestUser]{
		Limit:  2,
		Offset: 1,
		Sort: domain.SortMap{
			"id": domain.SortAsc,
		},
	}

	users, total, err := uow.FindAllWithPagination(ctx, params)
	assert.NoError(t, err)
	assert.Len(t, users, 2)
	assert.Equal(t, uint(5), total)
}

func TestUnitOfWork_Update(t *testing.T) {
	uow := setupTestDB(t)
	ctx := context.Background()

	// Create a user
	user := &TestUser{
		Name:  "Original Name",
		Email: "original@example.com",
		Slug:  "original",
	}

	insertedUser, err := uow.Insert(ctx, user)
	require.NoError(t, err)

	// Update the user
	updatedData := &TestUser{
		Name:  "Updated Name",
		Email: "updated@example.com",
		Slug:  "updated",
	}

	userIdentifier := identifier.NewIdentifier().Equal("id", insertedUser.GetID())
	updatedUser, err := uow.Update(ctx, userIdentifier, updatedData)
	assert.NoError(t, err)
	assert.NotNil(t, updatedUser)
	assert.Equal(t, "Updated Name", updatedUser.GetName())
}

func TestUnitOfWork_SoftDelete(t *testing.T) {
	uow := setupTestDB(t)
	ctx := context.Background()

	// Create a user
	user := &TestUser{
		Name:  "Delete Me",
		Email: "deleteme@example.com",
		Slug:  "delete-me",
	}

	insertedUser, err := uow.Insert(ctx, user)
	require.NoError(t, err)

	// Soft delete the user
	userIdentifier := identifier.NewIdentifier().Equal("id", insertedUser.GetID())
	deletedUser, err := uow.SoftDelete(ctx, userIdentifier)
	assert.NoError(t, err)
	assert.NotNil(t, deletedUser)

	// Verify user is soft deleted (should not appear in normal queries)
	_, err = uow.FindOneById(ctx, insertedUser.GetID())
	assert.Error(t, err) // Should not find the soft-deleted user
}

func TestUnitOfWork_BulkInsert(t *testing.T) {
	uow := setupTestDB(t)
	ctx := context.Background()

	users := []*TestUser{
		{Name: "Bulk 1", Email: "bulk1@example.com", Slug: "bulk-1"},
		{Name: "Bulk 2", Email: "bulk2@example.com", Slug: "bulk-2"},
		{Name: "Bulk 3", Email: "bulk3@example.com", Slug: "bulk-3"},
	}

	// Test bulk insert
	insertedUsers, err := uow.BulkInsert(ctx, users)
	assert.NoError(t, err)
	assert.Len(t, insertedUsers, 3)

	for _, user := range insertedUsers {
		assert.NotZero(t, user.GetID())
	}
}

func TestUnitOfWork_TransactionRollback(t *testing.T) {
	uow := setupTestDB(t)
	ctx := context.Background()

	// Begin transaction
	err := uow.BeginTransaction(ctx)
	require.NoError(t, err)

	// Insert a user in transaction
	user := &TestUser{
		Name:  "Transaction Test",
		Email: "transaction@example.com",
		Slug:  "transaction-test",
	}

	_, err = uow.Insert(ctx, user)
	require.NoError(t, err)

	// Rollback transaction
	uow.RollbackTransaction(ctx)

	// Verify user was not persisted
	users, err := uow.FindAll(ctx)
	assert.NoError(t, err)
	assert.Len(t, users, 0)
}

func TestUnitOfWork_TransactionCommit(t *testing.T) {
	uow := setupTestDB(t)
	ctx := context.Background()

	// Begin transaction
	err := uow.BeginTransaction(ctx)
	require.NoError(t, err)

	// Insert a user in transaction
	user := &TestUser{
		Name:  "Commit Test",
		Email: "commit@example.com",
		Slug:  "commit-test",
	}

	_, err = uow.Insert(ctx, user)
	require.NoError(t, err)

	// Commit transaction
	err = uow.CommitTransaction(ctx)
	assert.NoError(t, err)

	// Verify user was persisted
	users, err := uow.FindAll(ctx)
	assert.NoError(t, err)
	assert.Len(t, users, 1)
	assert.Equal(t, "Commit Test", users[0].GetName())
}
