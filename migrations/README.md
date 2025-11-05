# Go Migrations

This directory contains Go-based PocketBase migrations converted from the JavaScript migrations in `pb_migrations/`.

## Why Go Migrations?

Go migrations are:
- **Type-safe**: Compile-time checking prevents errors
- **Performant**: No JavaScript runtime overhead
- **Integrated**: Direct access to PocketBase API
- **Portable**: Compiled into the binary

## Migration Files

Each migration file follows the naming convention: `{timestamp}_{description}.go`

- `1762021170_created_servers.go` - Creates the servers collection
- `1762021448_created_matches.go` - Creates the matches collection with indexes
- `1762032953_created_player.go` - Creates the players collection
- `1762033323_created_match_player_stats.go` - Creates match_player_stats with relations
- `1762033504_created_match_weapon_stats.go` - Creates match_weapon_stats with indexes

## Usage

Migrations are automatically registered via `init()` functions when you import the package:

```go
import (
    _ "sandstorm-tracker/migrations"
)
```

This is done in `main.go`.

## Running Migrations

Migrations run automatically when the app starts. PocketBase tracks which migrations have been applied in the `_migrations` table.

To manually run migrations:
```bash
./sandstorm-tracker migrate up
```

To check migration status:
```bash
./sandstorm-tracker migrate collections
```

## Creating New Migrations

1. Use the PocketBase admin UI to make schema changes
2. Export the generated JavaScript migration
3. Convert to Go using this pattern:

```go
package migrations

import (
	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(app core.App) error {
		// Migration up logic
		collection := core.NewBaseCollection("my_collection")
		
		collection.Fields.Add(
			&core.TextField{
				Name:     "title",
				Required: true,
			},
		)
		
		return app.Save(collection)
	}, func(app core.App) error {
		// Migration down logic (rollback)
		collection, err := app.FindCollectionByNameOrId("my_collection")
		if err != nil || collection == nil {
			return nil
		}
		return app.Delete(collection)
	})
}
```

## Field Types

Common PocketBase field types:

- `core.TextField` - Text fields
- `core.NumberField` - Numeric fields
- `core.BoolField` - Boolean fields
- `core.DateField` - Date/time fields
- `core.AutodateField` - Auto-updated timestamps
- `core.RelationField` - Relations to other collections
- `core.SelectField` - Dropdown/select fields
- `core.FileField` - File uploads
- `core.JSONField` - JSON data

## References

- [PocketBase Go Migrations Documentation](https://pocketbase.io/docs/go-migrations/)
- [PocketBase Collections API](https://pocketbase.io/docs/api-collections/)
