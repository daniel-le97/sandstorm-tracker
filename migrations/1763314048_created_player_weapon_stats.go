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
					"autogeneratePattern": "",
					"hidden": false,
					"id": "text3208210256",
					"max": 0,
					"min": 0,
					"name": "id",
					"pattern": "^[a-z0-9]+$",
					"presentable": false,
					"primaryKey": true,
					"required": true,
					"system": true,
					"type": "text"
				},
				{
					"cascadeDelete": false,
					"collectionId": "pbc_2936669995",
					"hidden": false,
					"id": "_clone_8Dpm",
					"maxSelect": 1,
					"minSelect": 0,
					"name": "player",
					"presentable": false,
					"required": false,
					"system": false,
					"type": "relation"
				},
				{
					"hidden": false,
					"id": "text1234567890",
					"max": 0,
					"min": 0,
					"name": "weapon_name",
					"pattern": "",
					"presentable": true,
					"required": false,
					"system": false,
					"type": "text"
				},
				{
					"hidden": false,
					"id": "number9876543210",
					"max": null,
					"min": null,
					"name": "total_kills",
					"onlyInt": false,
					"presentable": false,
					"required": false,
					"system": false,
					"type": "number"
				}
			],
			"id": "pbc_1972907997",
			"indexes": [],
			"listRule": "",
			"name": "player_weapon_stats",
			"system": false,
			"type": "view",
			"updateRule": null,
			"viewQuery": "SELECT \n  (player || '_' || weapon_name) as id,\n  player,\n  weapon_name,\n  SUM(kills) as total_kills\nFROM match_weapon_stats\nGROUP BY player, weapon_name;",
			"viewRule": ""
		}`

		collection := &core.Collection{}
		if err := json.Unmarshal([]byte(jsonData), &collection); err != nil {
			return err
		}

		return app.Save(collection)
	}, func(app core.App) error {
		collection, err := app.FindCollectionByNameOrId("pbc_1972907997")
		if err != nil {
			return err
		}

		return app.Delete(collection)
	})
}
