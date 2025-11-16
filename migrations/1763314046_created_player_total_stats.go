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
					"cascadeDelete": false,
					"collectionId": "pbc_2936669995",
					"hidden": false,
					"id": "_clone_8Dpm",
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
					"id": "json3960842643",
					"maxSize": 1,
					"name": "total_kills",
					"presentable": false,
					"required": false,
					"system": false,
					"type": "json"
				},
				{
					"hidden": false,
					"id": "json254518791",
					"maxSize": 1,
					"name": "total_deaths",
					"presentable": false,
					"required": false,
					"system": false,
					"type": "json"
				},
				{
					"hidden": false,
					"id": "json4058751331",
					"maxSize": 1,
					"name": "total_score",
					"presentable": false,
					"required": false,
					"system": false,
					"type": "json"
				},
				{
					"hidden": false,
					"id": "json4238285754",
					"maxSize": 1,
					"name": "total_duration_seconds",
					"presentable": false,
					"required": false,
					"system": false,
					"type": "json"
				},
				{
					"hidden": false,
					"id": "json3853725401",
					"maxSize": 1,
					"name": "total_assists",
					"presentable": false,
					"required": false,
					"system": false,
					"type": "json"
				},
				{
					"hidden": false,
					"id": "json832729509",
					"maxSize": 1,
					"name": "total_ff_kills",
					"presentable": false,
					"required": false,
					"system": false,
					"type": "json"
				},
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
				}
			],
			"id": "pbc_1972907995",
			"indexes": [],
			"listRule": null,
			"name": "player_total_stats",
			"system": false,
			"type": "view",
			"updateRule": null,
			"viewQuery": "SELECT \n  player as id,\n  player,\n  COALESCE(SUM(kills), 0) as total_kills,\n  COALESCE(SUM(deaths), 0) as total_deaths,\n  COALESCE(SUM(score), 0) as total_score,\n  COALESCE(SUM(total_play_time), 0) as total_duration_seconds,\n  COALESCE(SUM(assists), 0) as total_assists,\n  COALESCE(SUM(friendly_fire_kills), 0) as total_ff_kills\nFROM match_player_stats\nGROUP BY player;",
			"viewRule": null
		}`

		collection := &core.Collection{}
		if err := json.Unmarshal([]byte(jsonData), &collection); err != nil {
			return err
		}

		return app.Save(collection)
	}, func(app core.App) error {
		collection, err := app.FindCollectionByNameOrId("pbc_1972907995")
		if err != nil {
			return err
		}

		return app.Delete(collection)
	})
}
