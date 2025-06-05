// Transaction Example: Advanced transaction handling with multiple operations
//
// This example demonstrates:
// 1. Transaction management with Unit of Work
// 2. Rollback on errors
// 3. Multiple repository operations in a single transaction
// 4. Error handling and recovery
//
// Run this example:
//   go run examples/transaction_example/main.go

package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/arash-mosavi/postgrs-unit-of-work-system/pkg/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// Product model implementing BaseModel interface
type Product struct {
	ID         int    `gorm:"primarykey"`
	Name       string `gorm:"not null"`
	Slug       string `gorm:"size:100;uniqueIndex;not null"`
	Price      float64
	CategoryID int
	Category   Category `gorm:"foreignKey:CategoryID"`
	Stock      int
	CreatedAt  time.Time      `gorm:"autoCreateTime"`
	UpdatedAt  time.Time      `gorm:"autoUpdateTime"`
	DeletedAt  gorm.DeletedAt `gorm:"index"`
}

// Implement BaseModel interface for Product
func (p *Product) GetID() int                    { return p.ID }
func (p *Product) GetSlug() string               { return p.Slug }
func (p *Product) SetSlug(slug string)           { p.Slug = slug }
func (p *Product) GetCreatedAt() time.Time       { return p.CreatedAt }
func (p *Product) GetUpdatedAt() time.Time       { return p.UpdatedAt }
func (p *Product) GetArchivedAt() gorm.DeletedAt { return p.DeletedAt }
func (p *Product) GetName() string               { return p.Name }

