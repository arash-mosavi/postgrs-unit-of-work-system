# SDK Setup and Distribution Guide

This guide explains how to prepare and distribute the PostgreSQL Unit of Work System as a Go module.

## 🚀 Publishing to GitHub

### 1. Initialize Git Repository
```bash
git init
git add .
git commit -m "Initial commit: PostgreSQL Unit of Work System SDK"
```

### 2. Add Remote Repository
```bash
git remote add origin https://github.com/arash-mosavi/postgrs-unit-of-work-system.git
git branch -M main
git push -u origin main
```

### 3. Create and Push Version Tags
```bash
# Create initial version tag
git tag v1.0.0
git push origin v1.0.0

# For subsequent releases
git tag v1.0.1
git push origin v1.0.1
```

## 📦 Module Installation for Users

Once published, users can install the SDK using:

```bash
go get github.com/arash-mosavi/postgrs-unit-of-work-system
```

For specific versions:
```bash
go get github.com/arash-mosavi/postgrs-unit-of-work-system@v1.0.0
```

## 🧪 Testing the SDK

### Local Testing
Users can test the SDK locally after installation:

```bash
# Create a new Go project
mkdir my-app && cd my-app
go mod init my-app

# Install the SDK
go get github.com/arash-mosavi/postgrs-unit-of-work-system

# Create main.go and use the SDK
```

### Example Usage in User Project

```go
package main

import (
    "context"
    "log"
    
    "github.com/arash-mosavi/postgrs-unit-of-work-system/pkg/postgres"
    "github.com/arash-mosavi/postgrs-unit-of-work-system/examples"
)

func main() {
    config := postgres.NewConfig()
    config.Host = "localhost"
    config.Port = 5432
    config.Database = "myapp"
    config.User = "postgres"
    config.Password = "password"
    config.SSLMode = "disable"

    userFactory := postgres.NewUnitOfWorkFactory[*examples.User](config)
    postFactory := postgres.NewUnitOfWorkFactory[*examples.Post](config)
    
    userService := examples.NewUserService(userFactory, postFactory)
    
    ctx := context.Background()
    user := &examples.User{
        Name:  "John Doe",
        Email: "john@example.com", 
        Slug:  "john-doe",
    }
    
    posts := []*examples.Post{
        {Name: "First Post", Content: "Hello World", Slug: "first-post"},
    }
    
    if err := userService.CreateUserWithPosts(ctx, user, posts); err != nil {
        log.Fatal(err)
    }
    
    log.Println("User and posts created successfully!")
}
```

## 📋 SDK Verification

### Before Publishing
Run these commands to ensure the SDK is ready:

```bash
# Build all packages
go build ./...

# Run all tests
go test ./...

# Run validation
go run validation.go

# Verify module info
go list -m all
```

### After Publishing
Users can verify the SDK works:

```bash
# Check module info
go list -m github.com/arash-mosavi/postgrs-unit-of-work-system

# Download and verify
go mod download github.com/arash-mosavi/postgrs-unit-of-work-system
```

## 🔄 Version Management

### Semantic Versioning
- `v1.0.0` - Initial stable release
- `v1.0.1` - Patch updates (bug fixes)
- `v1.1.0` - Minor updates (new features, backward compatible)
- `v2.0.0` - Major updates (breaking changes)

### Release Process
1. Update CHANGELOG.md
2. Update version in documentation
3. Create and test release branch
4. Create git tag
5. Push tag to trigger release

## 🏷️ Go Module Proxy

Once published, the module will be available through:
- `proxy.golang.org` (default Go module proxy)
- `goproxy.io`
- Direct GitHub access

## 📊 Usage Analytics

Monitor SDK usage through:
- GitHub repository insights
- Go module proxy stats
- pkg.go.dev analytics

---

**Note**: Ensure all sensitive information (passwords, keys) are removed from examples before publishing.
