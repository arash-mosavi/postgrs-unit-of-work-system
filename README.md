# PostgreSQL Unit of Work System

This is a Go SDK for working with PostgreSQL using the Unit of Work pattern.

## Quick Start

### Install

```bash
go get github.com/arash-mosavi/postgrs-unit-of-work-system
```

### Basic Example

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
    config.User = "postgres"
    config.Password = "password"
    config.Database = "myapp"
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

    log.Println("User and posts created successfully.")
}
```

## Example Project

You can run a real working demo from:

```bash
cd examples/basic_example
go run main.go
```

It shows:
- SQLite or PostgreSQL connection
- Full CRUD
- Service logic with transaction handling
- How BaseModel and Unit of Work connect

## Core Concepts

- Service owns business logic and starts a UnitOfWork
- Repositories handle actual database operations
- All changes are applied together when you call `CommitTransaction`
- Clean separation of concerns

## Config

```go
config := &postgres.Config{
    Host:     "localhost",
    Port:     5432,
    Database: "myapp",
    User:     "postgres",
    Password: "password",
    SSLMode:  "disable",
}
```

## Features

- Type-safe, generic UoW factories
- Transaction rollback by default
- Works with GORM
- Batch inserts
- Filtering/sorting helpers
- Clean structure and testable services

## Testing

```bash
go test ./...
go test -v ./pkg/postgres
go test -bench=. ./pkg/postgres
```

## Layout

```
pkg/
  postgres/         # Postgres impl
  persistence/      # Core interfaces
  errors/           # Error wrapping
  identifier/       # Filter builder
examples/           # Example services
```
