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
					"required": true,
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
					"required": true,
					"system": false,
					"type": "relation"
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
					"id": "number296718251",
					"max": null,
					"min": null,
					"name": "assists",
					"onlyInt": false,
					"presentable": false,
					"required": false,
					"system": false,
					"type": "number"
				},
				{
					"hidden": false,
					"id": "number3342828817",
					"max": null,
					"min": null,
					"name": "deaths",
					"onlyInt": false,
					"presentable": false,
					"required": false,
					"system": false,
					"type": "number"
				},
				{
					"hidden": false,
					"id": "number1109413046",
					"max": null,
					"min": null,
					"name": "friendly_fire_kills",
					"onlyInt": false,
					"presentable": false,
					"required": false,
					"system": false,
					"type": "number"
				},
				{
					"hidden": false,
					"id": "number848901969",
					"max": null,
					"min": null,
					"name": "score",
					"onlyInt": false,
					"presentable": false,
					"required": false,
					"system": false,
					"type": "number"
				},
				{
					"hidden": false,
					"id": "number1279700572",
					"max": null,
					"min": null,
					"name": "objectives_captured",
					"onlyInt": false,
					"presentable": false,
					"required": false,
					"system": false,
					"type": "number"
				},
				{
					"hidden": false,
					"id": "number2494212961",
					"max": null,
					"min": null,
					"name": "objectvies_destroyed",
					"onlyInt": false,
					"presentable": false,
					"required": false,
					"system": false,
					"type": "number"
				},
				{
					"hidden": false,
					"id": "number3628669570",
					"max": null,
					"min": null,
					"name": "total_play_time",
					"onlyInt": false,
					"presentable": false,
					"required": false,
					"system": false,
					"type": "number"
				},
				{
					"hidden": false,
					"id": "number1587870900",
					"max": null,
					"min": null,
					"name": "session_count",
					"onlyInt": false,
					"presentable": false,
					"required": false,
					"system": false,
					"type": "number"
				},
				{
					"hidden": false,
					"id": "date1604994294",
					"max": "",
					"min": "",
					"name": "first_joined_at",
					"presentable": false,
					"required": false,
					"system": false,
					"type": "date"
				},
				{
					"hidden": false,
					"id": "date1368287270",
					"max": "",
					"min": "",
					"name": "last_left_at",
					"presentable": false,
					"required": false,
					"system": false,
					"type": "date"
				},
				{
					"hidden": false,
					"id": "bool3821226774",
					"name": "is_currently_connected",
					"presentable": false,
					"required": false,
					"system": false,
					"type": "bool"
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
			"id": "pbc_3080700301",
			"indexes": [
				"CREATE INDEX ` + "`" + `idx_match_player_stats_match` + "`" + ` ON ` + "`" + `match_player_stats` + "`" + ` (` + "`" + `match` + "`" + `)",
				"CREATE INDEX ` + "`" + `idx_match_player_stats_player` + "`" + ` ON ` + "`" + `match_player_stats` + "`" + ` (` + "`" + `player` + "`" + `)",
				"CREATE INDEX ` + "`" + `idx_match_player_stats_connected` + "`" + ` ON ` + "`" + `match_player_stats` + "`" + ` (` + "`" + `is_currently_connected` + "`" + `)",
				"CREATE INDEX ` + "`" + `idx_match_player_stats_match_connected` + "`" + ` ON ` + "`" + `match_player_stats` + "`" + ` (\n  ` + "`" + `match` + "`" + `,\n  ` + "`" + `is_currently_connected` + "`" + `\n)"
			],
			"listRule": null,
			"name": "match_player_stats",
			"system": false,
			"type": "base",
			"updateRule": null,
			"viewRule": null
		}`

		collection := &core.Collection{}
		if err := json.Unmarshal([]byte(jsonData), &collection); err != nil {
			return err
		}

		return app.Save(collection)
	}, func(app core.App) error {
		collection, err := app.FindCollectionByNameOrId("pbc_3080700301")
		if err != nil {
			return err
		}

		return app.Delete(collection)
	})
}

