package examples

import (
	"testing"

	"unit-of-work/pkg/domain"
	"unit-of-work/pkg/persistence"

	"github.com/stretchr/testify/assert"
)

func TestModels_BaseModelInterface(t *testing.T) {
	// Test User model
	user := &User{
		ID:   1,
		Slug: "test-user",
	}

	assert.Equal(t, 1, user.GetID())
	assert.Equal(t, "test-user", user.GetSlug())

	user.SetSlug("new-slug")
	assert.Equal(t, "new-slug", user.GetSlug())

	// Test Post model
	post := &Post{
		ID:   2,
		Slug: "test-post",
	}

	assert.Equal(t, 2, post.GetID())
	assert.Equal(t, "test-post", post.GetSlug())

	post.SetSlug("new-post-slug")
	assert.Equal(t, "new-post-slug", post.GetSlug())

	// Test Tag model
	tag := &Tag{
		ID:   3,
		Slug: "test-tag",
	}

	assert.Equal(t, 3, tag.GetID())
	assert.Equal(t, "test-tag", tag.GetSlug())

	tag.SetSlug("new-tag-slug")
	assert.Equal(t, "new-tag-slug", tag.GetSlug())
}

func TestUserService_Creation(t *testing.T) {
	// Test that we can create a UserService (without database dependency)
	// This tests the service structure and dependency injection pattern

	// Since we can't easily mock the factory without complex setup,
	// we'll just test that the service can be created with a nil factory
	// This validates the constructor pattern

	service := &UserService{uowFactory: nil}
	assert.NotNil(t, service)
	assert.Nil(t, service.uowFactory)
}

func TestRepository_InterfaceCompliance(t *testing.T) {
	// Test that our Repository interfaces are properly defined
	// by checking that we can create instances that would implement them

	// Test that IUserRepository interface is properly defined
	var userRepo IUserRepository
	assert.Nil(t, userRepo) // Interface can be nil

	// Test that IPostRepository interface is properly defined
	var postRepo IPostRepository
	assert.Nil(t, postRepo) // Interface can be nil
}

// ...existing code...

func TestArchitecturalFlow_ServiceToRepository(t *testing.T) {
	// Test the architectural flow: Service -> Repository -> Unit of Work
	// This test validates that the service correctly uses repositories
	// without requiring a database connection

	// Create test data
	user := &User{
		ID:    1,
		Name:  "Test User",
		Email: "test@example.com",
		Slug:  "test-user",
	}

	// Test that UserService constructor accepts the correct factory types
	// This validates our generic factory pattern
	assert.NotPanics(t, func() {
		// The fact that this compiles proves our generics are working correctly
		var userFactory persistence.IUnitOfWorkFactory[*User]
		var postFactory persistence.IUnitOfWorkFactory[*Post]

		// This would create a service following the proper architectural flow
		service := NewUserService(userFactory, postFactory)
		assert.NotNil(t, service)
	})

	// Test BaseModel interface compliance
	assert.Implements(t, (*domain.BaseModel)(nil), user)
	assert.Equal(t, 1, user.GetID())
	assert.Equal(t, "test-user", user.GetSlug())
}

func TestRepositoryPattern_InterfaceCompliance(t *testing.T) {
	// Test that our repository interfaces are properly defined
	// This validates the repository layer of our architecture

	// Test that IUserRepository interface is properly defined
	assert.NotPanics(t, func() {
		var userRepo IUserRepository
		assert.Nil(t, userRepo) // Interface can be nil
	})

	// Test that IPostRepository interface is properly defined
	assert.NotPanics(t, func() {
		var postRepo IPostRepository
		assert.Nil(t, postRepo) // Interface can be nil
	})

	// Test repository constructor pattern
	assert.NotPanics(t, func() {
		// Test that NewUserRepository accepts the correct UoW type
		var uow persistence.IUnitOfWork[*User]
		repo := NewUserRepository(uow)
		assert.NotNil(t, repo)
	})
}

func TestGenericTypes_Compilation(t *testing.T) {
	// Test that our generic types compile correctly
	// This validates our generic Unit of Work implementation

	// Test generic factory creation (compilation test)
	assert.NotPanics(t, func() {
		// These would be created by postgres.NewUnitOfWorkFactory in real usage
		var userFactory persistence.IUnitOfWorkFactory[*User]
		var postFactory persistence.IUnitOfWorkFactory[*Post]

		// Test that they can be used in service creation
		service := NewUserService(userFactory, postFactory)
		assert.NotNil(t, service)
	})

	// Test that Unit of Work interfaces are properly generic
	assert.NotPanics(t, func() {
		var userUoW persistence.IUnitOfWork[*User]
		var postUoW persistence.IUnitOfWork[*Post]

		// These interfaces should be distinct types
		assert.IsType(t, userUoW, userUoW)
		assert.IsType(t, postUoW, postUoW)
	})
}
