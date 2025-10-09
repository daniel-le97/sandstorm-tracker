package main

import (
	"bufio"
	"os"
	"sandstorm-tracker/events"
	"strings"
	"testing"
)

func TestParseCheck(t *testing.T) {
	file, err := os.Open("hc.log")
	if err != nil {
		t.Fatalf("Error opening file: %v", err)
	}
	defer file.Close()

	parser := events.NewEventParser()
	scanner := NewScanner(file)
	lineNum := 0
	killEvents := 0
	rabbitKills := 0
	originKills := 0
	armoredBearKills := 0
	blueKills := 0
	unparsedKills := 0

	for scanner.Scan() {
		line := scanner.Text()
		lineNum++
		if strings.Contains(line, "killed") && strings.Contains(line, "with") && strings.Contains(line, "LogGameplayEvents") {
			killEvents++
			event, err := parser.ParseLine(line, "hc")
			if err != nil {
				unparsedKills++
				continue
			}
			if event == nil || event.Type != events.EventPlayerKill {
				unparsedKills++
				continue
			}
			killersData, ok := event.Data["killers"].([]events.Killer)
			if !ok {
				unparsedKills++
				continue
			}
			for _, killer := range killersData {
				switch {
				case strings.Contains(killer.Name, "Rabbit"):
					rabbitKills++
				case strings.Contains(killer.Name, "0rigin"):
					originKills++
				case strings.Contains(killer.Name, "ArmoredBear"):
					armoredBearKills++
				case strings.Contains(killer.Name, "Blue"):
					blueKills++
				}
			}
		}
	}

	t.Run("kill event count", func(t *testing.T) {
		if killEvents != 104 {
			t.Errorf("Expected 104 kill events, got %d", killEvents)
		}
	})
	t.Run("unparsed kill events", func(t *testing.T) {
		if unparsedKills != 0 {
			t.Errorf("Expected 0 unparsed kill events, got %d", unparsedKills)
		}
	})
	t.Run("Rabbit kills", func(t *testing.T) {
		if rabbitKills != 53 {
			t.Errorf("Expected 53 Rabbit kills, got %d", rabbitKills)
		}
	})
	t.Run("0rigin kills", func(t *testing.T) {
		if originKills != 36 {
			t.Errorf("Expected 36 0rigin kills, got %d", originKills)
		}
	})
	t.Run("ArmoredBear kills", func(t *testing.T) {
		if armoredBearKills != 19 {
			t.Errorf("Expected 19 ArmoredBear kills, got %d", armoredBearKills)
		}
	})
	t.Run("Blue kills", func(t *testing.T) {
		if blueKills != 9 {
			t.Errorf("Expected 9 Blue kills, got %d", blueKills)
		}
	})
	t.Run("total parsed kills", func(t *testing.T) {
		total := rabbitKills + originKills + armoredBearKills + blueKills
		if total != 117 {
			t.Errorf("Expected 117 total parsed kills, got %d", total)
		}
	})
}

// NewScanner is a helper to allow easy mocking in tests
func NewScanner(file *os.File) *bufio.Scanner {
	return bufio.NewScanner(file)
}
