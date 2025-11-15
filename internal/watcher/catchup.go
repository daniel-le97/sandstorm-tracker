package watcher

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"sandstorm-tracker/internal/config"
	"sandstorm-tracker/internal/database"
	"sandstorm-tracker/internal/parser"

	"github.com/pocketbase/pocketbase/core"
)

// CatchupProcessor handles startup catch-up logic
type CatchupProcessor struct {
	parser        *parser.LogParser
	a2sPool       A2SQuerier
	serverConfigs map[string]config.ServerConfig
	pbApp         core.App
	ctx           context.Context
}

// NewCatchupProcessor creates a new catchup processor
func NewCatchupProcessor(
	parser *parser.LogParser,
	a2sPool A2SQuerier,
	serverConfigs map[string]config.ServerConfig,
	pbApp core.App,
	ctx context.Context,
) *CatchupProcessor {
	return &CatchupProcessor{
		parser:        parser,
		a2sPool:       a2sPool,
		serverConfigs: serverConfigs,
		pbApp:         pbApp,
		ctx:           ctx,
	}
}

// CheckStartupCatchup determines if we should do catch-up processing on tracker startup
// Returns the offset to start watching from and whether catch-up was performed
func (c *CatchupProcessor) CheckStartupCatchup(filePath, serverID string) (int, bool) {
	// Get server config to access query address
	serverConfig, exists := c.serverConfigs[serverID]
	if !exists {
		log.Printf("[Catchup] No config found for server %s, skipping catch-up", serverID)
		return 0, false
	}

	// Check 1: Query A2S to verify server is online and get current map
	if serverConfig.QueryAddress == "" {
		log.Printf("[Catchup] No query address configured for %s, skipping catch-up", serverID)
		return 0, false
	}

	ctx, cancel := context.WithTimeout(c.ctx, 5*time.Second)
	defer cancel()

	serverStatus, err := c.a2sPool.QueryServer(ctx, serverConfig.QueryAddress)
	if err != nil {
		log.Printf("[Catchup] Server %s appears offline (A2S query failed: %v), skipping catch-up", serverID, err)
		return 0, false
	}

	if serverStatus == nil || serverStatus.Info == nil {
		log.Printf("[Catchup] Server %s returned no info, skipping catch-up", serverID)
		return 0, false
	}

	currentMap := serverStatus.Info.Map
	log.Printf("[Catchup] Server %s is online, current map: %s", serverID, currentMap)

	// Check 2: Is file recently modified?
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		log.Printf("Failed to stat file %s: %v", filePath, err)
		return 0, false
	}

	fileModTime := fileInfo.ModTime()
	timeSinceModification := time.Since(fileModTime)

	// Detect if SAW is active by checking for recent RCON logs
	sawActive := c.hasRecentRconLogs(filePath, 30*time.Second)

	// Use adaptive threshold based on whether SAW is active
	var fileModThreshold time.Duration
	if sawActive {
		fileModThreshold = 1 * time.Minute // SAW keeps file fresh with polling
		log.Printf("[Catchup] SAW detected for %s, using 1-minute threshold", serverID)
	} else {
		fileModThreshold = 9 * time.Hour // Servers restart every 8 hours, allow some buffer
		log.Printf("[Catchup] No SAW detected for %s, using 9-hour threshold", serverID)
	}

	fileRecentlyModified := timeSinceModification < fileModThreshold

	if !fileRecentlyModified {
		log.Printf("[Catchup] File %s not recently modified (%.1f minutes ago), skipping catch-up",
			serverID, timeSinceModification.Minutes())
		return 0, false
	}

	// Check 3: Find last map event in log file
	mapName, scenario, mapTime, startLineNum, err := c.parser.FindLastMapEvent(filePath, time.Now())
	if err != nil {
		log.Printf("[Catchup] No map event found for %s: %v, skipping catch-up", serverID, err)
		return 0, false
	}

	timeSinceMap := time.Since(mapTime)
	recentMapEvent := timeSinceMap < 30*time.Minute

	if !recentMapEvent {
		log.Printf("[Catchup] Map event too old for %s (%.1f minutes ago), skipping catch-up",
			serverID, timeSinceMap.Minutes())
		return 0, false
	}

	// Check 4: Does the log map match the current server map?
	if !strings.EqualFold(mapName, currentMap) {
		log.Printf("[Catchup] Map mismatch for %s: log has '%s' but server is on '%s', skipping catch-up (can't get scores for old match)",
			serverID, mapName, currentMap)
		return 0, false
	}

	// All conditions met - do catch-up!
	log.Printf("[Catchup] Starting catch-up for %s: server online, map matches (%s), file modified %.1fs ago, map loaded %.1fs ago",
		serverID, mapName, timeSinceModification.Seconds(), timeSinceMap.Seconds())

	// Get current file size as the catch-up end point
	catchupEndOffset := fileInfo.Size()

	// Extract player team from scenario
	var playerTeam *string
	if strings.Contains(scenario, "_Security") {
		team := "Security"
		playerTeam = &team
	} else if strings.Contains(scenario, "_Insurgents") {
		team := "Insurgents"
		playerTeam = &team
	}

	// Create match in database
	_, err = c.pbApp.FindFirstRecordByFilter(
		"matches",
		"server = {:server} && end_time = ''",
		map[string]any{"server": serverID},
	)

	// Only create match if one doesn't already exist
	if err != nil {
		_, err = database.CreateMatch(c.ctx, c.pbApp, serverID, &mapName, &scenario, &mapTime, playerTeam)
		if err != nil {
			log.Printf("[Catchup] Failed to create match for %s: %v", serverID, err)
			return 0, false
		}
		log.Printf("[Catchup] Created match for %s: %s (%s) at %v", serverID, mapName, scenario, mapTime)
	} else {
		log.Printf("[Catchup] Active match already exists for %s, using existing match", serverID)
	}

	// Process historical events from map event to current position
	linesProcessed := c.processHistoricalEvents(filePath, serverID, startLineNum, catchupEndOffset)

	log.Printf("[Catchup] Completed for %s: processed %d lines from %d to %d",
		serverID, linesProcessed, startLineNum, catchupEndOffset)

	return int(catchupEndOffset), true
}

