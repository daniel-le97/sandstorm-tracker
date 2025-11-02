package app

import (
	"fmt"
	"net/http"

	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tools/template"
)

// RegisterWebRoutes registers the HTML frontend routes using template rendering
func RegisterWebRoutes(app core.App) {
	registry := template.NewRegistry()

	// OnServe hook to register routes
	app.OnServe().BindFunc(func(e *core.ServeEvent) error {
		// Servers page
		e.Router.GET("/", func(re *core.RequestEvent) error {
			servers, err := re.App.FindAllRecords("servers")
			if err != nil {
				servers = []*core.Record{} // Empty if error
			}

			html, err := registry.LoadFiles(
				"templates/layout.html",
				"templates/servers.html",
			).Render(map[string]any{
				"ActivePage": "servers",
				"Servers":    servers,
			})

			if err != nil {
				return re.InternalServerError("Failed to render template", err)
			}

			return re.HTML(http.StatusOK, html)
		})

		// Matches page
		e.Router.GET("/matches", func(re *core.RequestEvent) error {
			matches, err := re.App.FindRecordsByFilter(
				"matches",
				"",
				"-start_time",
				50,
				0,
			)
			if err != nil {
				matches = []*core.Record{}
			}

			// Expand server relation
			re.App.ExpandRecords(matches, []string{"server"}, nil)

			html, err := registry.LoadFiles(
				"templates/layout.html",
				"templates/matches.html",
			).Render(map[string]any{
				"ActivePage": "matches",
				"Matches":    matches,
			})

			if err != nil {
				return re.InternalServerError("Failed to render template", err)
			}

			return re.HTML(http.StatusOK, html)
		})

		// Players page
		e.Router.GET("/players", func(re *core.RequestEvent) error {
			players, err := re.App.FindAllRecords("players")
			if err != nil {
				players = []*core.Record{}
			}

			// Calculate stats for each player
			type PlayerStats struct {
				Name        string
				ExternalID  string
				TotalKills  int
				TotalDeaths int
				KDRatio     string
				Created     string
			}

			playerStats := make([]PlayerStats, len(players))
			for i, player := range players {
				// Get total kills from match_weapon_stats
				kills := 0
				weaponStats, err := re.App.FindRecordsByFilter(
					"match_weapon_stats",
					"player = {:playerId}",
					"",
					-1,
					0,
					map[string]any{"playerId": player.Id},
				)
				if err == nil {
					for _, stat := range weaponStats {
						kills += stat.GetInt("kills")
					}
				}

				// Get total deaths from match_player_stats
				deaths := 0
				playerMatchStats, err := re.App.FindRecordsByFilter(
					"match_player_stats",
					"player = {:playerId}",
					"",
					-1,
					0,
					map[string]any{"playerId": player.Id},
				)
				if err == nil {
					for _, stat := range playerMatchStats {
						deaths += stat.GetInt("deaths")
					}
				}

				// Calculate K/D ratio
				kdRatio := "0.00"
				if deaths > 0 {
					kdRatio = fmt.Sprintf("%.2f", float64(kills)/float64(deaths))
				} else if kills > 0 {
					kdRatio = "∞"
				}

				playerStats[i] = PlayerStats{
					Name:        player.GetString("name"),
					ExternalID:  player.GetString("external_id"),
					TotalKills:  kills,
					TotalDeaths: deaths,
					KDRatio:     kdRatio,
					Created:     player.GetDateTime("created").Time().Format("2006-01-02 15:04"),
				}
			}

			html, err := registry.LoadFiles(
				"templates/layout.html",
				"templates/players.html",
			).Render(map[string]any{
				"ActivePage": "players",
				"Players":    playerStats,
			})

			if err != nil {
				return re.InternalServerError("Failed to render template", err)
			}

			return re.HTML(http.StatusOK, html)
		})

		// Weapons page
		e.Router.GET("/weapons", func(re *core.RequestEvent) error {
			weaponStats, err := re.App.FindAllRecords("match_weapon_stats")
			if err != nil {
				weaponStats = []*core.Record{}
			}

			// Aggregate weapon statistics
			type WeaponStat struct {
				Weapon     string
				TotalKills int
				TimesUsed  int
				AvgKills   float64
			}

			weaponMap := make(map[string]*WeaponStat)
			for _, stat := range weaponStats {
				weapon := stat.GetString("weapon")
				kills := stat.GetInt("kills")

				if weapon == "" {
					continue
				}

				if _, exists := weaponMap[weapon]; !exists {
					weaponMap[weapon] = &WeaponStat{
						Weapon: weapon,
					}
				}

				weaponMap[weapon].TotalKills += kills
				weaponMap[weapon].TimesUsed++
			}

			// Convert map to slice and calculate averages
			weapons := make([]WeaponStat, 0, len(weaponMap))
			for _, ws := range weaponMap {
				if ws.TimesUsed > 0 {
					ws.AvgKills = float64(ws.TotalKills) / float64(ws.TimesUsed)
				}
				weapons = append(weapons, *ws)
			}

			// Sort by total kills (descending) - simple bubble sort
			for i := 0; i < len(weapons); i++ {
				for j := i + 1; j < len(weapons); j++ {
					if weapons[j].TotalKills > weapons[i].TotalKills {
						weapons[i], weapons[j] = weapons[j], weapons[i]
					}
				}
			}

			html, err := registry.LoadFiles(
				"templates/layout.html",
				"templates/weapons.html",
			).Render(map[string]any{
				"ActivePage": "weapons",
				"Weapons":    weapons,
			})

			if err != nil {
				return re.InternalServerError("Failed to render template", err)
			}

			return re.HTML(http.StatusOK, html)
		})

		return e.Next()
	})
}
