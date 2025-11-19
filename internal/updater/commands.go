package updater

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/pocketbase/pocketbase/core"
	"github.com/spf13/cobra"
)

// RegisterCommands registers update-related commands with PocketBase and returns the updater instance
func RegisterCommands(app core.App, rootCmd *cobra.Command, config Config, logger *slog.Logger) *Updater {
	updater := New(config)

	// check-updates command
	checkCmd := &cobra.Command{
		Use:   "check-updates",
		Short: "Check if a new version is available",
		RunE: func(cmd *cobra.Command, args []string) error {
			available, newVersion, err := updater.CheckForUpdates()
			if err != nil {
				return fmt.Errorf("failed to check for updates: %w", err)
			}

			if available {
				logger.Info("Update available", "current", config.CurrentVersion, "new", newVersion)
				fmt.Printf("Update available: %s -> %s\n", config.CurrentVersion, newVersion)
				return nil
			}

			logger.Info("No updates available", "version", config.CurrentVersion)
			fmt.Printf("You are running the latest version (%s)\n", config.CurrentVersion)
			return nil
		},
	}

	// update command
	updateCmd := &cobra.Command{
		Use:   "update",
		Short: "Download and apply the latest update",
		RunE: func(cmd *cobra.Command, args []string) error {
			logger.Info("Checking for updates...")

			available, newVersion, err := updater.CheckForUpdates()
			if err != nil {
				return fmt.Errorf("failed to check for updates: %w", err)
			}

			if !available {
				logger.Info("Already at latest version", "version", config.CurrentVersion)
				fmt.Printf("Already at latest version: %s\n", config.CurrentVersion)
				return nil
			}

			logger.Info("Update available", "current", config.CurrentVersion, "new", newVersion)
			fmt.Printf("Updating from %s to %s...\n", config.CurrentVersion, newVersion)

			release, err := updater.GetReleaseInfo(newVersion)
			if err != nil {
				return fmt.Errorf("failed to get release info: %w", err)
			}

			asset, err := updater.FindAssetForPlatform(release)
			if err != nil {
				return fmt.Errorf("failed to find asset: %w", err)
			}

			logger.Info("Downloading", "asset", asset.Name)
			fmt.Printf("Downloading %s...\n", asset.Name)

			tmpFile := config.BinaryName + ".tmp"
			if err := updater.DownloadAsset(*asset, tmpFile); err != nil {
				return fmt.Errorf("failed to download asset: %w", err)
			}
			defer os.Remove(tmpFile)

			logger.Info("Download complete", "file", tmpFile)
			fmt.Println("Download complete. Update will be applied on restart.")

			return nil
		},
	}

	// version command
	versionCmd := &cobra.Command{
		Use:   "version",
		Short: "Show version information",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Printf("%s version %s\n", config.BinaryName, config.CurrentVersion)
			return nil
		},
	}

	rootCmd.AddCommand(checkCmd, updateCmd, versionCmd)

	return updater
}