// hasRecentRconLogs checks if there are recent RCON log entries (indicates SAW is active)
func (c *CatchupProcessor) hasRecentRconLogs(filePath string, threshold time.Duration) bool {
	file, err := os.Open(filePath)
	if err != nil {
		return false
	}
	defer file.Close()

	// Read last 100 lines to check for recent RCON activity
	scanner := bufio.NewScanner(file)
	var lastLines []string
	maxLines := 100

	for scanner.Scan() {
		lastLines = append(lastLines, scanner.Text())
		if len(lastLines) > maxLines {
			lastLines = lastLines[1:]
		}
	}

	// Check if any of the last lines contain RCON log entries within threshold
	cutoffTime := time.Now().Add(-threshold)
	timestampPattern := regexp.MustCompile(`^\[(\d{4}\.\d{2}\.\d{2}-\d{2}\.\d{2}\.\d{2}:\d{3})\]`)

	for _, line := range lastLines {
		if strings.Contains(line, "LogRcon:") {
			// Try to extract timestamp from line
			if matches := timestampPattern.FindStringSubmatch(line); len(matches) >= 2 {
				if ts, err := parseTimestampFromLog(matches[1]); err == nil {
					if ts.After(cutoffTime) {
						return true
					}
				}
			}
		}
	}

	return false
}

// processHistoricalEvents processes events from a specific line number to an offset
func (c *CatchupProcessor) processHistoricalEvents(filePath, serverID string, startLine int, endOffset int64) int {
	file, err := os.Open(filePath)
	if err != nil {
		log.Printf("[Catchup] Failed to open file for historical processing: %v", err)
		return 0
	}
	defer file.Close()

	// Create context that marks this as catchup mode
	// Events will be created, but side effects (scoring, RCON) will be skipped
	type contextKey string
	isCatchupModeKey := contextKey("isCatchupMode")
	catchupCtx := context.WithValue(c.ctx, isCatchupModeKey, true)
	log.Printf("[Catchup] Processing historical events in catchup mode (no scoring/RCON)")

	scanner := bufio.NewScanner(file)
	lineNum := 0
	linesProcessed := 0

	// Read until we hit the start line
	for scanner.Scan() && lineNum < startLine {
		lineNum++
	}

	// Process lines from startLine until we reach endOffset
	for scanner.Scan() {
		currentPos, _ := file.Seek(0, 1)
		if currentPos > endOffset {
			break
		}

		line := scanner.Text()
		if err := c.parser.ParseAndProcess(catchupCtx, line, serverID, filePath); err != nil {
			log.Printf("[Catchup] Error processing line %d: %v", lineNum, err)
		}

		linesProcessed++
		lineNum++
	}

	return linesProcessed
}

// parseTimestampFromLog parses a timestamp from log format (2025.10.04-15.23.38:790)
func parseTimestampFromLog(ts string) (time.Time, error) {
	colonIdx := strings.LastIndex(ts, ":")
	if colonIdx == -1 {
		return time.Time{}, fmt.Errorf("invalid timestamp format: %s", ts)
	}

	dateTimePart := ts[:colonIdx]
	msPart := ts[colonIdx+1:]

	// Parse in local timezone (log timestamps are in server's local time)
	dt, err := time.ParseInLocation("2006.01.02-15.04.05", dateTimePart, time.Local)
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to parse datetime part: %w", err)
	}

	ms, err := strconv.Atoi(msPart)
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to parse milliseconds: %w", err)
	}

	return dt.Add(time.Duration(ms) * time.Millisecond), nil
}
