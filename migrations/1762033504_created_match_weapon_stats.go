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
					"cascadeDelete": false,
					"collectionId": "pbc_2541054544",
					"hidden": false,
					"id": "relation2052834565",
					"maxSelect": 1,
					"minSelect": 0,
					"name": "match",
					"presentable": false,
					"required": false,
					"system": false,
					"type": "relation"
				},
				{
					"cascadeDelete": false,
					"collectionId": "pbc_2936669995",
					"hidden": false,
					"id": "relation2551806565",
					"maxSelect": 1,
					"minSelect": 0,
					"name": "player",
					"presentable": false,
					"required": false,
					"system": false,
					"type": "relation"
				},
				{
					"autogeneratePattern": "",
					"hidden": false,
					"id": "text3862730489",
					"max": 0,
					"min": 0,
					"name": "weapon_name",
					"pattern": "",
					"presentable": false,
					"primaryKey": false,
					"required": false,
					"system": false,
					"type": "text"
				},
				{
					"autogeneratePattern": "",
					"hidden": false,
					"id": "text1812504113",
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
					"id": "number795295649",
					"max": null,
					"min": null,
					"name": "kills",
					"onlyInt": false,
					"presentable": false,
					"required": false,
					"system": false,
					"type": "number"
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
			"id": "pbc_626477742",
			"indexes": [
				"CREATE INDEX ` + "`" + `idx_match_weapon_stats_match` + "`" + ` ON ` + "`" + `match_weapon_stats` + "`" + ` (` + "`" + `match` + "`" + `)",
				"CREATE INDEX ` + "`" + `idx_match_weapon_stats_player` + "`" + ` ON ` + "`" + `match_weapon_stats` + "`" + ` (` + "`" + `player` + "`" + `)",
				"CREATE INDEX ` + "`" + `idx_match_weapon_stats_weapon` + "`" + ` ON ` + "`" + `match_weapon_stats` + "`" + ` (` + "`" + `weapon_name` + "`" + `)",
				"CREATE INDEX ` + "`" + `idx_match_weapon_stats_player_weapon` + "`" + ` ON ` + "`" + `match_weapon_stats` + "`" + ` (\n  ` + "`" + `player` + "`" + `,\n  ` + "`" + `weapon_name` + "`" + `\n)"
			],
			"listRule": "",
			"name": "match_weapon_stats",
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
		collection, err := app.FindCollectionByNameOrId("pbc_626477742")
		if err != nil {
			return err
		}

		return app.Delete(collection)
	})
}
