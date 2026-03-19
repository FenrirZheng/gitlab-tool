package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Project represents the relevant fields from the GitLab API response.
type Project struct {
	SSHURLToRepo string `json:"ssh_url_to_repo"`
}

func main() {
	gitlabURL := flag.String("url", "https://gitlab.example.com", "GitLab instance URL")
	groupID := flag.String("group", "", "Group ID or URL-encoded path")
	token := flag.String("token", "", "Personal access token")
	cloneDir := flag.String("dir", ".", "Directory to clone into")
	flag.Parse()

	if *groupID == "" || *token == "" {
		fmt.Fprintln(os.Stderr, "Usage: -group and -token are required")
		flag.PrintDefaults()
		os.Exit(1)
	}

	baseURL := strings.TrimRight(*gitlabURL, "/")
	client := &http.Client{}

	page := 1
	for {
		url := fmt.Sprintf("%s/api/v4/groups/%s/projects?per_page=100&page=%d&include_subgroups=true",
			baseURL, *groupID, page)

		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error creating request: %v\n", err)
			os.Exit(1)
		}
		req.Header.Set("PRIVATE-TOKEN", *token)

		resp, err := client.Do(req)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error fetching page %d: %v\n", page, err)
			os.Exit(1)
		}

		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading response: %v\n", err)
			os.Exit(1)
		}

		if resp.StatusCode != http.StatusOK {
			fmt.Fprintf(os.Stderr, "API error (HTTP %d): %s\n", resp.StatusCode, string(body))
			os.Exit(1)
		}

		var projects []Project
		if err := json.Unmarshal(body, &projects); err != nil {
			fmt.Fprintf(os.Stderr, "Error parsing JSON: %v\n", err)
			os.Exit(1)
		}

		if len(projects) == 0 {
			break
		}

		for _, p := range projects {
			repoName := repoNameFromURL(p.SSHURLToRepo)
			repoPath := filepath.Join(*cloneDir, repoName)

			if info, err := os.Stat(repoPath); err == nil && info.IsDir() {
				fmt.Printf("Fetching: %s\n", repoName)
				runGit("git", "-C", repoPath, "fetch", "--all")
			} else {
				fmt.Printf("Cloning: %s\n", repoName)
				runGit("git", "clone", p.SSHURLToRepo, repoPath)
			}
		}

		page++
	}

	fmt.Println("Done.")
}

// repoNameFromURL extracts the repo name from an SSH URL, stripping the .git suffix.
// e.g. "git@gitlab.example.com:group/repo.git" -> "repo"
func repoNameFromURL(sshURL string) string {
	base := filepath.Base(sshURL)
	return strings.TrimSuffix(base, ".git")
}

func runGit(name string, args ...string) {
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "  Warning: %v\n", err)
	}
}
