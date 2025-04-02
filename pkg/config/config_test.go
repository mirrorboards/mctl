package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestInitConfig(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "mctl-config-test")
	if err != nil {
		t.Fatalf("Error creating temporary directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Change to the temporary directory
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Error getting current directory: %v", err)
	}
	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("Error changing to temporary directory: %v", err)
	}
	defer os.Chdir(originalDir)

	// Test successful config creation
	err = InitConfig()
	if err != nil {
		t.Fatalf("Error initializing config: %v", err)
	}

	// Check that the file exists
	configPath := filepath.Join(tempDir, configFileName)
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Errorf("Config file was not created")
	}

	// Test that calling it a second time returns an error
	err = InitConfig()
	if err == nil {
		t.Errorf("InitConfig should return an error when file already exists")
	}
}

func TestExtractRepoName(t *testing.T) {
	testCases := []struct {
		url      string
		expected string
	}{
		{"https://github.com/user/repo.git", "repo"},
		{"https://github.com/user/repo", "repo"},
		{"git@github.com:user/repo.git", "repo"},
		{"git@github.com:user/repo", "repo"},
		{"ssh://git@github.com/user/repo.git", "repo"},
		{"https://github.com/user/repo-with-dashes", "repo-with-dashes"},
		{"https://github.com/user/repo_with_underscores", "repo_with_underscores"},
		{"", ""},
	}

	for _, tc := range testCases {
		result := ExtractRepoName(tc.url)
		if result != tc.expected {
			t.Errorf("ExtractRepoName(%s) = %s, expected %s", tc.url, result, tc.expected)
		}
	}
}

func TestAddRepository(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "mctl-config-add-test")
	if err != nil {
		t.Fatalf("Error creating temporary directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Change to the temporary directory
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Error getting current directory: %v", err)
	}
	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("Error changing to temporary directory: %v", err)
	}
	defer os.Chdir(originalDir)

	// Test adding a repository with a name
	gitURL := "https://github.com/user/repo.git"
	targetPath := "./repos"
	name := "custom-name"

	err = AddRepository(gitURL, targetPath, name)
	if err != nil {
		t.Fatalf("Error adding repository: %v", err)
	}

	// Check that the config file exists and contains the repo entry
	configPath := filepath.Join(tempDir, configFileName)
	configData, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("Error reading config file: %v", err)
	}

	configContent := string(configData)
	if !strings.Contains(configContent, "url = \""+gitURL+"\"") {
		t.Errorf("Config doesn't contain the URL: %s", gitURL)
	}
	if !strings.Contains(configContent, "path = \""+targetPath+"\"") {
		t.Errorf("Config doesn't contain the path: %s", targetPath)
	}
	if !strings.Contains(configContent, "name = \""+name+"\"") {
		t.Errorf("Config doesn't contain the name: %s", name)
	}

	// Test adding a repository without a name
	gitURL2 := "https://github.com/user/repo2.git"
	targetPath2 := "./repos2"
	err = AddRepository(gitURL2, targetPath2, "")
	if err != nil {
		t.Fatalf("Error adding repository without name: %v", err)
	}

	// Read the updated config and check the second entry
	configData, err = os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("Error reading updated config file: %v", err)
	}

	configContent = string(configData)
	if !strings.Contains(configContent, "url = \""+gitURL2+"\"") {
		t.Errorf("Config doesn't contain the second URL: %s", gitURL2)
	}
	if !strings.Contains(configContent, "path = \""+targetPath2+"\"") {
		t.Errorf("Config doesn't contain the second path: %s", targetPath2)
	}
	if strings.Count(configContent, "name = \"") != 1 {
		t.Errorf("Config should have exactly one name entry")
	}
}

func TestGetAllRepositories(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "mctl-config-getall-test")
	if err != nil {
		t.Fatalf("Error creating temporary directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Change to the temporary directory
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Error getting current directory: %v", err)
	}
	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("Error changing to temporary directory: %v", err)
	}
	defer os.Chdir(originalDir)

	// Test error when config file doesn't exist
	_, err = GetAllRepositories()
	if err == nil {
		t.Errorf("GetAllRepositories should return error when config doesn't exist")
	}

	// Initialize an empty config
	if err := InitConfig(); err != nil {
		t.Fatalf("Error initializing config: %v", err)
	}

	// Test empty config
	repos, err := GetAllRepositories()
	if err != nil {
		t.Fatalf("Error getting repositories from empty config: %v", err)
	}
	if len(repos) != 0 {
		t.Errorf("Expected 0 repositories in empty config, got %d", len(repos))
	}

	// Add repositories to the config
	testRepos := []struct {
		url  string
		path string
		name string
	}{
		{"https://github.com/test1/repo1.git", "./path1", "name1"},
		{"https://github.com/test2/repo2.git", "./path2", "name2"},
		{"https://github.com/test3/repo3.git", "./path3", ""},
	}

	for _, repo := range testRepos {
		if err := AddRepository(repo.url, repo.path, repo.name); err != nil {
			t.Fatalf("Error adding repository %s: %v", repo.url, err)
		}
	}

	// Test getting all repositories
	repos, err = GetAllRepositories()
	if err != nil {
		t.Fatalf("Error getting repositories: %v", err)
	}

	if len(repos) != len(testRepos) {
		t.Errorf("Expected %d repositories, got %d", len(testRepos), len(repos))
	}

	// Check that all repositories were added correctly
	for i, expected := range testRepos {
		if repos[i].URL != expected.url {
			t.Errorf("Repository %d: expected URL %s, got %s", i, expected.url, repos[i].URL)
		}
		if repos[i].Path != expected.path {
			t.Errorf("Repository %d: expected Path %s, got %s", i, expected.path, repos[i].Path)
		}
		if repos[i].Name != expected.name {
			t.Errorf("Repository %d: expected Name %s, got %s", i, expected.name, repos[i].Name)
		}
	}
}
