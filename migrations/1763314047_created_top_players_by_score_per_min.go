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
					"id": "text0987654321",
					"max": 0,
					"min": 0,
					"name": "name",
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
					"name": "scorePerMin",
					"onlyInt": false,
					"presentable": false,
					"required": false,
					"system": false,
					"type": "number"
				},
				{
					"hidden": false,
					"id": "number1111111111",
					"max": null,
					"min": null,
					"name": "totalScore",
					"onlyInt": false,
					"presentable": false,
					"required": false,
					"system": false,
					"type": "number"
				},
				{
					"hidden": false,
					"id": "number2222222222",
					"max": null,
					"min": null,
					"name": "totalDurationSeconds",
					"onlyInt": false,
					"presentable": false,
					"required": false,
					"system": false,
					"type": "number"
				}
			],
			"id": "pbc_1972907996",
			"indexes": [],
			"listRule": null,
			"name": "top_players_by_score_per_min",
			"system": false,
			"type": "view",
			"updateRule": null,
			"viewQuery": "SELECT p.id, p.id as player, p.name, (CAST(stats.total_score AS REAL) / (CAST(stats.total_duration_seconds AS REAL) / 60.0)) as scorePerMin, stats.total_score as totalScore, stats.total_duration_seconds as totalDurationSeconds FROM players p INNER JOIN player_total_stats stats ON p.id = stats.id WHERE stats.total_duration_seconds >= 60 ORDER BY scorePerMin DESC;",
			"viewRule": null
		}`

		collection := &core.Collection{}
		if err := json.Unmarshal([]byte(jsonData), &collection); err != nil {
			return err
		}

		return app.Save(collection)
	}, func(app core.App) error {
		collection, err := app.FindCollectionByNameOrId("pbc_1972907996")
		if err != nil {
			return err
		}

		return app.Delete(collection)
	})
}
