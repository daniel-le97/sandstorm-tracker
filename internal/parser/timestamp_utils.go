package parser

import (
	"fmt"
	"regexp"
	"time"
)

// timestampPattern matches Insurgency log timestamps [YYYY.MM.DD-HH.MM.SS:mmm]
var timestampPattern = regexp.MustCompile(`\[(\d{4}\.\d{2}\.\d{2}-\d{2}\.\d{2}\.\d{2}:\d{1,3})\]`)

// ReplaceTimestamp replaces the timestamp in a log line with a new timestamp
// while preserving the rest of the line. The baseTime parameter is used to
// calculate relative timestamps from the original log timestamps.
//
// Example:
//
//	line := "[2025.10.04-21.27.51:780][866]LogNet: Join succeeded: Player"
//	newLine := ReplaceTimestamp(line, time.Now())
//	// Result: "[2025.11.11-14.30.00:123][866]LogNet: Join succeeded: Player"
func ReplaceTimestamp(line string, baseTime time.Time) string {
	return timestampPattern.ReplaceAllStringFunc(line, func(match string) string {
		// Format new timestamp in Insurgency format using baseTime directly
		return formatInsurgencyTimestamp(baseTime)
	})
}

// ReplaceTimestampWithOffset replaces the timestamp in a log line with a new timestamp
// based on a base time plus the offset from a reference timestamp. This preserves
// the relative timing of events in the log.
//
// Example:
//
//	referenceTime := parseTime("2025.10.04-21.27.00:000")
//	line := "[2025.10.04-21.27.51:780][866]LogNet: Join succeeded: Player"
//	newLine := ReplaceTimestampWithOffset(line, time.Now(), referenceTime)
//	// Result: Current time + 51.78 seconds
func ReplaceTimestampWithOffset(line string, baseTime, referenceTime time.Time) string {
	return timestampPattern.ReplaceAllStringFunc(line, func(match string) string {
		// Extract original timestamp
		originalTS := timestampPattern.FindStringSubmatch(match)
		if len(originalTS) < 2 {
			return match
		}

		// Parse original timestamp
		original, err := parseInsurgencyTimestamp(originalTS[1])
		if err != nil {
			return match
		}

		// Calculate offset from reference time
		offset := original.Sub(referenceTime)

		// Apply offset to base time
		newTime := baseTime.Add(offset)

		// Format new timestamp in Insurgency format
		return formatInsurgencyTimestamp(newTime)
	})
}

// parseInsurgencyTimestamp parses an Insurgency timestamp string
// Format: YYYY.MM.DD-HH.MM.SS:mmm
func parseInsurgencyTimestamp(ts string) (time.Time, error) {
	// Parse format: 2025.10.04-21.27.51:780
	layout := "2006.01.02-15.04.05"

	// Split timestamp and milliseconds
	parts := regexp.MustCompile(`^(.+):(\d{1,3})$`).FindStringSubmatch(ts)
	if len(parts) < 3 {
		return time.Time{}, fmt.Errorf("invalid timestamp format: %s", ts)
	}

	baseTime, err := time.Parse(layout, parts[1])
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to parse timestamp %s: %w", ts, err)
	}

	// Add milliseconds
	ms, _ := time.ParseDuration(parts[2] + "ms")
	return baseTime.Add(ms), nil
}

// formatInsurgencyTimestamp formats a time.Time into Insurgency log timestamp format
// Format: [YYYY.MM.DD-HH.MM.SS:mmm]
func formatInsurgencyTimestamp(t time.Time) string {
	ms := t.Nanosecond() / 1000000 // Convert nanoseconds to milliseconds
	return fmt.Sprintf("[%04d.%02d.%02d-%02d.%02d.%02d:%03d]",
		t.Year(), t.Month(), t.Day(),
		t.Hour(), t.Minute(), t.Second(), ms)
}

// UpdateLogTimestamps updates all timestamps in a slice of log lines to be
// relative to a new base time, preserving the original timing relationships.
// The first line's timestamp becomes the base time, and all subsequent lines
// maintain their original offset from the first line.
func UpdateLogTimestamps(lines []string, baseTime time.Time) ([]string, error) {
	if len(lines) == 0 {
		return lines, nil
	}

	// Find the first timestamp to use as reference
	var referenceTime time.Time
	for _, line := range lines {
		matches := timestampPattern.FindStringSubmatch(line)
		if len(matches) >= 2 {
			var err error
			referenceTime, err = parseInsurgencyTimestamp(matches[1])
			if err != nil {
				continue
			}
			break
		}
	}

	if referenceTime.IsZero() {
		return lines, fmt.Errorf("no valid timestamp found in log lines")
	}

	// Update all lines with new timestamps
	result := make([]string, len(lines))
	for i, line := range lines {
		result[i] = ReplaceTimestampWithOffset(line, baseTime, referenceTime)
	}

	return result, nil
}
