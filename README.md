# arc-sdk

Shared SDK for the Arc toolkit.

## Packages

| Package | Description |
|---------|-------------|
| `ai` | AI client for Anthropic/OpenAI |
| `config` | Configuration loading and paths |
| `db` | SQLite database utilities |
| `db/migrations` | Database migration system |
| `errors` | CLI error types |
| `git` | Git operations |
| `output` | JSON/YAML/Table output formatting |
| `store` | Data stores (repos, sessions, deps, env) |
| `utils` | Path normalization, humanize |
| `version` | Version information |

## Installation

```bash
go get github.com/mtreilly/arc-sdk
```

## Usage

```go
import (
    "github.com/mtreilly/arc-sdk/config"
    "github.com/mtreilly/arc-sdk/db"
    "github.com/mtreilly/arc-sdk/output"
)

// Load configuration
cfg, err := config.Load()

// Open database
database, err := db.Open(db.DefaultDBPath())

// Format output
output.JSON(data)
```

## License

MIT