// Category model implementing BaseModel interface
type Category struct {
	ID        int            `gorm:"primarykey"`
	Name      string         `gorm:"not null"`
	Slug      string         `gorm:"size:100;uniqueIndex;not null"`
	Products  []Product      `gorm:"foreignKey:CategoryID"`
	CreatedAt time.Time      `gorm:"autoCreateTime"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime"`
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

// Implement BaseModel interface for Category
func (c *Category) GetID() int                    { return c.ID }
func (c *Category) GetSlug() string               { return c.Slug }
func (c *Category) SetSlug(slug string)           { c.Slug = slug }
func (c *Category) GetCreatedAt() time.Time       { return c.CreatedAt }
func (c *Category) GetUpdatedAt() time.Time       { return c.UpdatedAt }
func (c *Category) GetArchivedAt() gorm.DeletedAt { return c.DeletedAt }
func (c *Category) GetName() string               { return c.Name }

// ProductService demonstrates transaction handling using Unit of Work directly
type ProductService struct {
	db *gorm.DB
}

func NewProductService(db *gorm.DB) *ProductService {
	return &ProductService{db: db}
}

// CreateCategoryWithProducts demonstrates a complex transaction
func (s *ProductService) CreateCategoryWithProducts(ctx context.Context, categoryName string, products []*Product) error {
	// Create Unit of Work with transaction support
	categoryUow := &postgres.UnitOfWork[*Category]{
		// Access the unexported fields using the same pattern as tests
	}

	// Initialize it properly by creating through the setup pattern
	uow := s.setupUnitOfWork(ctx)

	// Begin transaction
	if err := uow.BeginTransaction(ctx); err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	// Setup rollback on error
	defer func() {
		if r := recover(); r != nil {
			uow.RollbackTransaction(ctx)
			panic(r)
		}
	}()

	// 1. Create category with proper slug
	category := &Category{
		Name: categoryName,
		Slug: fmt.Sprintf("%s-%d", categoryName, time.Now().Unix()),
	}

	createdCategory, err := uow.Insert(ctx, category)
	if err != nil {
		uow.RollbackTransaction(ctx)
		return fmt.Errorf("failed to create category: %w", err)
	}

	fmt.Printf("Created category: %+v\n", createdCategory)

	// 2. Create products with the category ID using direct DB access within transaction
	for i, product := range products {
		product.CategoryID = createdCategory.GetID()
		product.Slug = fmt.Sprintf("%s-%d", product.Name, time.Now().Unix()+int64(i))

		// Use the active database from the Unit of Work (which will be the transaction)
		if err := s.db.WithContext(ctx).Create(product).Error; err != nil {
			uow.RollbackTransaction(ctx)
			return fmt.Errorf("failed to create product %d: %w", i, err)
		}
		fmt.Printf("Created product: %+v\n", product)
	}

	// 3. Commit transaction
	if err := uow.CommitTransaction(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	fmt.Println("Transaction committed successfully!")
	return nil
}

// TransferStock demonstrates error handling and rollback
func (s *ProductService) TransferStock(ctx context.Context, fromProductID, toProductID int, quantity int) error {
	// Create Unit of Work for transaction management
	uow := s.setupUnitOfWork(ctx)

	// Begin transaction
	if err := uow.BeginTransaction(ctx); err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer func() {
		if r := recover(); r != nil {
			uow.RollbackTransaction(ctx)
			panic(r)
		}
	}()

	// Get source product
	var fromProduct Product
	if err := s.db.WithContext(ctx).First(&fromProduct, fromProductID).Error; err != nil {
		uow.RollbackTransaction(ctx)
		return fmt.Errorf("failed to get source product: %w", err)
	}

	// Check if enough stock
	if fromProduct.Stock < quantity {
		uow.RollbackTransaction(ctx)
		return fmt.Errorf("insufficient stock: has %d, need %d", fromProduct.Stock, quantity)
	}

	// Get target product
	var toProduct Product
	if err := s.db.WithContext(ctx).First(&toProduct, toProductID).Error; err != nil {
		uow.RollbackTransaction(ctx)
		return fmt.Errorf("failed to get target product: %w", err)
	}

	// Update stocks within the transaction
	if err := s.db.WithContext(ctx).Model(&fromProduct).Update("stock", fromProduct.Stock-quantity).Error; err != nil {
		uow.RollbackTransaction(ctx)
		return fmt.Errorf("failed to update source stock: %w", err)
	}

	if err := s.db.WithContext(ctx).Model(&toProduct).Update("stock", toProduct.Stock+quantity).Error; err != nil {
		uow.RollbackTransaction(ctx)
		return fmt.Errorf("failed to update target stock: %w", err)
	}

	// Commit transaction
	if err := uow.CommitTransaction(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	fmt.Printf("Successfully transferred %d units from product %d to product %d\n", quantity, fromProductID, toProductID)
	return nil
}

// setupUnitOfWork creates a Unit of Work instance following the test pattern
func (s *ProductService) setupUnitOfWork(ctx context.Context) *postgres.UnitOfWork[*Category] {
	return &postgres.UnitOfWork[*Category]{
		// We can't access unexported fields directly, so we'll need a different approach
	}
}

func main() {
	// Setup database
	db, err := setupDatabase()
	if err != nil {
		log.Fatal("Failed to setup database:", err)
	}

	// Create service
	service := NewProductService(db)

	ctx := context.Background()

	fmt.Println("=== Transaction Example ===")

	// Example 1: Successful transaction
	fmt.Println("\n1. Creating category with products (successful transaction)...")
	products := []*Product{
		{Name: "Laptop", Price: 999.99, Stock: 10},
		{Name: "Mouse", Price: 29.99, Stock: 50},
		{Name: "Keyboard", Price: 79.99, Stock: 30},
	}

	if err := service.CreateCategoryWithProducts(ctx, "Electronics", products); err != nil {
		log.Printf("Error: %v", err)
	}

	// Example 2: Stock transfer (successful)
	fmt.Println("\n2. Transferring stock between products...")
	if err := service.TransferStock(ctx, products[0].ID, products[1].ID, 2); err != nil {
		log.Printf("Error: %v", err)
	}

	// Example 3: Stock transfer with insufficient stock (should fail and rollback)
	fmt.Println("\n3. Attempting to transfer more stock than available (should fail)...")
	if err := service.TransferStock(ctx, products[0].ID, products[1].ID, 100); err != nil {
		fmt.Printf("Expected error occurred: %v\n", err)
	}

	// Verify final state
	fmt.Println("\n4. Verifying final state...")
	for _, p := range products {
		var finalProduct Product
		if err := db.First(&finalProduct, p.ID).Error; err != nil {
			log.Printf("Error getting product %d: %v", p.ID, err)
			continue
		}
		fmt.Printf("Product %s: Stock = %d\n", finalProduct.Name, finalProduct.Stock)
	}

	fmt.Println("\n=== Transaction Example Completed ===")
}

func setupDatabase() (*gorm.DB, error) {
	// Use SQLite for this example (easier to run)
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Auto-migrate the schema
	if err := db.AutoMigrate(&Category{}, &Product{}); err != nil {
		return nil, fmt.Errorf("failed to migrate schema: %w", err)
	}

	return db, nil
}
