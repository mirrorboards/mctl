package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"
	"time"

	"github.com/mirrorboards/mctl/internal/errors"
	"github.com/mirrorboards/mctl/internal/snapshot"
	"github.com/spf13/cobra"
)

func newSnapshotsCmd() *cobra.Command {
	var (
		detailed bool
		limit    int
		id       string
	)

	cmd := &cobra.Command{
		Use:   "snapshots [options]",
		Short: "List repository snapshots",
		Long: `List repository snapshots.

This command lists all available snapshots, showing their ID, creation time,
and description. With the --detailed flag, it also shows information about
repositories in each snapshot.

Examples:
  mctl snapshots
  mctl snapshots --detailed
  mctl snapshots --limit=5
  mctl snapshots --id=20250405-123456-abcdef12`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runSnapshots(detailed, limit, id)
		},
	}

	// Add flags
	cmd.Flags().BoolVar(&detailed, "detailed", false, "Show detailed information about repositories in each snapshot")
	cmd.Flags().IntVar(&limit, "limit", 0, "Limit to the most recent n snapshots")
	cmd.Flags().StringVar(&id, "id", "", "Show details for a specific snapshot ID")

	return cmd
}

func runSnapshots(detailed bool, limit int, id string) error {
	// Get current directory
	currentDir, err := os.Getwd()
	if err != nil {
		return errors.Wrap(err, errors.ErrInternalError, "Failed to get current directory")
	}

	// Create snapshot manager
	snapshotManager := snapshot.NewManager(currentDir)

	// If specific ID is requested, show details for that snapshot
	if id != "" {
		return showSnapshotDetails(snapshotManager, id)
	}

	// List snapshots
	snapshots, err := snapshotManager.ListSnapshots()
	if err != nil {
		return errors.Wrap(err, errors.ErrInternalError, "Failed to list snapshots")
	}

	if len(snapshots) == 0 {
		fmt.Println("No snapshots found")
		return nil
	}

	// Apply limit if specified
	if limit > 0 && limit < len(snapshots) {
		snapshots = snapshots[:limit]
	}

	// Create tabwriter for aligned output
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tCREATED\tREPOSITORIES\tDESCRIPTION")

	// Print snapshots
	for _, snap := range snapshots {
		// Format creation time
		createdAt := formatTime(snap.CreatedAt)

		// Format repositories count
		reposCount := fmt.Sprintf("%d", len(snap.Repositories))

		// Format description (truncate if too long)
		description := snap.Description
		if len(description) > 50 {
			description = description[:47] + "..."
		}

		fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", snap.ID, createdAt, reposCount, description)

		// Print detailed repository information if requested
		if detailed {
			fmt.Fprintln(w, "")
			for _, repo := range snap.Repositories {
				fmt.Fprintf(w, "  %s\t%s\t%s\t%s\n", repo.Name, repo.Branch, repo.CommitHash[:8], repo.Status)
			}
			fmt.Fprintln(w, "")
		}
	}

	w.Flush()
	return nil
}

func showSnapshotDetails(snapshotManager *snapshot.Manager, id string) error {
	// Load snapshot
	snap, err := snapshotManager.LoadSnapshot(id)
	if err != nil {
		return errors.Wrap(err, errors.ErrInternalError, fmt.Sprintf("Failed to load snapshot: %s", id))
	}

	// Print snapshot details
	fmt.Printf("Snapshot ID: %s\n", snap.ID)
	fmt.Printf("Created: %s\n", formatTime(snap.CreatedAt))
	fmt.Printf("Description: %s\n", snap.Description)
	fmt.Printf("Repositories: %d\n\n", len(snap.Repositories))

	// Create tabwriter for aligned output
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "NAME\tBRANCH\tCOMMIT\tSTATUS\tPATH")

	// Print repository details
	for _, repo := range snap.Repositories {
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
			repo.Name, repo.Branch, repo.CommitHash[:8], repo.Status, repo.Path)
	}

	w.Flush()
	return nil
}

// formatTime formats a time.Time for display
func formatTime(t time.Time) string {
	// If time is less than 24 hours ago, show relative time
	if time.Since(t) < 24*time.Hour {
		hours := int(time.Since(t).Hours())
		minutes := int(time.Since(t).Minutes()) % 60

		if hours > 0 {
			return fmt.Sprintf("%dh %dm ago", hours, minutes)
		}
		return fmt.Sprintf("%dm ago", minutes)
	}

	// Otherwise show date
	return t.Format("2006-01-02 15:04:05")
}
