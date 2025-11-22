package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"sandstorm-tracker/assets"

	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tools/template"
)

// AppInterface defines the methods handlers need from the app
type AppInterface interface {
	core.App
	// Add any custom app methods here as needed
	SendRconCommand(serverID string, command string) (string, error)
}

// Register registers all HTTP routes for the web UI
func Register(app AppInterface, e *core.ServeEvent) {
	registry := template.NewRegistry()

	// Serve static files (PocketBase JS SDK, etc.) using PocketBase's apis.Static helper
	e.Router.GET("/static/{path...}", apis.Static(assets.StaticFS(), false))

	// Live Server Status page (homepage)
	e.Router.GET("/", func(re *core.RequestEvent) error {
		servers, err := re.App.FindAllRecords("servers")
		if err != nil {
			servers = []*core.Record{}
		}

		// Build server status info
		type PlayerInfo struct {
			Name string
		}

		type ServerStatus struct {
			ServerID           string
			ServerName         string
			Map                string
			Mode               string
			Round              int
			RoundObjective     int
			NumObjectives      int
			ObjectivePercent   int    // Calculated percentage for progress bar
			CurrentObjective   string // Current objective letter (A, B, C, etc.)
			TotalObjectivesStr string // Total objectives as letters (e.g., "A-F" for 6 objectives)
			CurrentPlayers     []PlayerInfo
			PlayerCount        int
			IsActive           bool
		}

		serverStatuses := make([]ServerStatus, 0, len(servers))
		for _, server := range servers {
			// Get active match for this server
			matches, err := re.App.FindRecordsByFilter(
				"matches",
				"server = {:serverId} && end_time = ''",
				"-start_time",
				1,
				0,
				map[string]any{"serverId": server.Id},
			)

			status := ServerStatus{
				ServerID:       server.Id,
				ServerName:     server.GetString("name"),
				IsActive:       false,
				CurrentPlayers: []PlayerInfo{},
			}

			if err == nil && len(matches) > 0 {
				match := matches[0]
				status.IsActive = true
				status.Map = match.GetString("map")
				status.Mode = match.GetString("mode")
				status.Round = match.GetInt("round")
				status.RoundObjective = match.GetInt("round_objective")
				status.NumObjectives = match.GetInt("num_objectives")

				// Calculate objective percentage
				if status.NumObjectives > 0 {
					status.ObjectivePercent = (status.RoundObjective * 100) / status.NumObjectives
				}

				// Convert objective numbers to letters
				// 0 = A, 1 = B, 2 = C, etc.
				status.CurrentObjective = string(rune('A' + status.RoundObjective))

				// Total objectives as letter range (e.g., "A-F" for 6 objectives)
				if status.NumObjectives > 0 {
					lastLetter := string(rune('A' + status.NumObjectives - 1))
					if status.NumObjectives == 1 {
						status.TotalObjectivesStr = "A"
					} else {
						status.TotalObjectivesStr = "A-" + lastLetter
					}
				}

				// Get currently connected players for this match
				playerStats, err := re.App.FindRecordsByFilter(
					"match_player_stats",
					"match = {:matchId} && is_currently_connected = true",
					"",
					-1,
					0,
					map[string]any{"matchId": match.Id},
				)

				if err == nil {
					// Expand player records to get names
					re.App.ExpandRecords(playerStats, []string{"player"}, nil)

					for _, stat := range playerStats {
						if playerRec := stat.ExpandedOne("player"); playerRec != nil {
							status.CurrentPlayers = append(status.CurrentPlayers, PlayerInfo{
								Name: playerRec.GetString("name"),
							})
						}
					}
					status.PlayerCount = len(status.CurrentPlayers)
				}
			}

			serverStatuses = append(serverStatuses, status)
		}

		html, err := registry.LoadFS(assets.GetWebAssets().FS(),
			"templates/layout.html",
			"templates/server_status.html",
		).Render(map[string]any{
			"ActivePage": "status",
			"Servers":    serverStatuses,
		})

		if err != nil {
			return re.InternalServerError("Failed to render template", err)
		}

		return re.HTML(http.StatusOK, html)
	})

	// Admin servers page (original functionality)
	e.Router.GET("/admin/servers", func(re *core.RequestEvent) error {
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
			TotalScore  int
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

			// Get total deaths and score from match_player_stats
			deaths := 0
			totalScore := 0
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
					totalScore += stat.GetInt("score")
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
				TotalScore:  totalScore,
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

	// Weapons page - shows each player's top 3 weapons
	e.Router.GET("/weapons", func(re *core.RequestEvent) error {
		searchQuery := re.Request.URL.Query().Get("search")

		// Get all players
		players, err := re.App.FindAllRecords("players")
		if err != nil {
			players = []*core.Record{}
		}

		// Get all weapon stats
		weaponStats, err := re.App.FindAllRecords("match_weapon_stats")
		if err != nil {
			weaponStats = []*core.Record{}
		}

		type PlayerWeapon struct {
			Weapon string
			Kills  int
		}

		type PlayerWeaponData struct {
			PlayerName string
			PlayerID   string
			TopWeapons []PlayerWeapon
		}

		playerWeaponMap := make(map[string]map[string]int)
		playerNameMap := make(map[string]string) // playerID -> playerName

		// Build player name map first
		for _, player := range players {
			playerNameMap[player.Id] = player.GetString("name")
		}

		// Aggregate weapons by player from weapon stats
		for _, stat := range weaponStats {
			weapon := stat.GetString("weapon_name")
			kills := stat.GetInt("kills")
			playerID := stat.GetString("player")

			if weapon == "" || playerID == "" || kills == 0 {
				continue
			}

			// Apply search filter to weapon name or player name
			if searchQuery != "" {
				playerName := playerNameMap[playerID]
				if !contains(weapon, searchQuery) && !contains(playerName, searchQuery) {
					continue
				}
			}

			if _, exists := playerWeaponMap[playerID]; !exists {
				playerWeaponMap[playerID] = make(map[string]int)
			}

			playerWeaponMap[playerID][weapon] += kills
		}

		// Build player weapon data with top 3 weapons
		playerWeapons := make([]PlayerWeaponData, 0)
		for playerID, weapons := range playerWeaponMap {
			playerName := playerNameMap[playerID]
			if playerName == "" {
				continue // Skip if player not found
			}

			// Convert to slice and sort by kills
			weaponSlice := make([]PlayerWeapon, 0, len(weapons))
			for weapon, kills := range weapons {
				weaponSlice = append(weaponSlice, PlayerWeapon{
					Weapon: weapon,
					Kills:  kills,
				})
			}

			// Sort by kills descending
			for i := 0; i < len(weaponSlice); i++ {
				for j := i + 1; j < len(weaponSlice); j++ {
					if weaponSlice[j].Kills > weaponSlice[i].Kills {
						weaponSlice[i], weaponSlice[j] = weaponSlice[j], weaponSlice[i]
					}
				}
			}

			// Take top 3
			topCount := 3
			if len(weaponSlice) < topCount {
				topCount = len(weaponSlice)
			}

			playerWeapons = append(playerWeapons, PlayerWeaponData{
				PlayerName: playerName,
				PlayerID:   playerID,
				TopWeapons: weaponSlice[:topCount],
			})
		}

		// Check if this is an HTMX request (partial update)
		isHTMX := re.Request.Header.Get("HX-Request") == "true"

		var html string
		if isHTMX {
			// Return just the table for HTMX updates
			html, err = registry.LoadFS(assets.GetWebAssets().FS(),
				"templates/weapons_table.html",
			).Render(map[string]any{
				"Players": playerWeapons,
			})
		} else {
			// Return full page
			html, err = registry.LoadFS(assets.GetWebAssets().FS(),
				"templates/layout.html",
				"templates/weapons.html",
			).Render(map[string]any{
				"ActivePage": "weapons",
				"Players":    playerWeapons,
			})
		}

		if err != nil {
			return re.InternalServerError("Failed to render template", err)
		}

		return re.HTML(http.StatusOK, html)
	})

	// Live Match - Current match details with player scores
	e.Router.GET("/live-match/:serverId", func(re *core.RequestEvent) error {
		serverID := re.Request.PathValue("serverId")

		server, err := re.App.FindRecordById("servers", serverID)
		if err != nil {
			return re.NotFoundError("Server not found", err)
		}

		matches, err := re.App.FindRecordsByFilter(
			"matches",
			"server = {:serverId} && end_time = ''",
			"-start_time",
			1,
			0,
			map[string]any{"serverId": serverID},
		)

		type PlayerScore struct {
			PlayerName string
			Team       string
			Kills      int
			Deaths     int
			Assists    int
			KDRatio    string
		}

		data := map[string]any{
			"ActivePage": "live-match",
			"IsActive":   false,
			"ServerName": server.GetString("name"),
		}

		if err == nil && len(matches) > 0 {
			match := matches[0]
			data["IsActive"] = true
			data["Map"] = match.GetString("map")
			data["Mode"] = match.GetString("mode")
			data["Round"] = match.GetInt("round")
			data["StartTime"] = match.GetDateTime("start_time").Time().Format("15:04")
			data["RoundObjective"] = match.GetInt("round_objective")
			data["NumObjectives"] = match.GetInt("num_objectives")

			if numObj := match.GetInt("num_objectives"); numObj > 0 {
				data["ObjectivePercent"] = (match.GetInt("round_objective") * 100) / numObj
				data["CurrentObjective"] = string(rune('A' + match.GetInt("round_objective")))
				lastLetter := string(rune('A' + numObj - 1))
				if numObj == 1 {
					data["TotalObjectivesStr"] = "A"
				} else {
					data["TotalObjectivesStr"] = "A-" + lastLetter
				}
			}

			playerStats, err := re.App.FindRecordsByFilter(
				"match_player_stats",
				"match = {:matchId}",
				"-kills, -deaths",
				-1,
				0,
				map[string]any{"matchId": match.Id},
			)

			if err == nil {
				re.App.ExpandRecords(playerStats, []string{"player"}, nil)

				players := make([]PlayerScore, 0, len(playerStats))
				securityKills, securityDeaths, insurgentKills, insurgentDeaths := 0, 0, 0, 0
				securityCount, insurgentCount := 0, 0

				for _, stat := range playerStats {
					playerRec := stat.ExpandedOne("player")
					if playerRec == nil {
						continue
					}

					kills := stat.GetInt("kills")
					deaths := stat.GetInt("deaths")
					assists := stat.GetInt("assists")

					kdRatio := 0.0
					if deaths > 0 {
						kdRatio = float64(kills) / float64(deaths)
					} else if kills > 0 {
						kdRatio = float64(kills)
					}

					player := PlayerScore{
						PlayerName: playerRec.GetString("name"),
						Kills:      kills,
						Deaths:     deaths,
						Assists:    assists,
						KDRatio:    fmt.Sprintf("%.2f", kdRatio),
					}

					// Determine team (would need to query match_player_stats for team info)
					// For now, alternate or get from stat record
					team := stat.GetString("team")
					if team == "" {
						team = "Unknown"
					}
					player.Team = team

					if team == "Security" {
						securityCount++
						securityKills += kills
						securityDeaths += deaths
					} else if team == "Insurgents" {
						insurgentCount++
						insurgentKills += kills
						insurgentDeaths += deaths
					}

					players = append(players, player)
				}

				data["Players"] = players
				data["SecurityPlayerCount"] = securityCount
				data["SecurityKills"] = securityKills
				data["SecurityDeaths"] = securityDeaths
				data["InsurgentPlayerCount"] = insurgentCount
				data["InsurgentKills"] = insurgentKills
				data["InsurgentDeaths"] = insurgentDeaths
			}
		}

		html, err := registry.LoadFS(assets.GetWebAssets().FS(),
			"templates/layout.html",
			"templates/live-match.html",
		).Render(data)

		if err != nil {
			return re.InternalServerError("Failed to render template", err)
		}

		return re.HTML(http.StatusOK, html)
	})

	// Match History - Historical matches with player stats
	e.Router.GET("/match-history", func(re *core.RequestEvent) error {
		page := 1
		pageSize := 10

		if p := re.Request.URL.Query().Get("page"); p != "" {
			if parsed, err := strconv.Atoi(p); err == nil && parsed > 0 {
				page = parsed
			}
		}

		// Get filter parameters
		selectedServer := re.Request.URL.Query().Get("server")
		selectedMap := re.Request.URL.Query().Get("map")
		selectedMode := re.Request.URL.Query().Get("mode")

		// Build filter
		filters := []string{"end_time != ''"}
		filterParams := make(map[string]any)

		if selectedServer != "" {
			filters = append(filters, "server = {:serverId}")
			filterParams["serverId"] = selectedServer
		}
		if selectedMap != "" {
			filters = append(filters, "map = {:mapName}")
			filterParams["mapName"] = selectedMap
		}
		if selectedMode != "" {
			filters = append(filters, "mode = {:modeName}")
			filterParams["modeName"] = selectedMode
		}

		filterStr := strings.Join(filters, " && ")

		// Get all servers for filter dropdown
		servers, _ := re.App.FindAllRecords("servers")

		// Get unique maps and modes
		allMatches, _ := re.App.FindAllRecords("matches")
		mapSet := make(map[string]bool)
		modeSet := make(map[string]bool)
		for _, m := range allMatches {
			mapSet[m.GetString("map")] = true
			modeSet[m.GetString("mode")] = true
		}

		maps := make([]string, 0, len(mapSet))
		modes := make([]string, 0, len(modeSet))
		for k := range mapSet {
			if k != "" {
				maps = append(maps, k)
			}
		}
		for k := range modeSet {
			if k != "" {
				modes = append(modes, k)
			}
		}

		// Get matches for current page
		offset := (page - 1) * pageSize
		matches, err := re.App.FindRecordsByFilter(
			"matches",
			filterStr,
			"-end_time",
			pageSize+1, // Get one extra to determine if there's a next page
			offset,
			filterParams,
		)

		hasNextPage := len(matches) > pageSize
		if hasNextPage {
			matches = matches[:pageSize]
		}

		type MatchPlayer struct {
			PlayerName string
			Team       string
			Kills      int
			Deaths     int
			Assists    int
			KDRatio    string
		}

		type MatchData struct {
			MatchId         string
			Map             string
			Mode            string
			Duration        string
			EndTime         string
			SecurityKills   int
			SecurityDeaths  int
			InsurgentKills  int
			InsurgentDeaths int
			Players         []MatchPlayer
		}

		matchData := make([]MatchData, 0, len(matches))
		for _, match := range matches {
			startTime := match.GetDateTime("start_time").Time()
			endTime := match.GetDateTime("end_time").Time()
			duration := endTime.Sub(startTime)

			md := MatchData{
				MatchId:  match.Id,
				Map:      match.GetString("map"),
				Mode:     match.GetString("mode"),
				Duration: fmt.Sprintf("%dh %dm", int(duration.Hours()), int(duration.Minutes())%60),
				EndTime:  endTime.Format("2006-01-02 15:04"),
			}

			// Get player stats for this match
			playerStats, err := re.App.FindRecordsByFilter(
				"match_player_stats",
				"match = {:matchId}",
				"-kills",
				-1,
				0,
				map[string]any{"matchId": match.Id},
			)

			if err == nil {
				re.App.ExpandRecords(playerStats, []string{"player"}, nil)

				for _, stat := range playerStats {
					playerRec := stat.ExpandedOne("player")
					if playerRec == nil {
						continue
					}

					kills := stat.GetInt("kills")
					deaths := stat.GetInt("deaths")
					assists := stat.GetInt("assists")

					kdRatio := 0.0
					if deaths > 0 {
						kdRatio = float64(kills) / float64(deaths)
					} else if kills > 0 {
						kdRatio = float64(kills)
					}

					player := MatchPlayer{
						PlayerName: playerRec.GetString("name"),
						Kills:      kills,
						Deaths:     deaths,
						Assists:    assists,
						KDRatio:    fmt.Sprintf("%.2f", kdRatio),
						Team:       stat.GetString("team"),
					}

					if player.Team == "Security" {
						md.SecurityKills += kills
						md.SecurityDeaths += deaths
					} else {
						md.InsurgentKills += kills
						md.InsurgentDeaths += deaths
					}

					md.Players = append(md.Players, player)
				}
			}

			matchData = append(matchData, md)
		}

		html, err := registry.LoadFS(assets.GetWebAssets().FS(),
			"templates/layout.html",
			"templates/match-history.html",
		).Render(map[string]any{
			"ActivePage":     "match-history",
			"Matches":        matchData,
			"Servers":        servers,
			"Maps":           maps,
			"Modes":          modes,
			"SelectedServer": selectedServer,
			"SelectedMap":    selectedMap,
			"SelectedMode":   selectedMode,
			"Page":           page,
			"NextPage":       page + 1,
			"HasNextPage":    hasNextPage,
		})

		if err != nil {
			return re.InternalServerError("Failed to render template", err)
		}

		return re.HTML(http.StatusOK, html)
	})

	// Server Stats page - player statistics per server
	e.Router.GET("/servers/{id}/stats", func(re *core.RequestEvent) error {
		serverID := re.Request.PathValue("id")
		searchQuery := re.Request.URL.Query().Get("search")
		sortBy := re.Request.URL.Query().Get("sort")
		if sortBy == "" {
			sortBy = "kills"
		}

		// Get all servers for dropdown
		servers, err := re.App.FindAllRecords("servers")
		if err != nil {
			servers = []*core.Record{}
		}

		type ServerInfo struct {
			ID   string
			Name string
		}

		serverList := make([]ServerInfo, len(servers))
		currentServerName := ""
		for i, server := range servers {
			serverList[i] = ServerInfo{
				ID:   server.Id,
				Name: server.GetString("name"),
			}
			if server.Id == serverID {
				currentServerName = server.GetString("name")
			}
		}

		type PlayerStatRow struct {
			Name       string
			Kills      int
			Deaths     int
			KDRatio    float64
			KDRatioStr string
			Score      int
			MatchCount int
			LastSeen   string
		}

		playerStats := make([]PlayerStatRow, 0)
		totalKills := 0
		totalDeaths := 0

		// Get all players that have played on this server
		if serverID != "" {
			// Get all matches for this server
			matches, err := re.App.FindRecordsByFilter(
				"matches",
				"server = {:serverId}",
				"",
				-1,
				0,
				map[string]any{"serverId": serverID},
			)

			playerMap := make(map[string]map[string]any)

			if err == nil {
				for _, match := range matches {
					// Get all player stats for this match
					playerStatsRecords, err := re.App.FindRecordsByFilter(
						"match_player_stats",
						"match = {:matchId}",
						"",
						-1,
						0,
						map[string]any{"matchId": match.Id},
					)

					if err == nil {
						for _, pstat := range playerStatsRecords {
							playerID := pstat.GetString("player")
							playerRecord, err := re.App.FindRecordById("players", playerID)
							if err != nil {
								continue
							}

							playerName := playerRecord.GetString("name")

							if _, exists := playerMap[playerID]; !exists {
								playerMap[playerID] = map[string]any{
									"name":      playerName,
									"kills":     0,
									"deaths":    0,
									"score":     0,
									"matches":   0,
									"last_seen": pstat.GetDateTime("updated").Time(),
								}
							}

							// Get weapon stats for this player in this match
							weaponStats, err := re.App.FindRecordsByFilter(
								"match_weapon_stats",
								"player = {:playerId} && match = {:matchId}",
								"",
								-1,
								0,
								map[string]any{
									"playerId": playerID,
									"matchId":  match.Id,
								},
							)

							kills := 0
							if err == nil {
								for _, ws := range weaponStats {
									kills += ws.GetInt("kills")
								}
							}

							current := playerMap[playerID]
							current["kills"] = current["kills"].(int) + kills
							current["deaths"] = current["deaths"].(int) + pstat.GetInt("deaths")
							current["score"] = current["score"].(int) + pstat.GetInt("score")
							current["matches"] = current["matches"].(int) + 1

							// Update last seen
							updated := pstat.GetDateTime("updated").Time()
							if updated.After(current["last_seen"].(time.Time)) {
								current["last_seen"] = updated
							}
						}
					}
				}
			}

			// Apply search filter
			for _, data := range playerMap {
				playerName := data["name"].(string)
				if searchQuery != "" && !strings.Contains(strings.ToLower(playerName), strings.ToLower(searchQuery)) {
					continue
				}

				kills := data["kills"].(int)
				deaths := data["deaths"].(int)
				score := data["score"].(int)
				matchCount := data["matches"].(int)
				lastSeen := data["last_seen"].(time.Time)

				kdRatio := 0.0
				kdRatioStr := "0.00"
				if deaths > 0 {
					kdRatio = float64(kills) / float64(deaths)
					kdRatioStr = fmt.Sprintf("%.2f", kdRatio)
				} else if kills > 0 {
					kdRatio = float64(kills)
					kdRatioStr = "∞"
				}

				playerStats = append(playerStats, PlayerStatRow{
					Name:       playerName,
					Kills:      kills,
					Deaths:     deaths,
					KDRatio:    kdRatio,
					KDRatioStr: kdRatioStr,
					Score:      score,
					MatchCount: matchCount,
					LastSeen:   lastSeen.Format("2006-01-02 15:04"),
				})

				totalKills += kills
				totalDeaths += deaths
			}

			// Sort based on parameter
			switch sortBy {
			case "deaths":
				// Sort by deaths descending
				for i := 0; i < len(playerStats)-1; i++ {
					for j := i + 1; j < len(playerStats); j++ {
						if playerStats[j].Deaths > playerStats[i].Deaths {
							playerStats[i], playerStats[j] = playerStats[j], playerStats[i]
						}
					}
				}
			case "kd":
				// Sort by K/D ratio descending
				for i := 0; i < len(playerStats)-1; i++ {
					for j := i + 1; j < len(playerStats); j++ {
						if playerStats[j].KDRatio > playerStats[i].KDRatio {
							playerStats[i], playerStats[j] = playerStats[j], playerStats[i]
						}
					}
				}
			case "score":
				// Sort by score descending
				for i := 0; i < len(playerStats)-1; i++ {
					for j := i + 1; j < len(playerStats); j++ {
						if playerStats[j].Score > playerStats[i].Score {
							playerStats[i], playerStats[j] = playerStats[j], playerStats[i]
						}
					}
				}
			case "matches":
				// Sort by matches descending
				for i := 0; i < len(playerStats)-1; i++ {
					for j := i + 1; j < len(playerStats); j++ {
						if playerStats[j].MatchCount > playerStats[i].MatchCount {
							playerStats[i], playerStats[j] = playerStats[j], playerStats[i]
						}
					}
				}
			default:
				// Sort by kills descending (default)
				for i := 0; i < len(playerStats)-1; i++ {
					for j := i + 1; j < len(playerStats); j++ {
						if playerStats[j].Kills > playerStats[i].Kills {
							playerStats[i], playerStats[j] = playerStats[j], playerStats[i]
						}
					}
				}
			}
		}

		// Calculate server K/D ratio
		serverKDRatio := "0.00"
		if totalDeaths > 0 {
			serverKDRatio = fmt.Sprintf("%.2f", float64(totalKills)/float64(totalDeaths))
		} else if totalKills > 0 {
			serverKDRatio = "∞"
		}

		// Check if this is an HTMX request (partial update)
		isHTMX := re.Request.Header.Get("HX-Request") == "true"

		var html string
		if isHTMX {
			// Return just the table for HTMX updates
			html, err = registry.LoadFS(assets.GetWebAssets().FS(),
				"templates/server_stats_table.html",
			).Render(map[string]any{
				"PlayerStats": playerStats,
				"SearchQuery": searchQuery,
			})
		} else {
			// Return full page
			html, err = registry.LoadFS(assets.GetWebAssets().FS(),
				"templates/layout.html",
				"templates/server_stats.html",
			).Render(map[string]any{
				"ActivePage":      "server-stats",
				"CurrentServerID": serverID,
				"ServerName":      currentServerName,
				"Servers":         serverList,
				"PlayerStats":     playerStats,
				"SearchQuery":     searchQuery,
				"SortBy":          sortBy,
				"TotalPlayers":    len(playerStats),
				"TotalKills":      totalKills,
				"TotalDeaths":     totalDeaths,
				"ServerKDRatio":   serverKDRatio,
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

	app.Logger().Info("Registered custom HTTP handlers")

	// Note: Server management API endpoints are registered by the servermgr plugin
	// See internal/servermgr/plugin.go for:
	// - POST /api/server/start
	// - POST /api/server/stop
	// - GET /api/server/status
	// - GET /api/server/list
}

// contains performs a case-insensitive substring search
func contains(s, substr string) bool {
	s = strings.ToLower(s)
	substr = strings.ToLower(substr)
	return strings.Contains(s, substr)
}
