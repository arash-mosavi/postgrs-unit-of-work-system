package identifier

import (
	"fmt"
	"strings"
)

// IIdentifier defines the interface for query building and identification
type IIdentifier interface {
	// Query building methods
	Equal(field string, value interface{}) IIdentifier
	In(field string, values []interface{}) IIdentifier
	Like(field string, pattern string) IIdentifier
	GreaterThan(field string, value interface{}) IIdentifier
	LessThan(field string, value interface{}) IIdentifier
	Between(field string, start, end interface{}) IIdentifier
	IsNull(field string) IIdentifier
	IsNotNull(field string) IIdentifier

	// Utility methods
	Add(key string, value interface{}) IIdentifier
	AddIf(condition bool, key string, value interface{}) IIdentifier

	// Query access methods
	ToMap() map[string]interface{}
	ToSQL() (string, []interface{})
	GetQuery() map[string]interface{}
	Has(key string) bool
	Get(key string) (interface{}, bool)
	String() string
}

// Identifier provides flexible query building with O(1) operations
type Identifier struct {
	query map[string]interface{}
}

// New creates a new identifier instance
func New() *Identifier {
	return &Identifier{
		query: make(map[string]interface{}),
	}
}

// Equal adds an equality condition
func (i *Identifier) Equal(field string, value interface{}) IIdentifier {
	i.query[field] = value
	return i
}

// In adds an IN condition
func (i *Identifier) In(field string, values []interface{}) IIdentifier {
	i.query[field+" IN"] = values
	return i
}

// Like adds a LIKE condition
func (i *Identifier) Like(field string, pattern string) IIdentifier {
	i.query[field+" LIKE"] = pattern
	return i
}

// GreaterThan adds a > condition
func (i *Identifier) GreaterThan(field string, value interface{}) IIdentifier {
	i.query[field+" >"] = value
	return i
}

// LessThan adds a < condition
func (i *Identifier) LessThan(field string, value interface{}) IIdentifier {
	i.query[field+" <"] = value
	return i
}

// Between adds a BETWEEN condition
func (i *Identifier) Between(field string, start, end interface{}) IIdentifier {
	i.query[field+" BETWEEN"] = []interface{}{start, end}
	return i
}

// IsNull adds an IS NULL condition
func (i *Identifier) IsNull(field string) IIdentifier {
	i.query[field+" IS NULL"] = true
	return i
}

// IsNotNull adds an IS NOT NULL condition
func (i *Identifier) IsNotNull(field string) IIdentifier {
	i.query[field+" IS NOT NULL"] = true
	return i
}

// Add adds a key-value pair to the query
func (i *Identifier) Add(key string, value interface{}) IIdentifier {
	i.query[key] = value
	return i
}

// AddIf conditionally adds a key-value pair
func (i *Identifier) AddIf(condition bool, key string, value interface{}) IIdentifier {
	if condition {
		i.Add(key, value)
	}
	return i
}

// ToMap returns the query map for use with GORM
func (i *Identifier) ToMap() map[string]interface{} {
	return i.query
}

// ToSQL converts the identifier to SQL conditions (basic implementation)
func (i *Identifier) ToSQL() (string, []interface{}) {
	var conditions []string
	var args []interface{}

	for key, value := range i.query {
		if strings.Contains(key, " ") {
			// Handle operators
			parts := strings.SplitN(key, " ", 2)
			field, operator := parts[0], parts[1]

			switch operator {
			case "IN":
				if vals, ok := value.([]interface{}); ok {
					placeholders := strings.Repeat("?,", len(vals)-1) + "?"
					conditions = append(conditions, fmt.Sprintf("%s IN (%s)", field, placeholders))
					args = append(args, vals...)
				}
			case "LIKE":
				conditions = append(conditions, fmt.Sprintf("%s LIKE ?", field))
				args = append(args, value)
			case ">", "<", ">=", "<=":
				conditions = append(conditions, fmt.Sprintf("%s %s ?", field, operator))
				args = append(args, value)
			case "BETWEEN":
				if vals, ok := value.([]interface{}); ok && len(vals) == 2 {
					conditions = append(conditions, fmt.Sprintf("%s BETWEEN ? AND ?", field))
					args = append(args, vals[0], vals[1])
				}
			case "IS NULL", "IS NOT NULL":
				conditions = append(conditions, fmt.Sprintf("%s %s", field, operator))
			}
		} else {
			// Simple equality
			conditions = append(conditions, fmt.Sprintf("%s = ?", key))
			args = append(args, value)
		}
	}

	return strings.Join(conditions, " AND "), args
}

// Convenience constructors
func ByID(id interface{}) IIdentifier {
	return New().Equal("id", id)
}

func BySlug(slug string) IIdentifier {
	return New().Equal("slug", slug)
}

func ByEmail(email string) IIdentifier {
	return New().Equal("email", email)
}

func Active() IIdentifier {
	return New().Equal("active", true)
}

func Inactive() IIdentifier {
	return New().Equal("active", false)
}

// NewIdentifier creates a new identifier
func NewIdentifier() IIdentifier {
	return &Identifier{
		query: make(map[string]interface{}),
	}
}

// GetQuery returns the query map
func (i *Identifier) GetQuery() map[string]interface{} {
	result := make(map[string]interface{}, len(i.query))
	for k, v := range i.query {
		result[k] = v
	}
	return result
}

// String returns a string representation
func (i *Identifier) String() string {
	if len(i.query) == 0 {
		return "{}"
	}

	var builder strings.Builder
	builder.WriteString("{")

	first := true
	for key, value := range i.query {
		if !first {
			builder.WriteString(", ")
		}
		builder.WriteString(fmt.Sprintf("%s: %v", key, value))
		first = false
	}

	builder.WriteString("}")
	return builder.String()
}

// Has checks if a key exists
func (i *Identifier) Has(key string) bool {
	_, exists := i.query[key]
	return exists
}

// Get retrieves a value by key
func (i *Identifier) Get(key string) (interface{}, bool) {
	value, exists := i.query[key]
	return value, exists
}

// NewIDIdentifier creates an identifier for ID-based queries
func NewIDIdentifier(id int64) IIdentifier {
	return NewIdentifier().Add("id", id)
}

// NewSlugIdentifier creates an identifier for slug-based queries
func NewSlugIdentifier(slug string) IIdentifier {
	return NewIdentifier().Add("slug", slug)
}
