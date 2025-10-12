package events

import (
	"testing"
)

func TestParseServerArgsEvent(t *testing.T) {
	line := `[2025.10.04-15.23.38:790][  0]LogInit: Command Line: "/Game/Maps/Hideout?Scenario=Scenario_Security?MaxPlayers=10?Lighting=Day" -Port=7777 -QueryPort=27015 -log=9fa1f292-8394-401f-986f-26207fb9f9e8.log -hostname="-=312th=- TD's Tough Tavern Dallas I [HC]" -GSLTToken=SECRET123`
	event := parseServerArgsEvent(line)
	if event == nil {
		t.Fatal("expected event, got nil")
	}
	if len(event.Args) == 0 {
		t.Error("expected args to be parsed")
	}
	foundRedacted := false
	for _, arg := range event.Args {
		if arg == "-GSLTToken=REDACTED" {
			foundRedacted = true
		}
		if arg == "-GSLTToken=SECRET123" {
			t.Error("GSLTToken was not redacted")
		}
	}
	if !foundRedacted {
		t.Error("expected GSLTToken to be redacted")
	}
	if event.ServerID != "9fa1f292-8394-401f-986f-26207fb9f9e8" {
		t.Errorf("expected serverID 9fa1f292-8394-401f-986f-26207fb9f9e8, got %s", event.ServerID)
	}
	if event.RawCommand == "" {
		t.Error("expected raw command to be set")
	}

	// Extract map, scenario, players, port, queryport, hostname from args
	var mapName, scenario, players, port, queryport, hostname string
	for _, arg := range event.Args {
		if len(arg) > 0 && arg[0] == '"' {
			// Map arg: "/Game/Maps/Hideout?Scenario=Scenario_Security?MaxPlayers=10?Lighting=Day"
			s := arg
			if idx := findMapName(s); idx != "" {
				mapName = idx
			}
			if idx := findScenario(s); idx != "" {
				scenario = idx
			}
			if idx := findPlayers(s); idx != "" {
				players = idx
			}
		} else if len(arg) > 6 && arg[:6] == "-Port=" {
			port = arg[6:]
		} else if len(arg) > 11 && arg[:11] == "-QueryPort=" {
			queryport = arg[11:]
		} else if len(arg) > 10 && arg[:10] == "-hostname=" {
			hn := arg[10:]
			// Remove surrounding quotes if present
			if len(hn) > 1 && hn[0] == '"' && hn[len(hn)-1] == '"' {
				hn = hn[1 : len(hn)-1]
			}
			hostname = hn
		}
	}
	if mapName != "Hideout" {
		t.Errorf("expected map Hideout, got %s", mapName)
	}
	if scenario != "Security" {
		t.Errorf("expected scenario Security, got %s", scenario)
	}
	if players != "10" {
		t.Errorf("expected players 10, got %s", players)
	}
	if port != "7777" {
		t.Errorf("expected port 7777, got %s", port)
	}
	if queryport != "27015" {
		t.Errorf("expected queryport 27015, got %s", queryport)
	}
	if hostname != "-=312th=- TD's Tough Tavern Dallas I [HC]" {
		t.Errorf("expected hostname, got %s", hostname)
	}
}

// Helpers for extracting values from the quoted map arg
func findMapName(s string) string {
	// "/Game/Maps/Hideout?Scenario=Scenario_Security?MaxPlayers=10?Lighting=Day"
	s = s[1:]
	if idx := findIndex(s, "/Game/Maps/"); idx != -1 {
		s = s[idx+11:]
		if end := findIndex(s, "?"); end != -1 {
			return s[:end]
		}
	}
	return ""
}
func findScenario(s string) string {
	if idx := findIndex(s, "Scenario=Scenario_"); idx != -1 {
		s = s[idx+18:]
		if end := findIndex(s, "?"); end != -1 {
			return s[:end]
		}
	}
	return ""
}
func findPlayers(s string) string {
	if idx := findIndex(s, "MaxPlayers="); idx != -1 {
		s = s[idx+11:]
		if end := findIndex(s, "?"); end != -1 {
			return s[:end]
		}
		return s
	}
	return ""
}
func findIndex(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}
