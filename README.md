# MultiKey CLI

MultiKey CLI is an open source developer tool that manages multiple GitHub SSH identities and applies the correct identity based on folder or repo location. It simplifies working with multiple GitHub accounts (personal, work, clients) by providing profile-based SSH routing tied to directory paths and repositories.

## Features

- **Multiple SSH Profile Management** - Create, edit, and delete SSH profiles for different GitHub accounts
- **Passphrase-Protected Keys** - Optionally encrypt generated SSH keys with a passphrase (with macOS Keychain integration)
- **Folder-to-Profile Mapping** - Automatically apply the correct profile based on folder location
- **Repo-Level Mapping** - Override profile per repository
- **Automatic Repo Scanning** - Find and analyze Git repositories
- **Auto-Fix** - Automatically update remote URLs and Git config to match profiles
- **Diagnostics** - Check repository alignment and configuration health
- **Guided Setup** - Interactive setup wizard for first-time users

## Installation

### Using Homebrew (Recommended)

```bash
brew tap Torres-Atlantic/multikey
brew trust Torres-Atlantic/multikey
brew install multikey
```

> **Note:** current Homebrew versions require `brew trust` before installing from a third-party tap. Run it once per machine.

### Using Go Install

```bash
go install github.com/Torres-Atlantic/multikey-cli/cmd/multikey@latest
```

### From Source

```bash
git clone https://github.com/Torres-Atlantic/multikey-cli.git
cd multikey-cli
make build
# Binary will be in build/multikey
```

## Quick Start

1. **Run setup** (first time only):
   ```bash
   multikey setup
   ```

2. **Create a profile**:
   ```bash
   multikey profile add
   ```

3. **Map a folder to a profile**:
   ```bash
   multikey map add ~/code/work --profile work
   ```

4. **Scan and fix repositories**:
   ```bash
   multikey scan ~/code/work
   multikey apply ~/code/work
   ```

## Commands

### Profile Management

- `multikey profile add` - Create a new SSH profile
- `multikey profile list` - List all profiles
- `multikey profile edit <id>` - Edit an existing profile
- `multikey profile remove <id>` - Remove a profile

### Mapping Management

- `multikey map add <path> --profile <id>` - Map a folder or repository to a profile
- `multikey map list` - List all mappings
- `multikey map remove <path>` - Remove a mapping

### Repository Operations

- `multikey scan <path>` - Scan for Git repositories and show alignment status
- `multikey apply <path>` - Automatically fix repositories to match their profiles (use `--dry-run` to preview changes without modifying anything)
- `multikey assign <path> --profile <id>` - Assign a profile to a repository and apply fixes
- `multikey status <path>` - Show repository status vs expected configuration
- `multikey diagnose <path>` - Run full health check

### Utility Commands

- `multikey setup` - Run guided setup wizard
- `multikey export` - Export configuration as JSON
- `multikey import <file>` - Import configuration from JSON file
- `multikey sponsor` - Show sponsor information
- `multikey version` - Show version information

## How It Works

MultiKey CLI works by:

1. **SSH Config Management**: Creates SSH host aliases (e.g., `github-work`) that route to different SSH keys
2. **Profile Mapping**: Maps folders and repositories to profiles
3. **Remote URL Rewriting**: Updates Git remote URLs to use the correct SSH host alias
4. **Git Config Updates**: Sets per-repository `user.email` and `user.name` to match the profile

### Example

If you have a profile `work` with SSH host `github-work`:

- Remote URL: `git@github.com:company/repo.git` → `git@github-work:company/repo.git`
- Git config: `user.email` → `work@company.com`, `user.name` → `work-username`

## Configuration

Configuration is stored at `~/.config/multikey/config.json`.

SSH configuration is written to `~/.ssh/multikey.conf` and included in `~/.ssh/config`.

## Development

### Building

```bash
make build          # Build for current platform
make release        # Build for all platforms
make install        # Install locally
make test           # Run tests
```

### Project Structure

```
multikey/
├── cmd/multikey/      # Main entry point
├── internal/
│   ├── cli/          # CLI commands (Cobra)
│   ├── config/       # Configuration management
│   ├── git/          # Git operations
│   ├── mapping/      # Profile mapping logic
│   ├── profile/      # Profile management
│   ├── scanner/      # Repository scanning
│   ├── ssh/          # SSH operations
│   └── diagnostics/  # Diagnostics
└── Makefile
```

## License

MultiKey CLI is available under the MIT License for individual use. This software is owned by **Torres Atlantic, LLC**.

### Individual Use (Free Trial)
MultiKey CLI is available as a **free trial with no end date** for individual developers. You are free to:
- Use the software for personal or individual projects
- Fork and modify the code
- Contribute improvements back to the project

All features are available without restrictions for individual use.

### Corporate License
For corporate or commercial use, please contact us at **mkc@torresatlantic.com** to discuss licensing options.

### Purchasing a License
If you'd like to support the project and purchase a license, visit [www.multikeycli.com](https://www.multikeycli.com) for more information.

See [LICENSE](LICENSE) file for full license details.

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## Support

For support, visit: [www.multikeycli.com](https://www.multikeycli.com)
