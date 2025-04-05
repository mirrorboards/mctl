package cmd

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
	"text/tabwriter"

	"github.com/mirrorboards/mctl/internal/config"
	"github.com/mirrorboards/mctl/internal/errors"
	"github.com/mirrorboards/mctl/internal/repository"
	"github.com/spf13/cobra"
)

func newListCmd() *cobra.Command {
	var (
		format   string
		columns  string
		filter   string
		sortBy   string
		detailed bool
	)

	cmd := &cobra.Command{
		Use:     "list [options]",
		Aliases: []string{"ls"},
		Short:   "List repositories under MCTL management",
		Long: `List repositories under MCTL management.

This command displays information about repositories managed by MCTL.
You can customize the output format, columns, filtering, and sorting.

Available columns:
- id: Unique repository identifier
- name: Repository name
- path: Local filesystem path
- url: Remote repository URL
- branch: Current branch
- status: Repository status
- last_sync: Last synchronization timestamp

Examples:
  mctl list
  mctl list --format=json
  mctl list --columns=id,name,status
  mctl list --filter="status=CLEAN"
  mctl list --sort=name
  mctl list --detailed`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runList(format, columns, filter, sortBy, detailed)
		},
	}

	// Add flags
	cmd.Flags().StringVar(&format, "format", "table", "Output format (table, json, text, csv)")
	cmd.Flags().StringVar(&columns, "columns", "id,name,path,branch,status", "Columns to display (comma-separated)")
	cmd.Flags().StringVar(&filter, "filter", "", "Filter expression for repository selection")
	cmd.Flags().StringVar(&sortBy, "sort", "name", "Sort order specification")
	cmd.Flags().BoolVar(&detailed, "detailed", false, "Include extended metadata")

	return cmd
}

func runList(format, columns, filter, sortBy string, detailed bool) error {
	// Get current directory
	currentDir, err := os.Getwd()
	if err != nil {
		return errors.Wrap(err, errors.ErrInternalError, "Failed to get current directory")
	}

	// Load configuration
	cfg, err := config.LoadConfig(currentDir)
	if err != nil {
		return errors.Wrap(err, errors.ErrConfigNotFound, "Failed to load configuration")
	}

	// Create repository manager
	repoManager := repository.NewManager(cfg, currentDir)

	// Get all repositories
	repos, err := repoManager.GetAllRepositories()
	if err != nil {
		return errors.Wrap(err, errors.ErrInternalError, "Failed to get repositories")
	}

	// Apply filter if specified
	if filter != "" {
		repos, err = filterRepositories(repos, filter)
		if err != nil {
			return errors.Wrap(err, errors.ErrInvalidArgument, "Invalid filter expression")
		}
	}

	// Sort repositories
	sortRepositories(repos, sortBy)

	// Parse columns
	columnList := strings.Split(columns, ",")

	// Display repositories in the specified format
	switch format {
	case "table":
		displayTableFormat(repos, columnList, detailed)
	case "json":
		displayJSONFormat(repos, columnList, detailed)
	case "text":
		displayTextFormat(repos, columnList, detailed)
	case "csv":
		displayCSVFormat(repos, columnList, detailed)
	default:
		return errors.New(errors.ErrInvalidArgument, "Invalid format specification")
	}

	return nil
}

func filterRepositories(repos []*repository.Repository, filter string) ([]*repository.Repository, error) {
	// Simple filter implementation
	parts := strings.SplitN(filter, "=", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid filter format: %s", filter)
	}

	field := strings.TrimSpace(parts[0])
	value := strings.TrimSpace(parts[1])

	var filtered []*repository.Repository
	for _, repo := range repos {
		switch field {
		case "id":
			if repo.Config.ID == value {
				filtered = append(filtered, repo)
			}
		case "name":
			if repo.Config.Name == value {
				filtered = append(filtered, repo)
			}
		case "path":
			if repo.Config.Path == value {
				filtered = append(filtered, repo)
			}
		case "branch":
			if repo.Metadata.Status.Branch == value {
				filtered = append(filtered, repo)
			}
		case "status":
			if string(repo.Metadata.Status.Current) == value {
				filtered = append(filtered, repo)
			}
		default:
			return nil, fmt.Errorf("unknown field: %s", field)
		}
	}

	return filtered, nil
}

func sortRepositories(repos []*repository.Repository, sortBy string) {
	switch sortBy {
	case "id":
		sort.Slice(repos, func(i, j int) bool {
			return repos[i].Config.ID < repos[j].Config.ID
		})
	case "name":
		sort.Slice(repos, func(i, j int) bool {
			return repos[i].Config.Name < repos[j].Config.Name
		})
	case "path":
		sort.Slice(repos, func(i, j int) bool {
			return repos[i].Config.Path < repos[j].Config.Path
		})
	case "branch":
		sort.Slice(repos, func(i, j int) bool {
			return repos[i].Metadata.Status.Branch < repos[j].Metadata.Status.Branch
		})
	case "status":
		sort.Slice(repos, func(i, j int) bool {
			return string(repos[i].Metadata.Status.Current) < string(repos[j].Metadata.Status.Current)
		})
	case "last_sync":
		sort.Slice(repos, func(i, j int) bool {
			return repos[i].Metadata.Basic.LastSync.Before(repos[j].Metadata.Basic.LastSync)
		})
	}
}

