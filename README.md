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
| `store` | Data stores (repos, sessions, deps, env) and a generic key-value store |
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

## Key-Value Store

The `store` package provides a simple `KVStore` interface for optional module state persistence, using SQLite by default with automatic in-memory fallback.

```go
import (
    "context"
    "github.com/mtreilly/arc-sdk/store"
)

// Open persistent store (uses default arc.db location)
kv, err := store.OpenSQLiteStore("")
if err != nil {
    // Fallback: in-memory store (stateless mode)
    kv = store.NewMemoryStore()
}
defer kv.Close()

// Define typed state
type State struct {
    Counter int    `json:"counter"`
    Updated string `json:"updated"`
}

// Save state
ctx := context.Background()
store.SetJSON(ctx, kv, "my-module:state", State{Counter: 1, Updated: "now"})

// Load state
var s State
if _, err := store.GetJSON[State](ctx, kv, "my-module:state"); err != nil {
    // handle missing state (store.ErrNotFound) or other errors
}
```

## License

MIT
