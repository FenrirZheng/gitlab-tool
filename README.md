# Git Tool

A collection of Go CLI tools for managing multiple Git repositories at once. Designed for developers who work across many projects in a GitLab group.

## Tools

### fetch-group

Bulk clone and fetch all projects from a GitLab group (including subgroups).

- Queries the GitLab API to list all projects under a group
- Clones new repositories that don't exist locally
- Runs `git fetch --all` on repositories that already exist
- Handles pagination automatically

**Usage:**

```bash
cd fetch-group && go build -o fetch-group .

./fetch-group \
  -url https://gitlab.example.com \
  -group <group-id-or-path> \
  -token <personal-access-token> \
  -dir ./repos
```

| Flag     | Description                      | Default                    |
|----------|----------------------------------|----------------------------|
| `-url`   | GitLab instance URL              | `https://gitlab.example.com` |
| `-group` | Group ID or URL-encoded path     | *(required)*               |
| `-token` | Personal access token            | *(required)*               |
| `-dir`   | Directory to clone into          | `.`                        |

### checkout-last-brach

Scans a directory of Git repositories and checks out the most recently committed branch in each one.

- Fetches latest remote info before determining the most recent branch
- Considers both local and remote branches
- Automatically creates a local tracking branch if the most recent branch only exists on the remote
- Skips non-Git directories and repos already on the correct branch

**Usage:**

```bash
cd checkout-last-brach && go build -o checkout-last-brach .

./checkout-last-brach ./repos
```

**Output example:**

```
[DONE] project-a — checked out to: feature-x (was: main)
[OK]   project-b — already on most recent branch: develop
[SKIP] not-a-repo — not a git repo
```

## Typical Workflow

```bash
# 1. Clone / fetch all repos in your GitLab group
./fetch-group -group my-team -token glpat-xxx -dir ./repos

# 2. Switch each repo to its most recently active branch
./checkout-last-brach ./repos
```

## Requirements

- Go 1.19+
- Git installed and available in `$PATH`
- A GitLab personal access token with `read_api` scope (for fetch-group)
