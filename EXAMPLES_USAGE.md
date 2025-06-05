# PostgreSQL Unit of Work SDK - Usage Examples

This guide provides comprehensive, runnable examples demonstrating how to use the PostgreSQL Unit of Work SDK effectively.

## 📋 Table of Contents

1. [Quick Start](#quick-start)
2. [Prerequisites](#prerequisites)
3. [Running the Examples](#running-the-examples)
4. [Example 1: Basic CRUD Operations](#example-1-basic-crud-operations)
5. [Example 2: Advanced Transaction Handling](#example-2-advanced-transaction-handling)
6. [Example 3: Testing Patterns](#example-3-testing-patterns)
7. [Production vs Demo Setup](#production-vs-demo-setup)
8. [Best Practices](#best-practices)
9. [Troubleshooting](#troubleshooting)

## 🚀 Quick Start

### Prerequisites

- Go 1.21+
- PostgreSQL 12+ (for production) or SQLite (for demo)
- Git

### Installation

```bash
# Clone the repository
git clone https://github.com/arash-mosavi/postgrs-unit-of-work-system.git
cd postgrs-unit-of-work-system

# Install dependencies
go mod download
```

## 🏃 Running the Examples

All examples are designed to work with both SQLite (for easy demonstration) and PostgreSQL (for production).

### Option 1: Run with SQLite (Demo Mode)

Each example includes SQLite setup for immediate execution without external dependencies:

```bash
# Basic CRUD operations
go run examples/basic_example/main.go

# Advanced transaction handling
go run examples/transaction_example/main.go

# Run tests (includes benchmarks)
go test examples/testing_example/ -v -bench=.
```

### Option 2: Run with PostgreSQL (Production Mode)

1. **Setup PostgreSQL Database:**

```bash
# Using Docker
docker run --name postgres-uow \
  -e POSTGRES_PASSWORD=password \
  -e POSTGRES_DB=testdb \
  -p 5432:5432 \
  -d postgres:15

# Or using docker-compose (included in repository)
docker-compose up -d postgres
```

2. **Set Environment Variables:**

```bash
export DB_HOST=localhost
export DB_USER=postgres
export DB_PASSWORD=password
export DB_NAME=testdb
export DB_PORT=5432
export DB_SSLMODE=disable
```

3. **Run Examples:**

```bash
# The examples will automatically detect environment variables
# and use PostgreSQL instead of SQLite
go run examples/basic_example/main.go
go run examples/transaction_example/main.go
```

## 📖 Example 1: Basic CRUD Operations

**File:** `examples/basic_example/main.go`

This example demonstrates:
- Setting up database connection
- Creating repositories with Unit of Work pattern
- Basic CRUD operations (Create, Read, Update, Delete)
- Error handling

### Key Features Demonstrated:

```go
// 1. Model Definition (implements BaseModel interface)
type User struct {
    ID        int            `gorm:"primarykey"`
    Name      string         `gorm:"not null"`
    Email     string         `gorm:"uniqueIndex;not null"`
    Slug      string         `gorm:"uniqueIndex;not null"`
    CreatedAt time.Time      `gorm:"autoCreateTime"`
    UpdatedAt time.Time      `gorm:"autoUpdateTime"`
    DeletedAt gorm.DeletedAt `gorm:"index"`
}

// 2. Repository Pattern with Unit of Work
type UserRepository struct {
    uow persistence.IUnitOfWork[*User]
}

// 3. CRUD Operations
func (r *UserRepository) Create(ctx context.Context, user *User) error {
    _, err := r.uow.Insert(ctx, user)
    return err
}
```

### Running the Basic Example:

```bash
go run examples/basic_example/main.go
```

**Expected Output:**
```
=== Basic CRUD Example ===
Note: This example uses SQLite for demonstration.
In production, use PostgreSQL with proper configuration.

1. Creating user...
   Created user with ID: 1
2. Reading user by ID...
   Found user: &{ID:1 Name:John Doe Email:john.doe@example.com ...}
3. Reading user by email...
   Found user by email: &{ID:1 Name:John Doe Email:john.doe@example.com ...}
4. Updating user...
   Updated user: &{ID:1 Name:John Smith Email:john.doe@example.com ...}
5. Listing all users...
   Total users: 1
6. Deleting user...
   User deleted successfully
7. Verifying deletion...
   User successfully deleted (not found)

=== Basic CRUD Example Completed ===
```

## 📖 Example 2: Advanced Transaction Handling

**File:** `examples/transaction_example/main.go`

This example demonstrates:
- Complex transactions with multiple operations
- Transaction rollback on errors
- Multiple repository operations in a single transaction
- Error handling and recovery

### Key Features Demonstrated:

```go
// 1. Service Layer with Multiple Repositories
type ProductService struct {
    productFactory  persistence.IUnitOfWorkFactory[*Product]
    categoryFactory persistence.IUnitOfWorkFactory[*Category]
}

// 2. Transaction Management
func (s *ProductService) CreateCategoryWithProducts(ctx context.Context, categoryName string, products []*Product) error {
    categoryUow := s.categoryFactory.CreateWithContext(ctx)
    productUow := s.productFactory.CreateWithContext(ctx)
    
    // Begin transaction
    if err := categoryUow.BeginTransaction(ctx); err != nil {
        return fmt.Errorf("failed to begin transaction: %w", err)
    }
    
    // Setup rollback on error
    defer func() {
        if r := recover(); r != nil {
            categoryUow.RollbackTransaction(ctx)
            panic(r)
        }
    }()
    
    // Multiple operations...
    // Commit or rollback based on success/failure
}
```

### Running the Transaction Example:

```bash
go run examples/transaction_example/main.go
```

**Expected Output:**
```
=== Transaction Example ===

1. Creating category with products (successful transaction)...
Created category: &{ID:1 Name:Electronics Products:[]}
Created product: &{ID:1 Name:Laptop Price:999.99 CategoryID:1 Stock:10}
Created product: &{ID:2 Name:Mouse Price:29.99 CategoryID:1 Stock:50}
Created product: &{ID:3 Name:Keyboard Price:79.99 CategoryID:1 Stock:30}
Transaction committed successfully!

2. Transferring stock between products...
Successfully transferred 2 units from product 1 to product 2

3. Attempting to transfer more stock than available (should fail)...
Expected error occurred: insufficient stock: has 8, need 100

4. Verifying final state...
Product Laptop: Stock = 8
Product Mouse: Stock = 52
Product Keyboard: Stock = 30

=== Transaction Example Completed ===
```

## 📖 Example 3: Testing Patterns

**File:** `examples/testing_example/main.go`

This example demonstrates:
- Unit testing with the SDK
- Testing transaction scenarios
- Mocking and test isolation
- Performance benchmarking

### Key Testing Patterns:

```go
// 1. Test Setup with Isolated Database
func setupTestDB(t *testing.T) *gorm.DB {
    db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
    require.NoError(t, err)
    require.NoError(t, db.AutoMigrate(&TestUser{}))
    return db
}

// 2. Transaction Testing
func TestUserRepository_Transaction(t *testing.T) {
    db := setupTestDB(t)
    factory := postgres.NewUnitOfWorkFactory[*TestUser](db)
    
    uow := factory.CreateWithContext(context.Background())
    repo := NewTestUserRepository(uow)
    
    // Test transaction scenarios...
}

// 3. Performance Benchmarking
func BenchmarkUserRepository_Create(b *testing.B) {
    db := setupTestDB(&testing.T{})
    factory := postgres.NewUnitOfWorkFactory[*TestUser](db)
    // Run benchmark...
}
```

### Running the Tests:

```bash
# Run all tests
go test examples/testing_example/ -v

# Run tests with benchmarks
go test examples/testing_example/ -v -bench=.

# Run specific test
go test examples/testing_example/ -v -run TestUserRepository_CRUD
```

**Expected Output:**
```
=== RUN   TestUserRepository_CRUD
--- PASS: TestUserRepository_CRUD (0.01s)
=== RUN   TestUserRepository_Update
--- PASS: TestUserRepository_Update (0.01s)
=== RUN   TestUserRepository_Delete
--- PASS: TestUserRepository_Delete (0.01s)
...
BenchmarkUserRepository_Create-8       1000000    1023 ns/op
BenchmarkUserRepository_Read-8         2000000     512 ns/op
PASS
```

## 🏭 Production vs Demo Setup

### Demo Setup (SQLite)

All examples include SQLite setup for immediate execution:

```go
func setupDatabase() (*gorm.DB, error) {
    // SQLite for demonstration
    db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
    if err != nil {
        return nil, fmt.Errorf("failed to connect to database: %w", err)
    }
    return db, nil
}
```

### Production Setup (PostgreSQL)

For production, use PostgreSQL configuration:

```go
func setupProductionDatabase() (*postgres.Config, error) {
    config := postgres.NewConfig()
    config.Host = os.Getenv("DB_HOST")
    config.Port, _ = strconv.Atoi(os.Getenv("DB_PORT"))
    config.User = os.Getenv("DB_USER")
    config.Password = os.Getenv("DB_PASSWORD")
    config.Database = os.Getenv("DB_NAME")
    config.SSLMode = os.Getenv("DB_SSLMODE")
    
    return config, nil
}
```

### Environment Variables

Set these environment variables for PostgreSQL:

```bash
export DB_HOST=localhost
export DB_USER=postgres
export DB_PASSWORD=your_password
export DB_NAME=your_database
export DB_PORT=5432
export DB_SSLMODE=disable  # or 'require' for production
```

## ✅ Best Practices

### 1. Model Definition

Always implement the `BaseModel` interface:

```go
type YourModel struct {
    ID        int            `gorm:"primarykey"`
    Name      string         `gorm:"not null"`
    Slug      string         `gorm:"uniqueIndex;not null"`
    CreatedAt time.Time      `gorm:"autoCreateTime"`
    UpdatedAt time.Time      `gorm:"autoUpdateTime"`
    DeletedAt gorm.DeletedAt `gorm:"index"`
}

// Implement all required methods
func (m *YourModel) GetID() int { return m.ID }
func (m *YourModel) GetSlug() string { return m.Slug }
func (m *YourModel) SetSlug(slug string) { m.Slug = slug }
func (m *YourModel) GetCreatedAt() time.Time { return m.CreatedAt }
func (m *YourModel) GetUpdatedAt() time.Time { return m.UpdatedAt }
func (m *YourModel) GetArchivedAt() gorm.DeletedAt { return m.DeletedAt }
func (m *YourModel) GetName() string { return m.Name }
```

### 2. Repository Pattern

Use the service → repository → unit of work flow:

```go
// Service Layer
type UserService struct {
    userFactory persistence.IUnitOfWorkFactory[*User]
}

// Repository Layer
type UserRepository struct {
    uow persistence.IUnitOfWork[*User]
}

// Unit of Work handles database operations
func (r *UserRepository) Create(ctx context.Context, user *User) error {
    _, err := r.uow.Insert(ctx, user)
    return err
}
```

### 3. Error Handling

Always handle errors appropriately:

```go
func (s *UserService) CreateUser(ctx context.Context, user *User) error {
    uow := s.userFactory.CreateWithContext(ctx)
    repo := NewUserRepository(uow)
    
    if err := repo.Create(ctx, user); err != nil {
        return fmt.Errorf("failed to create user: %w", err)
    }
    
    return nil
}
```

### 4. Transaction Management

Use transactions for multi-operation scenarios:

```go
func (s *UserService) ComplexOperation(ctx context.Context) error {
    uow := s.userFactory.CreateWithContext(ctx)
    
    if err := uow.BeginTransaction(ctx); err != nil {
        return err
    }
    
    defer func() {
        if r := recover(); r != nil {
            uow.RollbackTransaction(ctx)
            panic(r)
        }
    }()
    
    // Multiple operations...
    
    return uow.CommitTransaction(ctx)
}
```

## 🔧 Troubleshooting

### Common Issues

**1. BaseModel Interface Not Implemented**

```
Error: *YourModel does not satisfy domain.BaseModel (missing method GetArchivedAt)
```

**Solution:** Ensure your model implements all BaseModel methods:

```go
func (m *YourModel) GetArchivedAt() gorm.DeletedAt {
    return m.DeletedAt
}
```

**2. Factory Configuration Issues**

```
Error: cannot use db (variable of type *gorm.DB) as *postgres.Config value
```

**Solution:** Use proper configuration:

```go
config := postgres.NewConfig()
factory := postgres.NewUnitOfWorkFactory[*YourModel](config)
```

**3. Missing Repository Methods**

```
Error: repo.Create undefined (type *YourRepository has no field or method Create)
```

**Solution:** Implement repository methods using the UnitOfWork:

```go
func (r *YourRepository) Create(ctx context.Context, entity *YourModel) error {
    _, err := r.uow.Insert(ctx, entity)
    return err
}
```

### Getting Help

1. Check the examples in `examples/` directory
2. Review the test files for patterns
3. Consult the API documentation
4. Open an issue on GitHub

## 🎯 Next Steps

After running these examples:

1. **Adapt to your models:** Modify the User/Product models to match your domain
2. **Configure PostgreSQL:** Set up your production database
3. **Write tests:** Use the testing patterns from Example 3
4. **Build services:** Create service layers using the repository pattern
5. **Add features:** Implement advanced querying, caching, etc.

Happy coding! 🚀
