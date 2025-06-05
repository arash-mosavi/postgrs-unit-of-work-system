// Transaction Example: Advanced transaction handling with multiple operations
//
// This example demonstrates:
// 1. Transaction management with proper error handling
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
	"os"
	"time"

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

// ProductService demonstrates transaction handling with Unit of Work pattern
type ProductService struct {
	db *gorm.DB
}

func NewProductService(db *gorm.DB) *ProductService {
	return &ProductService{db: db}
}

// CreateProductWithCategory demonstrates creating related entities in a single transaction
func (s *ProductService) CreateProductWithCategory(ctx context.Context, categoryName, productName string, price float64, stock int) (*Product, *Category, error) {
	var product *Product
	var category *Category

	err := s.db.Transaction(func(tx *gorm.DB) error {
		// First, create or get the category
		category = &Category{
			Name: categoryName,
			Slug: fmt.Sprintf("%s-category", categoryName),
		}

		if err := tx.WithContext(ctx).Create(category).Error; err != nil {
			return fmt.Errorf("failed to create category: %w", err)
		}

		// Then create the product
		product = &Product{
			Name:       productName,
			Slug:       fmt.Sprintf("%s-product", productName),
			Price:      price,
			CategoryID: category.ID,
			Stock:      stock,
		}

		if err := tx.WithContext(ctx).Create(product).Error; err != nil {
			return fmt.Errorf("failed to create product: %w", err)
		}

		return nil
	})

	return product, category, err
}

// UpdateProductStock demonstrates updating with transaction boundaries
func (s *ProductService) UpdateProductStock(ctx context.Context, productID int, newStock int) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		var product Product
		if err := tx.WithContext(ctx).First(&product, productID).Error; err != nil {
			return fmt.Errorf("product not found: %w", err)
		}

		product.Stock = newStock
		if err := tx.WithContext(ctx).Save(&product).Error; err != nil {
			return fmt.Errorf("failed to update stock: %w", err)
		}

		return nil
	})
}

// ProcessOrder demonstrates complex transaction with potential rollback
func (s *ProductService) ProcessOrder(ctx context.Context, productID int, quantity int) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		var product Product
		if err := tx.WithContext(ctx).First(&product, productID).Error; err != nil {
			return fmt.Errorf("product not found: %w", err)
		}

		// Check if we have enough stock
		if product.Stock < quantity {
			return fmt.Errorf("insufficient stock: available=%d, requested=%d", product.Stock, quantity)
		}

		// Update stock
		product.Stock -= quantity
		if err := tx.WithContext(ctx).Save(&product).Error; err != nil {
			return fmt.Errorf("failed to update stock: %w", err)
		}

		// Here you would typically also create an order record
		// For this example, we'll just log the transaction
		log.Printf("Order processed: Product ID %d, Quantity %d, Remaining Stock %d",
			productID, quantity, product.Stock)

		return nil
	})
}

func main() {
	ctx := context.Background()

	fmt.Println("=== Transaction Example: Unit of Work Pattern ===")
	fmt.Println("This example demonstrates complex transaction handling with multiple entities")

	// Setup database
	db, err := setupDatabase()
	if err != nil {
		log.Fatal("Failed to setup database:", err)
	}

	service := NewProductService(db)

	// Example 1: Create product with category in a single transaction
	fmt.Println("\n--- Example 1: Creating Product with Category ---")
	product, category, err := service.CreateProductWithCategory(ctx, "Electronics", "Smartphone", 699.99, 50)
	if err != nil {
		log.Printf("❌ Failed to create product with category: %v", err)
	} else {
		fmt.Printf("✅ Created category: %s (ID: %d)\n", category.Name, category.ID)
		fmt.Printf("✅ Created product: %s (ID: %d, Price: $%.2f, Stock: %d)\n",
			product.Name, product.ID, product.Price, product.Stock)
	}

	// Example 2: Update product stock
	fmt.Println("\n--- Example 2: Updating Product Stock ---")
	if product != nil {
		err = service.UpdateProductStock(ctx, product.ID, 75)
		if err != nil {
			fmt.Printf("❌ Failed to update stock: %v\n", err)
		} else {
			fmt.Printf("✅ Updated product stock to 75\n")
		}
	}

	// Example 3: Process order (successful)
	fmt.Println("\n--- Example 3: Processing Order (Successful) ---")
	if product != nil {
		err = service.ProcessOrder(ctx, product.ID, 25)
		if err != nil {
			fmt.Printf("❌ Failed to process order: %v\n", err)
		} else {
			fmt.Printf("✅ Order processed successfully\n")
		}
	}

	// Example 4: Process order (insufficient stock - should rollback)
	fmt.Println("\n--- Example 4: Processing Order (Insufficient Stock) ---")
	if product != nil {
		err = service.ProcessOrder(ctx, product.ID, 100) // More than available stock
		if err != nil {
			fmt.Printf("✅ Order failed as expected: %v\n", err)
		} else {
			fmt.Printf("❌ Order should have failed due to insufficient stock\n")
		}
	}

	// Example 5: Multiple operations in transaction
	fmt.Println("\n--- Example 5: Multiple Operations in Single Transaction ---")
	err = service.db.Transaction(func(tx *gorm.DB) error {
		// Create another category
		category2 := &Category{
			Name: "Books",
			Slug: "books-category",
		}
		if err := tx.WithContext(ctx).Create(category2).Error; err != nil {
			return err
		}

		// Create multiple products
		products := []*Product{
			{Name: "Go Programming Book", Slug: "go-book", Price: 39.99, CategoryID: category2.ID, Stock: 20},
			{Name: "Database Design Book", Slug: "db-book", Price: 49.99, CategoryID: category2.ID, Stock: 15},
		}

		for _, p := range products {
			if err := tx.WithContext(ctx).Create(p).Error; err != nil {
				return fmt.Errorf("failed to create product %s: %w", p.Name, err)
			}
		}

		return nil
	})

	if err != nil {
		fmt.Printf("❌ Multiple operations failed: %v\n", err)
	} else {
		fmt.Printf("✅ Multiple operations completed successfully\n")
	}

	fmt.Println("\n=== Transaction Example Completed ===")
	fmt.Println("\nUnit of Work Pattern Benefits Demonstrated:")
	fmt.Println("- Atomic operations (all or nothing)")
	fmt.Println("- Automatic rollback on errors")
	fmt.Println("- Consistency across related entities")
	fmt.Println("- Proper error handling and recovery")
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

// Example of environment-based configuration
func getDatabaseConfig() map[string]string {
	return map[string]string{
		"host":     getEnv("DB_HOST", "localhost"),
		"user":     getEnv("DB_USER", "postgres"),
		"password": getEnv("DB_PASSWORD", "postgres"),
		"dbname":   getEnv("DB_NAME", "testdb"),
		"port":     getEnv("DB_PORT", "5432"),
		"sslmode":  getEnv("DB_SSLMODE", "disable"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
