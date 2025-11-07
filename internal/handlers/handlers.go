package handlers

import (
	"fmt"
	"net/http"
	"strings"

	"sandstorm-tracker/assets"

	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tools/template"
)

// AppInterface defines the methods handlers need from the app
type AppInterface interface {
	core.App
	// Add any custom app methods here as needed
	// SendRconCommand(serverID string, command string) (string, error)
}

// Register registers all HTTP routes for the web UI
func Register(app AppInterface) {
	registry := template.NewRegistry()

	// OnServe hook to register routes
	app.OnServe().BindFunc(func(e *core.ServeEvent) error {
		// Servers page
		e.Router.GET("/", func(re *core.RequestEvent) error {
			servers, err := re.App.FindAllRecords("servers")
			if err != nil {
				servers = []*core.Record{} // Empty if error
			}

			html, err := registry.LoadFS(assets.GetWebAssets().FS(),
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

		// Server matches endpoint
		e.Router.GET("/servers/{id}/matches", func(re *core.RequestEvent) error {
			serverID := re.Request.PathValue("id")

			// Get server info
			server, err := re.App.FindRecordById("servers", serverID)
			if err != nil {
				return re.NotFoundError("Server not found", err)
			}

			// Get matches for this server, ordered by start_time DESC (active first, then most recent)
			matches, err := re.App.FindRecordsByFilter(
				"matches",
				"server = {:serverId}",
				"-end_time, -start_time", // NULL end_time (active) sorts first, then by start_time DESC
				-1,
				0,
				map[string]any{"serverId": serverID},
			)
			if err != nil {
				matches = []*core.Record{}
			}

			// Format match data
			type MatchInfo struct {
				Map       string
				Mode      string
				Scenario  string
				StartTime string
				EndTime   string
			}

			matchInfos := make([]MatchInfo, len(matches))
			for i, match := range matches {
				endTime := ""
				if !match.GetDateTime("end_time").IsZero() {
					endTime = match.GetDateTime("end_time").Time().Format("2006-01-02 15:04")
				}

				matchInfos[i] = MatchInfo{
					Map:       match.GetString("map"),
					Mode:      match.GetString("mode"),
					Scenario:  match.GetString("scenario"),
					StartTime: match.GetDateTime("start_time").Time().Format("2006-01-02 15:04"),
					EndTime:   endTime,
				}
			}

			html, err := registry.LoadFS(assets.GetWebAssets().FS(),
				"templates/server_matches.html",
			).Render(map[string]any{
				"ServerName": server.GetString("external_id"),
				"Matches":    matchInfos,
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

			// Format match data
			type MatchInfo struct {
				ServerName string
				Map        string
				Mode       string
				StartTime  string
				EndTime    string
				IsActive   bool
			}

			matchInfos := make([]MatchInfo, len(matches))
			for i, match := range matches {
				serverName := ""
				if serverRec := match.ExpandedOne("server"); serverRec != nil {
					serverName = serverRec.GetString("external_id")
				}

				endTime := ""
				isActive := match.GetDateTime("end_time").IsZero()
				if !isActive {
					endTime = match.GetDateTime("end_time").Time().Format("2006-01-02 15:04")
				}

				matchInfos[i] = MatchInfo{
					ServerName: serverName,
					Map:        match.GetString("map"),
					Mode:       match.GetString("mode"),
					StartTime:  match.GetDateTime("start_time").Time().Format("2006-01-02 15:04"),
					EndTime:    endTime,
					IsActive:   isActive,
				}
			}

			html, err := registry.LoadFS(assets.GetWebAssets().FS(),
				"templates/layout.html",
				"templates/matches.html",
			).Render(map[string]any{
				"ActivePage": "matches",
				"Matches":    matchInfos,
			})

			if err != nil {
				return re.InternalServerError("Failed to render template", err)
			}

			return re.HTML(http.StatusOK, html)
		})

		// Players page
		e.Router.GET("/players", func(re *core.RequestEvent) error {
			searchQuery := re.Request.URL.Query().Get("search")

			// Build filter for search (only by name)
			filter := ""
			if searchQuery != "" {
				filter = "name ~ {:search}"
			}

			var players []*core.Record
			var err error
			if filter != "" {
				players, err = re.App.FindRecordsByFilter(
					"players",
					filter,
					"",
					-1,
					0,
					map[string]any{"search": searchQuery},
				)
			} else {
				players, err = re.App.FindAllRecords("players")
			}

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

			// Check if this is an HTMX request (partial update)
			isHTMX := re.Request.Header.Get("HX-Request") == "true"

			var html string
			if isHTMX {
				// Return just the table for HTMX updates
				html, err = registry.LoadFS(assets.GetWebAssets().FS(),
					"templates/players_table.html",
				).Render(map[string]any{
					"Players": playerStats,
				})
			} else {
				// Return full page
				html, err = registry.LoadFS(assets.GetWebAssets().FS(),
					"templates/layout.html",
					"templates/players.html",
				).Render(map[string]any{
					"ActivePage": "players",
					"Players":    playerStats,
				})
			}

			if err != nil {
				return re.InternalServerError("Failed to render template", err)
			}

			return re.HTML(http.StatusOK, html)
		})

		// Weapons page
		e.Router.GET("/weapons", func(re *core.RequestEvent) error {
			searchQuery := re.Request.URL.Query().Get("search")

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

				// Apply search filter
				if searchQuery != "" {
					if !contains(weapon, searchQuery) {
						continue
					}
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

			// Check if this is an HTMX request (partial update)
			isHTMX := re.Request.Header.Get("HX-Request") == "true"

			var html string
			if isHTMX {
				// Return just the table for HTMX updates
				html, err = registry.LoadFS(assets.GetWebAssets().FS(),
					"templates/weapons_table.html",
				).Render(map[string]any{
					"Weapons": weapons,
				})
			} else {
				// Return full page
				html, err = registry.LoadFS(assets.GetWebAssets().FS(),
					"templates/layout.html",
					"templates/weapons.html",
				).Render(map[string]any{
					"ActivePage": "weapons",
					"Weapons":    weapons,
				})
			}

			if err != nil {
				return re.InternalServerError("Failed to render template", err)
			}

			return re.HTML(http.StatusOK, html)
		})

		// Health check endpoint
		e.Router.GET("/health", func(re *core.RequestEvent) error {
			health := map[string]any{
				"status": "ok",
				"database": map[string]any{
					"connected": true,
				},
			}

			// Try to get RCON pool info if app has the method
			type rconPoolGetter interface {
				GetRconPoolStatus() map[string]any
			}

			if customApp, ok := app.(rconPoolGetter); ok {
				health["rcon"] = customApp.GetRconPoolStatus()
			}

			// Try to get A2S pool info if app has the method
			type a2sPoolGetter interface {
				GetA2SPool() interface{ ListServers() []string }
			}

			if customApp, ok := app.(a2sPoolGetter); ok {
				pool := customApp.GetA2SPool()
				if pool != nil {
					servers := pool.ListServers()
					health["a2s"] = map[string]any{
						"available":     true,
						"total_servers": len(servers),
						"servers":       servers,
					}
				}
			}

			return re.JSON(http.StatusOK, health)
		})

		return e.Next()
	})
}

// contains performs a case-insensitive substring search
func contains(s, substr string) bool {
	s = strings.ToLower(s)
	substr = strings.ToLower(substr)
	return strings.Contains(s, substr)
}
