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
					"cascadeDelete": true,
					"collectionId": "pbc_2541054544",
					"hidden": false,
					"id": "relation_match",
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
					"id": "relation_killer",
					"maxSelect": 1,
					"minSelect": 0,
					"name": "killer",
					"presentable": false,
					"required": true,
					"system": false,
					"type": "relation"
				},
				{
					"cascadeDelete": false,
					"collectionId": "pbc_2936669995",
					"hidden": false,
					"id": "relation_victim",
					"maxSelect": 1,
					"minSelect": 0,
					"name": "victim",
					"presentable": false,
					"required": true,
					"system": false,
					"type": "relation"
				},
				{
					"hidden": false,
					"id": "text_weapon",
					"max": 100,
					"min": 0,
					"name": "weapon",
					"pattern": "",
					"presentable": false,
					"required": true,
					"system": false,
					"type": "text"
				},
				{
					"hidden": false,
					"id": "datetime_timestamp",
					"max": "",
					"min": "",
					"name": "timestamp",
					"presentable": false,
					"required": true,
					"system": false,
					"type": "date"
				},
				{
					"hidden": false,
					"id": "number_killer_team",
					"max": null,
					"min": null,
					"name": "killer_team",
					"onlyInt": true,
					"presentable": false,
					"required": false,
					"system": false,
					"type": "number"
				},
				{
					"hidden": false,
					"id": "number_victim_team",
					"max": null,
					"min": null,
					"name": "victim_team",
					"onlyInt": true,
					"presentable": false,
					"required": false,
					"system": false,
					"type": "number"
				},
				{
					"hidden": false,
					"id": "number_time_since_match_start",
					"max": null,
					"min": 0,
					"name": "time_since_match_start_seconds",
					"onlyInt": false,
					"presentable": false,
					"required": false,
					"system": false,
					"type": "number"
				},
				{
					"hidden": false,
					"id": "number_time_since_last_ff",
					"max": null,
					"min": 0,
					"name": "time_since_last_ff_seconds",
					"presentable": false,
					"required": false,
					"system": false,
					"type": "number"
				},
				{
					"hidden": false,
					"id": "number_killer_total_kills_in_match",
					"max": null,
					"min": 0,
					"name": "killer_total_kills_in_match",
					"onlyInt": true,
					"presentable": false,
					"required": false,
					"system": false,
					"type": "number"
				},
				{
					"hidden": false,
					"id": "number_killer_ff_count_in_match",
					"max": null,
					"min": 0,
					"name": "killer_ff_count_in_match",
					"onlyInt": true,
					"presentable": false,
					"required": false,
					"system": false,
					"type": "number"
				},
				{
					"hidden": false,
					"id": "bool_is_explosive",
					"name": "is_explosive_weapon",
					"presentable": false,
					"required": false,
					"system": false,
					"type": "bool"
				},
				{
					"hidden": false,
					"id": "bool_is_vehicle",
					"name": "is_vehicle_weapon",
					"presentable": false,
					"required": false,
					"system": false,
					"type": "bool"
				},
				{
					"hidden": false,
					"id": "text_map",
					"max": 100,
					"min": 0,
					"name": "map",
					"pattern": "",
					"presentable": false,
					"required": false,
					"system": false,
					"type": "text"
				},
				{
					"hidden": false,
					"id": "text_game_mode",
					"max": 50,
					"min": 0,
					"name": "game_mode",
					"pattern": "",
					"presentable": false,
					"required": false,
					"system": false,
					"type": "text"
				},
				{
					"hidden": false,
					"id": "select_classification",
					"maxSelect": 1,
					"name": "accident_classification",
					"presentable": false,
					"required": false,
					"system": false,
					"type": "select",
					"values": [
						"likely_accident",
						"possibly_intentional",
						"likely_intentional",
						"unclassified"
					]
				},
				{
					"hidden": false,
					"id": "number_confidence_score",
					"max": 1.0,
					"min": 0.0,
					"name": "confidence_score",
					"presentable": false,
					"required": false,
					"system": false,
					"type": "number"
				},
				{
					"autogeneratePattern": "",
					"hidden": false,
					"id": "autodate2990389176",
					"name": "created",
					"onCreate": true,
					"onUpdate": false,
					"presentable": false,
					"system": true,
					"type": "autodate"
				},
				{
					"autogeneratePattern": "",
					"hidden": false,
					"id": "autodate3332085495",
					"name": "updated",
					"onCreate": true,
					"onUpdate": true,
					"presentable": false,
					"system": true,
					"type": "autodate"
				}
			],
			"id": "pbc_friendly_fire_incidents",
			"indexes": [
				"CREATE INDEX IF NOT EXISTS ` + "`" + `idx_ff_match` + "`" + ` ON ` + "`" + `friendly_fire_incidents` + "`" + ` (` + "`" + `match` + "`" + `)",
				"CREATE INDEX IF NOT EXISTS ` + "`" + `idx_ff_killer` + "`" + ` ON ` + "`" + `friendly_fire_incidents` + "`" + ` (` + "`" + `killer` + "`" + `)",
				"CREATE INDEX IF NOT EXISTS ` + "`" + `idx_ff_timestamp` + "`" + ` ON ` + "`" + `friendly_fire_incidents` + "`" + ` (` + "`" + `timestamp` + "`" + `)",
				"CREATE INDEX IF NOT EXISTS ` + "`" + `idx_ff_classification` + "`" + ` ON ` + "`" + `friendly_fire_incidents` + "`" + ` (` + "`" + `accident_classification` + "`" + `)"
			],
			"listRule": null,
			"name": "friendly_fire_incidents",
			"system": false,
			"type": "base",
			"updateRule": null,
			"viewRule": null
		}`

		collection := &core.Collection{}
		if err := json.Unmarshal([]byte(jsonData), collection); err != nil {
			return err
		}

		return app.Save(collection)
	}, func(app core.App) error {
		collection, err := app.FindCollectionByNameOrId("pbc_friendly_fire_incidents")
		if err != nil {
			return err
		}

		return app.Delete(collection)
	})
}
