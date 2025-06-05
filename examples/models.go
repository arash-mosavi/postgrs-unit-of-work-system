package examples

import (
	"time"

	"gorm.io/gorm"
)

// User demonstrates a typical domain entity implementing BaseModel
type User struct {
	ID        int            `gorm:"primaryKey;autoIncrement" json:"id"`
	Name      string         `gorm:"size:100;not null" json:"name"`
	Email     string         `gorm:"size:100;uniqueIndex;not null" json:"email"`
	Slug      string         `gorm:"size:100;uniqueIndex;not null" json:"slug"`
	CreatedAt time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`

	// Relationships
	Posts []Post `gorm:"foreignKey:UserID" json:"posts,omitempty"`
}

// Implement BaseModel interface
func (u *User) GetID() int {
	return u.ID
}

func (u *User) GetSlug() string {
	return u.Slug
}

func (u *User) SetSlug(slug string) {
	u.Slug = slug
}

func (u *User) GetCreatedAt() time.Time {
	return u.CreatedAt
}

func (u *User) GetUpdatedAt() time.Time {
	return u.UpdatedAt
}

func (u *User) GetArchivedAt() gorm.DeletedAt {
	return u.DeletedAt
}

func (u *User) GetName() string {
	return u.Name
}

// Post demonstrates a related entity with foreign key
type Post struct {
	ID        int            `gorm:"primaryKey;autoIncrement" json:"id"`
	Name      string         `gorm:"size:200;not null" json:"name"`
	Content   string         `gorm:"type:text" json:"content"`
	Slug      string         `gorm:"size:200;uniqueIndex;not null" json:"slug"`
	UserID    int            `gorm:"not null;index" json:"user_id"`
	CreatedAt time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`

	// Relationships
	User User  `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Tags []Tag `gorm:"many2many:post_tags;" json:"tags,omitempty"`
}

// Implement BaseModel interface
func (p *Post) GetID() int {
	return p.ID
}

func (p *Post) GetSlug() string {
	return p.Slug
}

func (p *Post) SetSlug(slug string) {
	p.Slug = slug
}

func (p *Post) GetCreatedAt() time.Time {
	return p.CreatedAt
}

func (p *Post) GetUpdatedAt() time.Time {
	return p.UpdatedAt
}

func (p *Post) GetArchivedAt() gorm.DeletedAt {
	return p.DeletedAt
}

func (p *Post) GetName() string {
	return p.Name
}

// Tag demonstrates many-to-many relationships
type Tag struct {
	ID        int            `gorm:"primaryKey;autoIncrement" json:"id"`
	Name      string         `gorm:"size:50;uniqueIndex;not null" json:"name"`
	Slug      string         `gorm:"size:50;uniqueIndex;not null" json:"slug"`
	CreatedAt time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`

	// Relationships
	Posts []Post `gorm:"many2many:post_tags;" json:"posts,omitempty"`
}

// Implement BaseModel interface
func (t *Tag) GetID() int {
	return t.ID
}

func (t *Tag) GetSlug() string {
	return t.Slug
}

func (t *Tag) SetSlug(slug string) {
	t.Slug = slug
}

func (t *Tag) GetCreatedAt() time.Time {
	return t.CreatedAt
}

func (t *Tag) GetUpdatedAt() time.Time {
	return t.UpdatedAt
}

func (t *Tag) GetArchivedAt() gorm.DeletedAt {
	return t.DeletedAt
}

func (t *Tag) GetName() string {
	return t.Name
}
