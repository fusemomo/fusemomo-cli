# FuseMomo CLI

[![Build Status](https://github.com/fusemomo/fusemomo-cli/actions/workflows/release.yaml/badge.svg)](https://github.com/fusemomo/fusemomo-cli/actions)
[![Go Version](https://img.shields.io/github/go-mod/go-version/fusemomo/fusemomo-cli)](https://go.dev/)
[![License](https://img.shields.io/github/license/fusemomo/fusemomo-cli)](LICENSE)

**FuseMomo CLI** is a command-line interface designed for AI agents to interact with the [FuseMomo](https://fusemomo.com).

> [!NOTE]
> This CLI is optimized for machine-to-machine interaction. It defaults to structured JSON output and provides deterministic exit codes for use in automated pipelines.

---

## 🚀 Key Features

- **Agent-First Design**: Pure JSON output (standard), TTY-aware formatting, and comprehensive error codes.
- **Interactive REPL**: A built-in `prompt` mode for exploratory use and manual debugging.
- **L1 — Identity Resolution**: Resolve and link disparate identifiers into a single canonical entity.
- **L2 — Behavioral Graph**: Log and track immutable interaction events across your agent ecosystem.
- **L3 — Behavioral Intelligence**: Retrieve success-optimized recommendations for next-best-actions.

---

## 📦 Installation

### Quick Install (One-liner)
Install the latest pre-compiled binary directly to `/usr/local/bin`:

```bash
curl -sL https://raw.githubusercontent.com/fusemomo/fusemomo-cli/main/scripts/install.sh | sudo bash
```

### From Source
Requires Go 1.25+.

```bash
git clone https://github.com/fusemomo/fusemomo-cli.git
cd fusemomo-cli
make build
# Binary will be in bin/fusemomo
```

### Direct Install
```bash
go install github.com/fusemomo/fusemomo-cli@latest
```

---

## 🔧 Configuration

Run the interactive setup to configure your environment:

```bash
fusemomo setup
```

This will prompt for your API key and base URL, saving them to `~/.fusemomo/config.yaml`.

### Priority Order
The CLI resolves configuration in the following order:
1.  **Flags**: `--api-key`, `--api-url`
2.  **Environment Variables**: `FUSEMOMO_API_KEY`, `FUSEMOMO_API_URL`
3.  **Config File**: `~/.fusemomo/config.yaml`

---

## 📖 Usage

### Core Commands

| Command | Description |
|---|---|
| `fusemomo setup` | One-time interactive configuration |
| `fusemomo entity` | Manage and resolve behavioral entities |
| `fusemomo interaction` | Log single or batch interaction events |
| `fusemomo recommend` | Get and track outcome of recommendations |
| `fusemomo prompt` | Launch the interactive REPL |
| `fusemomo version` | Display build information |

### Examples

**Resolve an Entity**
```bash
fusemomo entity resolve --id email=user@example.com --id stripe_id=cus_123
```

**Log an Interaction**
```bash
fusemomo interaction log --entity-id <uuid> --type "purchase" --intent "buy_premium"
```

**Get a Recommendation**
```bash
fusemomo recommend get --entity-id <uuid> --strategy "conversion_optimizer"
```

### Output Formats
Control output using the `--output` global flag:
- `json` (Default): Minified JSON for pipes, pretty-printed for TTY.
- `table`: Human-readable colored tables for manual inspection.

```bash
fusemomo entity list --output table
```

---

## 🛠️ Development

The project includes a `Makefile` for standard development tasks:

| Target | Description |
|---|---|
| `make build` | Compile the binary to `bin/fusemomo` |
| `make test` | Run all tests with race detection |
| `make lint` | Run golangci-lint check |
| `make install` | Install binary to `$GOPATH/bin` |
| `make clean` | Remove build artifacts |

### Release Process
This project uses **GoReleaser** for automated cross-platform builds.
To create a new release:
1.  Push a new tag: `git tag -a v1.0.0 -m "Release v1.0.0" && git push origin v1.0.0`
2.  GitHub Actions will automatically build and publish the release.

---

## 🛡️ Security

- **API Keys**: Never logged or stored in history.
- **Permissions**: The config file is created with `0600` permissions (owner read/write only).
- **Transport**: TLS is enforced by default (`https://`). Using `http://` will trigger a warning.

---

© 2026 FuseMomo Team.
