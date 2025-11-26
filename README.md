# Ghost MR Checker

A GitLab tool to detect "ghost" merge requests - MRs that were merged into the `release` branch but are missing from the `master` branch.

## Overview

This tool helps identify merge requests that have been merged into your `release` branch but haven't been properly merged into `master`. It checks the actual commit history to detect true ghost commits, accounting for squash merges and other merge strategies.

## Features

- ✅ Detects MRs merged to `release` but missing from `master`
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

1. **Fetch Release MRs**: Queries GitLab API for MRs merged into `release` after the specified date
2. **Get Release Branch History**: Fetches the complete commit history of the `release` branch
3. **Match Commits**: For each MR, checks if its commits exist in the release branch by:
   - Matching commit SHAs (for regular merges)
   - Matching commit titles/messages (for squash merges)
4. **Check Master Status**: For MRs found in release, checks if they exist in `master`
5. **Report Status**:
   - **[MISSING]**: MR merged to release but no corresponding MR to master exists
   - **[OPEN]**: MR to master exists but is still open
   - **[CLOSED]**: MR to master exists but was closed without merging
   - **[MERGED]**: MR properly merged to master (not shown in output)

## Example Output

```
[MISSING] Fix critical bug in payment processing (Source: bugfix/payment-issue)
[OPEN] Add new feature for user profiles (Source: feature/user-profiles)
[CLOSED] Update dependencies (Source: chore/deps-update)
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
