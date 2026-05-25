# command-context

`command-context` (`cmdctx`) is a lightweight Go library for building hierarchical CLI applications. Commands are organized in a tree; each node may have sub-commands. A single `Parse` call resolves the target command from the argument list and returns a fully-initialized `CommandContext` that is passed through every lifecycle callback.

## Installation

```sh
go get github.com/moisespsena-go/command-context
```

## Quick start

```go
package main

import (
    "fmt"
    "os"
    "net/http"
    "path/filepath"

    cc "github.com/moisespsena-go/command-context"
)

func main() {
    const bindKey = "bindAddr"

    root := &cc.Command{
        Name:        filepath.Base(os.Args[0]),
        Description: "runs the http server",
        New: func(ctx *cc.CommandContext) error {
            var bind string
            ctx.WithValue(bindKey, &bind)
            ctx.Flags().StringVar(&bind, "bind", "0.0.0.0:8000", "address to listen on")
            return nil
        },
        Run: func(ctx *cc.CommandContext) error {
            bind := *ctx.Value(bindKey).(*string)
            fmt.Fprintf(ctx.Out, "listening on %s\n", bind)

            mux := http.NewServeMux()
            mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
                fmt.Fprint(w, "Welcome to the github.com/moisespsena-go/command-context example page!")
            })
            return http.ListenAndServe(bind, mux)
        },
    }

    passwd := &cc.Command{
        Name:        "passwd",
        Description: "changes user password",
        Usage:       "USER_NAME",
        ParseArgs: func(ctx *cc.CommandContext) error {
            return ctx.Args.Eq(1) // require exactly one positional argument
        },
        Run: func(ctx *cc.CommandContext) error {
            userName := ctx.Args[0]
            fmt.Fprintf(ctx.Out, "changing password for %s\n", ctx.Args[0])
            return nil
        },
    }

    root.Sub(passwd)

    ctx, err := root.Parse(nil)
    if err != nil {
        fmt.Fprintln(os.Stderr, "ERROR:", err)
        os.Exit(1)
    }

    if err = ctx.Run(); err != nil {
        fmt.Fprintln(os.Stderr, "ERROR:", cc.ToContextExecuteError(ctx, err))
        os.Exit(1)
    }
}
```

Run a sub-command:

```sh
./myapp passwd alice
```

Print help:

```sh
./myapp help
./myapp --help
./myapp passwd --help
```

## Command lifecycle

For each invocation `Command.Parse` runs the following steps in order:

| Step | Callback | Purpose |
|------|----------|---------|
| 1 | `Command.New` | Register flags on `ctx.Flags()` before parsing |
| 2 | *(internal)* | Parse flags; handle `-OUT`, `-ERR`, `help`, `--help` |
| 3 | `Command.ParseArgs` | Validate / transform positional arguments |
| 4 | `CommandContext.Run` → `Command.Run` | Execute command logic |

## API reference

### `CommandContext`

The context object threaded through every callback.

| Field / Method | Description |
|----------------|-------------|
| `Out io.Writer` | Standard output (default `os.Stdout`; override with `-OUT <file>`) |
| `Err io.Writer` | Standard error (default `os.Stderr`; override with `-ERR <file>`) |
| `Context context.Context` | Embedded context for cancellation and value propagation |
| `InputArgs []string` | Raw arguments received before flag parsing |
| `Args Args` | Positional arguments remaining after flag parsing |
| `NamedArgs any` | Optional structured value populated in `ParseArgs` |
| `Name() string` | Name of the current command as invoked |
| `Path() []string` | Full path from root to the current command |
| `FullName() string` | Human-readable path with quoted names and `🠆` separators |
| `Parent() *CommandContext` | Parent context, or `nil` for the root |
| `Cmd() *Command` | The `Command` associated with this context |
| `Flags() *flag.FlagSet` | Flag set for registering and reading flags |
| `WithValue(name, value any) *CommandContext` | Store a value in the embedded context (chainable) |
| `Value(name any) any` | Retrieve a value from the embedded context |
| `Fork() *CommandContext` | Create a child context with the current one as parent |
| `Run() error` | Invoke `Command.Run`; no-op if `Run` is nil |
| `Help() error` | Write help text to `Err` |

### `Command`

Describes a command or sub-command.

| Field | Description |
|-------|-------------|
| `Name string` | Command name matched from parent's argument list |
| `Usage string` | One-line argument synopsis shown in help |
| `Description string` | Long-form description shown in help |
| `New func(*CommandContext) error` | Called before flag parsing; register flags here |
| `ParseArgs func(*CommandContext) error` | Called after flag parsing; validate positional args here |
| `Run func(*CommandContext) error` | Main logic; called by `CommandContext.Run` |
| `Help func(*CommandContext) error` | Extra help text; called by `Helper.Execute` |

| Method | Description |
|--------|-------------|
| `Sub(sub *Command) *Command` | Register a sub-command (chainable) |
| `SubCb(sub *Command, cb func(*Command)) *Command` | Register and configure a sub-command inline (chainable) |
| `Parse(ctx *CommandContext) (*CommandContext, error)` | Resolve and initialize the target command |

### `Args`

A `[]string` slice with count-validation helpers.

| Method | Description |
|--------|-------------|
| `Arg(i int) string` | Argument at index `i` (panics if out of range) |
| `Min(v int) error` | Error if count < `v` |
| `Eq(v int) error` | Error if count ≠ `v` |
| `Max(v int) error` | Error if count > `v` |
| `Range(min, max int) error` | Error if count outside `[min, max]` |
| `ShiftN(n int) ([]string, Args, error)` | Split into first `n` elements and the remainder |

### Error types

| Type / Function | Description |
|-----------------|-------------|
| `ContextParserError` | Error from the parse phase; includes command path and event name |
| `NewContextParserError(ctx, event, err)` | Construct a `ContextParserError` |
| `ToContextParserError(ctx, event, err)` | Wrap `err` unless it already is a `ContextParserError` |
| `ContextExecuteError` | Error from the execution phase; includes command path |
| `NewContextExecuteError(ctx, err)` | Construct a `ContextExecuteError` |
| `ToContextExecuteError(ctx, err)` | Wrap `err` unless it already is a `ContextExecuteError` |

Both error types implement `Unwrap()` for use with `errors.Is` / `errors.As`, and expose `Ctx() *CommandContext` to retrieve the failing command's context.

### `Helper`

Writes formatted help text to `CommandContext.Err`.

| Method | Description |
|--------|-------------|
| `Help(ctx) *Helper` | Constructor; sub-command list disabled by default |
| `SubCommands(bool) *Helper` | Toggle sub-command listing (chainable) |
| `WithSubCommands() *Helper` | Enable sub-command listing (chainable) |
| `Execute() error` | Write usage, description, sub-commands, then call `Command.Help` if set |

## Built-in flags

Every command automatically supports:

| Flag | Default | Description |
|------|---------|-------------|
| `-OUT <file>` | `-` (stdout) | Redirect `ctx.Out` to a file |
| `-ERR <file>` | `-` (stderr) | Redirect `ctx.Err` to a file |

## Built-in arguments

| Argument | Effect |
|----------|--------|
| `help` or `--help` | Print help with sub-command list and exit |
| `--` | Stop sub-command resolution; remaining args are positional |

## Example

A complete working example is available in the [`example/`](example/) directory. It demonstrates:

- Defining a root command and a `passwd` sub-command
- Validating argument count with `Args.Eq`
- Secure password input via [go-tty](https://github.com/mattn/go-tty)
- Wrapping execution errors with `ToContextExecuteError`
