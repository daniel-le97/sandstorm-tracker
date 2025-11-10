package migrations

import (
	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(app core.App) error {
		collection, err := app.FindCollectionByNameOrId("pbc_3080700301")
		if err != nil {
			return err
		}

		// add new "status" field
		collection.Fields.Add(&core.SelectField{
			Id:          "select4169818199",
			Name:        "status",
			Required:    true,
			Presentable: false,
			MaxSelect:   1,
			Values:      []string{"ongoing", "disconnected", "finished"},
		})

		return app.Save(collection)
	}, func(app core.App) error {
		collection, err := app.FindCollectionByNameOrId("pbc_3080700301")
		if err != nil {
			return err
		}

		// remove "status" field
		collection.Fields.RemoveById("select4169818199")

		return app.Save(collection)
	})
}
