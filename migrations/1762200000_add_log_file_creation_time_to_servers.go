package migrations

import (
	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(app core.App) error {
		collection, err := app.FindCollectionByNameOrId("servers")
		if err != nil {
			return err
		}

		// Add log_file_creation_time field
		collection.Fields.Add(&core.TextField{
			Name:     "log_file_creation_time",
			Required: false,
		})

		return app.Save(collection)
	}, func(app core.App) error {
		collection, err := app.FindCollectionByNameOrId("servers")
		if err != nil {
			return err
		}

		// Remove log_file_creation_time field on rollback
		collection.Fields.RemoveById("log_file_creation_time")

		return app.Save(collection)
	})
}
