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
					"id": "text2477632187",
					"max": 0,
					"min": 0,
					"name": "map",
					"pattern": "",
					"presentable": false,
					"primaryKey": false,
					"required": false,
					"system": false,
					"type": "text"
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
					"id": "select2063623452",
					"maxSelect": 1,
					"name": "status",
					"presentable": false,
					"required": false,
					"system": false,
					"type": "select",
					"values": [
						"ongoing",
						"finished",
						"crashed"
					]
				},
				{
					"hidden": false,
					"id": "number3320769076",
					"max": null,
					"min": null,
					"name": "round",
					"onlyInt": false,
					"presentable": false,
					"required": false,
					"system": false,
					"type": "number"
				},
				{
					"hidden": false,
					"id": "number846593620",
					"max": null,
					"min": null,
					"name": "round_objective",
					"onlyInt": false,
					"presentable": false,
					"required": false,
					"system": false,
					"type": "number"
				},
				{
					"hidden": false,
					"id": "number3137288",
					"max": null,
					"min": null,
					"name": "winner_team",
					"onlyInt": false,
					"presentable": false,
					"required": false,
					"system": false,
					"type": "number"
				},
				{
					"hidden": false,
					"id": "date1345189255",
					"max": "",
					"min": "",
					"name": "start_time",
					"presentable": false,
					"required": false,
					"system": false,
					"type": "date"
				},
				{
					"hidden": false,
					"id": "date1096160257",
					"max": "",
					"min": "",
					"name": "end_time",
					"presentable": false,
					"required": false,
					"system": false,
					"type": "date"
				},
				{
					"autogeneratePattern": "",
					"hidden": false,
					"id": "text2546616235",
					"max": 0,
					"min": 0,
					"name": "mode",
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
					"id": "text1727723052",
					"max": 0,
					"min": 0,
					"name": "player_team",
					"pattern": "",
					"presentable": false,
					"primaryKey": false,
					"required": false,
					"system": false,
					"type": "text"
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
			"id": "pbc_2541054544",
			"indexes": [
				"CREATE INDEX IF NOT EXISTS ` + "`" + `idx_matches_server` + "`" + ` ON ` + "`" + `matches` + "`" + ` (` + "`" + `server` + "`" + `)",
				"CREATE INDEX IF NOT EXISTS ` + "`" + `idx_matches_start_time` + "`" + ` ON ` + "`" + `matches` + "`" + ` (` + "`" + `start_time` + "`" + `)",
				"CREATE INDEX IF NOT EXISTS ` + "`" + `idx_matches_end_time` + "`" + ` ON ` + "`" + `matches` + "`" + ` (` + "`" + `end_time` + "`" + `)",
				"CREATE INDEX IF NOT EXISTS ` + "`" + `idx_matches_server_end_time` + "`" + ` ON ` + "`" + `matches` + "`" + ` (` + "`" + `server` + "`" + `, ` + "`" + `end_time` + "`" + `)",
				"CREATE INDEX IF NOT EXISTS ` + "`" + `idx_matches_server_start_time` + "`" + ` ON ` + "`" + `matches` + "`" + ` (` + "`" + `server` + "`" + `, ` + "`" + `start_time` + "`" + ` DESC)"
			],
			"listRule": "",
			"name": "matches",
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
		collection, err := app.FindCollectionByNameOrId("pbc_2541054544")
		if err != nil {
			return err
		}

		return app.Delete(collection)
	})
}
