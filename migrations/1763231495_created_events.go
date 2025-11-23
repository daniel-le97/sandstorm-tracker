package migrations

import (
	"encoding/json"

	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(app core.App) error {
		jsonData := `{
			"createRule": null,
			"deleteRule": null,
			"fields": [
				{
					"autogeneratePattern": "[a-z0-9]{15}",
					"hidden": false,
					"id": "text3208210256",
					"max": 15,
					"min": 15,
					"name": "id",
					"pattern": "^[a-z0-9]+$",
					"presentable": false,
					"primaryKey": true,
					"required": true,
					"system": true,
					"type": "text"
				},
				{
					"autogeneratePattern": "",
					"hidden": false,
					"id": "text2363381545",
					"max": 0,
					"min": 0,
					"name": "type",
					"pattern": "",
					"presentable": false,
					"primaryKey": false,
					"required": false,
					"system": false,
					"type": "text"
				},
				{
					"hidden": false,
					"id": "json2918445923",
					"maxSize": 0,
					"name": "data",
					"presentable": false,
					"required": false,
					"system": false,
					"type": "json"
				},
				{
					"cascadeDelete": false,
					"collectionId": "pbc_3738798621",
					"hidden": false,
					"id": "relation1517147638",
					"maxSelect": 1,
					"minSelect": 0,
					"name": "server",
					"presentable": false,
					"required": false,
					"system": false,
					"type": "relation"
				},
				{
					"hidden": false,
					"id": "autodate2990389176",
					"name": "created",
					"onCreate": true,
					"onUpdate": false,
					"presentable": false,
					"system": false,
					"type": "autodate"
				},
				{
					"hidden": false,
					"id": "autodate3332085495",
					"name": "updated",
					"onCreate": true,
					"onUpdate": true,
					"presentable": false,
					"system": false,
					"type": "autodate"
				}
			],
			"id": "pbc_1687431684",
			"indexes": [
				"CREATE INDEX ` + "`" + `idx_events_type` + "`" + ` ON ` + "`" + `events` + "`" + ` (` + "`" + `type` + "`" + `)",
				"CREATE INDEX ` + "`" + `idx_events_server` + "`" + ` ON ` + "`" + `events` + "`" + ` (` + "`" + `server` + "`" + `)",
				"CREATE INDEX ` + "`" + `idx_events_created` + "`" + ` ON ` + "`" + `events` + "`" + ` (` + "`" + `created` + "`" + `)",
				"CREATE INDEX ` + "`" + `idx_events_server_type` + "`" + ` ON ` + "`" + `events` + "`" + ` (\n  ` + "`" + `server` + "`" + `,\n  ` + "`" + `type` + "`" + `\n)"
			],
			"listRule": "",
			"name": "events",
			"system": false,
			"type": "base",
			"updateRule": null,
			"viewRule": ""
		}`

		collection := &core.Collection{}
		if err := json.Unmarshal([]byte(jsonData), &collection); err != nil {
			return err
		}

		return app.Save(collection)
	}, func(app core.App) error {
		collection, err := app.FindCollectionByNameOrId("pbc_1687431684")
		if err != nil {
			return err
		}

		return app.Delete(collection)
	})
}
