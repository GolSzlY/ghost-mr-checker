# Ghost MR Checker

A GitLab tool to detect "ghost" merge requests - MRs that were merged into branches but whose commits are actually missing from those branches.

## Overview

This tool helps identify merge requests that appear as "merged" in GitLab but whose commits are actually missing from the target branches. It checks both `release` and `master` branches to detect true ghost commits, accounting for squash merges and other merge strategies.

## Features

- ✅ Detects ghost MRs in both `release` and `master` branches
- ✅ Handles squash merges by checking commit titles/messages
- ✅ Configurable date range for checking
- ✅ YAML-based configuration
- ✅ Detailed status reporting

## Prerequisites

- **Go 1.20+**
- **GitLab Access Token** with `read_api` scope

## Installation

1. **Clone the repository:**
   ```bash
   git clone https://github.com/GolSzlY/ghost-mr-checker.git
   cd ghost-mr-checker
   ```

2. **Install dependencies:**
   ```bash
   go mod download
   ```

3. **Build the binary:**
   ```bash
   go build -o ghost-mr-checker
   ```

## Configuration

### Setup Config File

1. **Copy the example configuration:**
   ```bash
   cp config.yaml.example config.yaml
   ```

2. **Edit `config.yaml` with your GitLab details:**
   ```yaml
   gitlab:
     token: "glpat-YOUR_GITLAB_ACCESS_TOKEN_HERE"
     project_id: "YOUR_GITLAB_PROJECT_ID_HERE"
     url: "https://YOUR_GITLAB_URL_HERE/api/v4"
   check:
     since: "2025-11-09"
   ```

### Configuration Options

| Field | Description | Required | Default |
|-------|-------------|----------|---------|
| `gitlab.token` | Your GitLab personal access token | ✅ Yes | - |
| `gitlab.project_id` | GitLab project ID or URL-encoded path | ✅ Yes | - |
| `gitlab.url` | GitLab API base URL (e.g., `https://gitlab.com/api/v4`) | ✅ Yes | - |
| `check.since` | Start date for checking (YYYY-MM-DD) | No | `2025-11-09` |

### Getting a GitLab Access Token

1. Go to GitLab → **Settings** → **Access Tokens**
2. Create a new token with `read_api` scope
3. Copy the token and add it to your `config.yaml`

> **⚠️ Security Note:** Never commit `config.yaml` with real credentials to version control. The file is already in `.gitignore` to prevent accidental commits.

## Usage

### Basic Usage

```bash
./ghost-mr-checker
```

This will use the default `config.yaml` file in the current directory.

### Custom Config File

```bash
./ghost-mr-checker --config /path/to/custom-config.yaml
```

### Command-Line Flags

| Flag | Description | Default |
|------|-------------|---------|
| `--config` | Path to configuration file | `config.yaml` |

## How It Works

1. **Fetch Branch MRs**: Queries GitLab API for MRs merged into both `release` and `master` branches after the specified date
2. **Get Branch History**: Fetches the complete commit history of both branches
3. **Match Commits**: For each MR, checks if its commits actually exist in the target branch by:
   - Matching commit SHAs (for regular merges)
   - Matching commit titles/messages (for squash merges)
4. **Report Ghost Status**:
   - **[GHOST]**: MR marked as merged in GitLab but commits are missing from the target branch

## Example Output

```
[GHOST] Fix critical bug in payment processing (Branch: release, Source: bugfix/payment-issue)
[GHOST] Add new feature for user profiles (Branch: master, Source: feature/user-profiles)
[GHOST] Update dependencies (Branch: release, Source: chore/deps-update)
```

## Development

### Run Tests

```bash
go test ./...
```

### Run with Verbose Output

```bash
go run main.go --config config.yaml
```

## Troubleshooting

### "Failed to read config file"
- Ensure `config.yaml` exists in the current directory or specify the correct path with `--config`

### "GitLab token and project ID are required"
- Check that your `config.yaml` has both `gitlab.token` and `gitlab.project_id` filled in

### "Invalid date format"
- Ensure the `check.since` date is in `YYYY-MM-DD` format (e.g., `2025-11-09`)

## License

MIT

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.