func displayTableFormat(repos []*repository.Repository, columns []string, detailed bool) {
	// Create a new tabwriter
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	defer w.Flush()

	// Print header
	var header []string
	for _, col := range columns {
		header = append(header, strings.ToUpper(col))
	}
	fmt.Fprintln(w, strings.Join(header, "\t"))

	// Print repositories
	for _, repo := range repos {
		var row []string
		for _, col := range columns {
			switch col {
			case "id":
				row = append(row, repo.Config.ID)
			case "name":
				row = append(row, repo.Config.Name)
			case "path":
				row = append(row, repo.Config.Path)
			case "url":
				row = append(row, repo.Config.URL)
			case "branch":
				row = append(row, repo.Metadata.Status.Branch)
			case "status":
				row = append(row, string(repo.Metadata.Status.Current))
			case "last_sync":
				if !repo.Metadata.Basic.LastSync.IsZero() {
					row = append(row, repo.Metadata.Basic.LastSync.Format("2006-01-02 15:04:05"))
				} else {
					row = append(row, "Never")
				}
			default:
				row = append(row, "N/A")
			}
		}
		fmt.Fprintln(w, strings.Join(row, "\t"))
	}
}

func displayJSONFormat(repos []*repository.Repository, columns []string, detailed bool) {
	type jsonRepository struct {
		ID       string `json:"id,omitempty"`
		Name     string `json:"name,omitempty"`
		Path     string `json:"path,omitempty"`
		URL      string `json:"url,omitempty"`
		Branch   string `json:"branch,omitempty"`
		Status   string `json:"status,omitempty"`
		LastSync string `json:"last_sync,omitempty"`
	}

	var result []jsonRepository
	for _, repo := range repos {
		jr := jsonRepository{}
		for _, col := range columns {
			switch col {
			case "id":
				jr.ID = repo.Config.ID
			case "name":
				jr.Name = repo.Config.Name
			case "path":
				jr.Path = repo.Config.Path
			case "url":
				jr.URL = repo.Config.URL
			case "branch":
				jr.Branch = repo.Metadata.Status.Branch
			case "status":
				jr.Status = string(repo.Metadata.Status.Current)
			case "last_sync":
				if !repo.Metadata.Basic.LastSync.IsZero() {
					jr.LastSync = repo.Metadata.Basic.LastSync.Format("2006-01-02 15:04:05")
				} else {
					jr.LastSync = "Never"
				}
			}
		}
		result = append(result, jr)
	}

	// Marshal to JSON
	jsonData, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error marshaling to JSON: %v\n", err)
		return
	}

	fmt.Println(string(jsonData))
}

func displayTextFormat(repos []*repository.Repository, columns []string, detailed bool) {
	for _, repo := range repos {
		fmt.Printf("Repository: %s\n", repo.Config.Name)
		for _, col := range columns {
			switch col {
			case "id":
				fmt.Printf("  ID: %s\n", repo.Config.ID)
			case "name":
				fmt.Printf("  Name: %s\n", repo.Config.Name)
			case "path":
				fmt.Printf("  Path: %s\n", repo.Config.Path)
			case "url":
				fmt.Printf("  URL: %s\n", repo.Config.URL)
			case "branch":
				fmt.Printf("  Branch: %s\n", repo.Metadata.Status.Branch)
			case "status":
				fmt.Printf("  Status: %s\n", repo.Metadata.Status.Current)
			case "last_sync":
				if !repo.Metadata.Basic.LastSync.IsZero() {
					fmt.Printf("  Last Sync: %s\n", repo.Metadata.Basic.LastSync.Format("2006-01-02 15:04:05"))
				} else {
					fmt.Printf("  Last Sync: Never\n")
				}
			}
		}
		fmt.Println()
	}
}

func displayCSVFormat(repos []*repository.Repository, columns []string, detailed bool) {
	// Create a new CSV writer
	w := csv.NewWriter(os.Stdout)
	defer w.Flush()

	// Write header
	if err := w.Write(columns); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing CSV header: %v\n", err)
		return
	}

	// Write repositories
	for _, repo := range repos {
		var row []string
		for _, col := range columns {
			switch col {
			case "id":
				row = append(row, repo.Config.ID)
			case "name":
				row = append(row, repo.Config.Name)
			case "path":
				row = append(row, repo.Config.Path)
			case "url":
				row = append(row, repo.Config.URL)
			case "branch":
				row = append(row, repo.Metadata.Status.Branch)
			case "status":
				row = append(row, string(repo.Metadata.Status.Current))
			case "last_sync":
				if !repo.Metadata.Basic.LastSync.IsZero() {
					row = append(row, repo.Metadata.Basic.LastSync.Format("2006-01-02 15:04:05"))
				} else {
					row = append(row, "Never")
				}
			default:
				row = append(row, "")
			}
		}
		if err := w.Write(row); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing CSV row: %v\n", err)
			return
		}
	}
}
