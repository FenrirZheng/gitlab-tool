package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "usage: %s <path>\n", os.Args[0])
		os.Exit(1)
	}
	targetPath := os.Args[1]

	entries, err := os.ReadDir(targetPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to read directory %s: %v\n", targetPath, err)
		os.Exit(1)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		projectDir := filepath.Join(targetPath, entry.Name())

		if !isGitRepo(projectDir) {
			fmt.Printf("[SKIP] %s — not a git repo\n", entry.Name())
			continue
		}

		branch, err := getMostRecentBranch(projectDir)
		if err != nil {
			fmt.Printf("[ERROR] %s — %v\n", entry.Name(), err)
			continue
		}

		currentBranch, _ := getCurrentBranch(projectDir)
		if currentBranch == branch {
			fmt.Printf("[OK] %s — already on most recent branch: %s\n", entry.Name(), branch)
			continue
		}

		if err := checkoutBranch(projectDir, branch); err != nil {
			fmt.Printf("[ERROR] %s — checkout failed: %v\n", entry.Name(), err)
			continue
		}

		fmt.Printf("[DONE] %s — checked out to: %s (was: %s)\n", entry.Name(), branch, currentBranch)
	}
}

// isGitRepo checks if a directory contains a .git folder
func isGitRepo(dir string) bool {
	info, err := os.Stat(filepath.Join(dir, ".git"))
	if err != nil {
		return false
	}
	return info.IsDir()
}

// getCurrentBranch returns the current branch name
func getCurrentBranch(dir string) (string, error) {
	cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

// getMostRecentBranch finds the branch with the most recent commit.
// It considers both local and remote branches.
// For remote-only branches, it returns the short name (without "origin/").
func getMostRecentBranch(dir string) (string, error) {
	// Fetch latest remote info so we have up-to-date committer dates
	exec.Command("git", "-C", dir, "fetch", "--all", "--quiet").Run()

	// List all branches sorted by most recent committer date
	cmd := exec.Command("git", "for-each-ref",
		"--sort=-committerdate",
		"--format=%(refname:short)",
		"refs/heads/", "refs/remotes/",
	)
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("for-each-ref failed: %w", err)
	}

	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	if len(lines) == 0 || lines[0] == "" {
		return "", fmt.Errorf("no branches found")
	}

	// Pick the first (most recent) branch
	branch := lines[0]

	// Skip "origin/HEAD" — it's a symbolic ref, not a real branch
	for _, line := range lines {
		if !strings.HasSuffix(line, "/HEAD") {
			branch = line
			break
		}
	}

	// If it's a remote branch like "origin/feature-x", extract "feature-x"
	if strings.HasPrefix(branch, "origin/") {
		branch = strings.TrimPrefix(branch, "origin/")
	}

	return branch, nil
}

// checkoutBranch checks out the given branch.
// If the branch doesn't exist locally, it creates a tracking branch from origin.
func checkoutBranch(dir, branch string) error {
	// Try local checkout first
	cmd := exec.Command("git", "checkout", branch)
	cmd.Dir = dir
	if err := cmd.Run(); err != nil {
		// Branch might only exist on remote — create tracking branch
		cmd = exec.Command("git", "checkout", "-b", branch, "origin/"+branch)
		cmd.Dir = dir
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("checkout %s: %w", branch, err)
		}
	}
	return nil
}
