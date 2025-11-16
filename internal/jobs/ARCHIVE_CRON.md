# Archive Cron Job - Data Retention

## Overview

A cron job that automatically archives data older than 30 days to keep the database lean and performant.

## Configuration

- **Schedule**: Daily at 2 AM UTC (configurable via cron expression `0 2 * * *`)
- **Retention Period**: 30 days
- **Transaction Mode**: All deletions run in a single transaction for consistency

## What Gets Archived

The cron job deletes the following in order:

1. **Events** - Game events older than 30 days
2. **Match Player Stats** - Per-player match statistics
3. **Match Weapon Stats** - Per-weapon match statistics
4. **Matches** - Match records that ended before the cutoff date

## Implementation Details

### File: `internal/jobs/archive_cron.go`

- `RegisterArchiveOldData(app, logger)` - Registers the cron job at startup
- `archiveOldData()` - Main archive logic with transaction wrapper
- `archiveMatches()` - Deletes old matches based on end_time
- `archiveMatchPlayerStats()` - Deletes old player statistics
- `archiveMatchWeaponStats()` - Deletes old weapon statistics
- `archiveEvents()` - Deletes old game events

### Registration

Added to `internal/app/app.go` in the `onServe()` function:

```go
jobs.RegisterArchiveOldData(app.PocketBase, app.Logger().With("component", "ARCHIVE_JOB"))
```

### Query Strategy

- Uses PocketBase `FindRecordsByFilter()` to find records older than cutoff date
- Sorts by creation time descending for efficiency
- Limits to 10,000 records per batch to avoid memory issues
- Uses `created < {:cutoff}` timestamp comparisons

## Logging

The job logs:

- Start/completion of archive operations
- Count of records archived per collection
- Cutoff date applied
- Any errors encountered during deletion (non-fatal - continues with other records)

## Customization

To change the archive schedule, edit the cron expression in `RegisterArchiveOldData()`:

- `"0 2 * * *"` = Daily at 2 AM UTC
- `"0 0 * * 0"` = Weekly on Sunday at midnight
- `"0 0 1 * *"` = Monthly on the 1st at midnight

To change the retention period, modify the cutoff calculation:

```go
cutoffDate := time.Now().AddDate(0, 0, -30)  // Change -30 to desired days
```

## Performance Considerations

- Deletes run in batches of up to 10,000 records
- Uses transactions to ensure all-or-nothing consistency
- Soft deletes are not used; records are permanently removed
- Best run during low-traffic periods (2 AM UTC)

## Future Enhancements

- Add option to export data to archive storage before deletion
- Make retention period configurable
- Add metrics/dashboards for archive operations
- Support for selective archiving (e.g., old matches only)
