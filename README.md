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

Run a nested sub-command (sub-command with its own sub-commands):

```sh
./myapp user --role admin --dept eng add bob
./myapp user rm bob
```

When args do not match any sub-command, the parent's `Run` is called with the remaining positional args:

```sh
./myapp user --role admin --dept eng   # calls user.Run with empty Args
./myapp user unknown                   # calls user.Run with Args=["unknown"]
```

Print help:

```sh
./myapp help
./myapp --help
./myapp passwd --help
./myapp user --help
```

---

## Patterns

### Use a local struct for multiple flags

When a command accepts multiple flags, define a local struct, bind its fields in `New`, store the pointer via `ctx.WithValue`, and retrieve it in `Run` via `ctx.Value`. This avoids global variables and keeps flag definitions next to their usage.

```go
cmd := &cc.Command{
    Name:  "deploy",
    Usage: "--env NAME --branch BRANCH",
    New: func(ctx *cc.CommandContext) error {
        f := &struct {
            Env    string
            Branch string
        }{}
        flags := ctx.Flags()
        flags.StringVar(&f.Env, "env", "staging", "environment")
        flags.StringVar(&f.Branch, "branch", "main", "git branch")
        ctx.WithValue("flags", f)
        return nil
    },
    Run: func(ctx *cc.CommandContext) error {
        f := ctx.Value("flags").(*struct {
            Env    string
            Branch string
        })
        fmt.Fprintf(ctx.Out, "deploying %s to %s\n", f.Branch, f.Env)
        return nil
    },
}
```

The struct is scoped to the `New`/`Run` closure — no package-level state. Child contexts inherit the value through `Fork()` and can access the same pointer.

### Capture `ctx.Flags()` in a local variable

When registering multiple flags in the same `New` callback, assign `ctx.Flags()` to a local variable and call methods on it. This avoids repeated method calls and improves readability.

```go
cmd := &cc.Command{
    Name: "serve",
    New: func(ctx *cc.CommandContext) error {
        f := &struct {
            Verbose bool
            Level   string
        }{}
        flags := ctx.Flags()
        flags.BoolVar(&f.Verbose, "verbose", false, "enable verbose output")
        flags.StringVar(&f.Level, "level", "info", "log level")
        ctx.WithValue("flags", f)
        return nil
    },
    Run: func(ctx *cc.CommandContext) error {
        f := ctx.Value("flags").(*struct {
            Verbose bool
            Level   string
        })
        if f.Verbose {
            fmt.Fprintf(ctx.Out, "log level: %s\n", f.Level)
        }
        return nil
    },
}
```

## Command lifecycle

For each invocation `Command.Parse` runs the following steps in order:

| Step | Callback / phase | Purpose |
|------|------------------|---------|
| 1 | `Command.New` | Register custom flags on `ctx.Flags()` |
| 2 | *(internal)* | Parse flags; apply `-OUT`/`-ERR` redirects |
| 3 | `Command.ParseArgs` | Validate or transform `ctx.Args`; set `ctx.NamedArgs` |
| 4 | *(sub-command)* | If `ctx.Args[0]` matches a registered sub-command, **repeat from step 1** for that sub-command with the remaining args |
| 5 | `CommandContext.Run` → `Command.Run` | Execute command logic |

**Sub-command resolution in detail** (step 4):

1. If `ctx.Args` is empty → no resolution, move to step 5.
2. If `ctx.Args[0]` is `help` or `--help` → install a help-printing `Run` and return.
3. If `ctx.Args[0]` matches a sub-command name:
   - `Fork()` a child context with `InputArgs = ctx.Args[1:]`
   - Repeat the full lifecycle (steps 1–4) for the sub-command.
4. If `ctx.Args[0]` does **not** match any sub-command **and** the current command has `Run` set → `Run` is called with the full `ctx.Args` (including the unmatched argument).
5. If `ctx.Args[0]` does **not** match any sub-command **and** `Run` is nil → a `ContextParserError` ("unknown command") is returned.

---

## API reference

### `CommandContext`

The context object threaded through every callback. Created by `Command.Parse` and passed to `New`, `ParseArgs`, `Run`, and `Help`.

#### Fields

| Field | Type | Description |
|-------|------|-------------|
| `Out` | `io.Writer` | Standard output (default `os.Stdout`; redirect with `-OUT <file>`) |
| `Err` | `io.Writer` | Standard error (default `os.Stderr`; redirect with `-ERR <file>`) |
| `Context` | `context.Context` | Embedded Go context for cancellation and value propagation |
| `InputArgs` | `[]string` | Raw arguments received by this command **before** flag parsing |
| `Args` | `Args` | Positional arguments **after** flag parsing; modified during sub-command resolution |
| `NamedArgs` | `any` | Optional structured value—populate in `ParseArgs` for typed access in `Run` |

#### Methods

| Method | Returns | Description |
|--------|---------|-------------|
| `Name()` | `string` | Command name as invoked (e.g. `"passwd"`) |
| `Path()` | `[]string` | Full invocation path from root (e.g. `["myapp", "user", "add"]`) |
| `FullName()` | `string` | Quoted path joined with `🠆` for error messages (e.g. `"myapp" 🠆 "user" 🠆 "add"`) |
| `Parent()` | `*CommandContext` | Parent context, or `nil` for the root |
| `Cmd()` | `*Command` | The resolved `Command` struct for this context |
| `Flags()` | `*flag.FlagSet` | The flag set; register custom flags in `New`, read flag values in `Run` |
| `WithValue(name, value any)` | `*CommandContext` | Store a key/value in the embedded `Context` (chainable) |
| `Value(name any)` | `any` | Retrieve a value stored by `WithValue` |
| `Fork()` | `*CommandContext` | Create a child context (shallow copy with `parent` set to the receiver) |
| `Run()` | `error` | Call `Cmd().Run(ctx)`; returns nil when `Run` is nil |
| `Help()` | `error` | Write formatted help text to `Err` |

---

### `Command`

Describes a single command or sub-command in the tree. All callback fields are optional.

#### Fields

| Field | Signature | When called | Purpose |
|-------|-----------|-------------|---------|
| `Name` | `string` | — | Name used to match this command from the parent's positional args; must match `ctx.Args[0]` during sub-command resolution |
| `Usage` | `string` | — | Short argument synopsis shown in help (e.g. `"USER_NAME"`) |
| `Description` | `string` | — | Long description shown in help output |
| `New` | `func(*CommandContext) error` | Step 1 of Parse, before flag parsing | Register custom flags on `ctx.Flags()`; set up `ctx.WithValue`; fail parsing by returning an error |
| `ParseArgs` | `func(*CommandContext) error` | Step 3 of Parse, after flag parsing but before sub-command resolution | Validate `ctx.Args` counts/values; parse into `ctx.NamedArgs` |
| `Run` | `func(*CommandContext) error` | When `ctx.Run()` is called by the caller | Execute the command's business logic |
| `Help` | `func(*CommandContext) error` | At the end of `Helper.Execute` | Print additional help text after the auto-generated usage block |

**Important ordering note:** `ParseArgs` runs **before** sub-command resolution. This means:
- `ParseArgs` sees positional args that include the potential sub-command name.
- To allow sub-command matching, do not reject args that could be a sub-command name.
  A common pattern is to use `Args.Min(1)` (at least one required) instead of `Args.Eq(1)` (exactly one) when sub-commands exist.

#### Methods

| Method | Signature | Description |
|--------|-----------|-------------|
| `Sub` | `(sub *Command) *Command` | Register `sub` as a sub-command; returns the receiver for chaining |
| `SubCb` | `(sub *Command, cb func(*Command)) *Command` | Register `sub` and call `cb(sub)` for inline configuration; returns the receiver for chaining |
| `GetSub` | `(name string) *Command` | Look up a sub-command by name; returns nil if not found |
| `IsSub` | `(name string) bool` | Check whether a sub-command with the given name is registered |
| `Parse` | `(ctx *CommandContext) (*CommandContext, error)` | Resolve the target command, run lifecycle callbacks (New → ParseArgs), and return a context ready for `Run()`. Pass `nil` to bootstrap from `os.Args` |

**`Parse` behaviour summary:**

| Condition | Result |
|-----------|--------|
| `ctx` is `nil` | Bootstrap root context from `os.Args[0]` / `os.Args[1:]` |
| `ctx.InputArgs` is empty | Return context as-is; `Run` will be called with empty `Args` |
| First positional arg is `help` or `--help` | Install a help-printing `Run` on a copy of the current command |
| First positional arg matches a sub-command | Fork context, shift args, **repeat lifecycle** for the sub-command |
| First positional arg does **not** match; `Run` is set | Return context — `Run` receives all positional args (including the unmatched name) |
| First positional arg does **not** match; `Run` is nil | Return `ContextParserError` with event `"sub"` |
| `--OUT` / `--ERR` flags | Redirect `ctx.Out` / `ctx.Err` to the specified files |

---

### `Args`

A `[]string` slice with validation helpers for positional argument counts. Methods return descriptive errors that callers may check or display.

| Method | Signature | Behaviour |
|--------|-----------|-----------|
| `Arg` | `(i int) string` | Return argument at index `i`; **panics** if `i` is out of range |
| `Min` | `(v int) error` | Error if `len(args) < v` |
| `Max` | `(v int) error` | Error if `len(args) > v` |
| `Eq` | `(v int) error` | Error if `len(args) != v` |
| `Range` | `(min, max int) error` | Error if `len(args)` outside `[min, max]` (delegates to `Min` + `Max`) |
| `ShiftN` | `(n int) ([]string, Args, error)` | Split first `n` elements from the rest; error if fewer than `n` available |

---

### Error types

Two concrete error types distinguish parse-phase failures from execution-phase failures. Both implement `Unwrap()` for `errors.Is` / `errors.As` and expose `Ctx()` to retrieve the failing context.

#### `ContextParserError`

Returned when a command fails during `Parse` (steps 1–4 of the lifecycle). Includes which lifecycle event (`"new"`, `"parse"`, `"parseArgs"`, or `"sub"`) caused the failure.

| Constructor / Function | Signature | Behaviour |
|------------------------|-----------|-----------|
| `NewContextParserError` | `(ctx *CommandContext, event string, err error) *ContextParserError` | Create a new parser error |
| `ToContextParserError` | `(ctx *CommandContext, event string, err error) *ContextParserError` | Wrap `err`; returns as-is if `err` already is a `ContextParserError` |

| Method | Returns | Description |
|--------|---------|-------------|
| `Error()` | `string` | Format: `parse command "root" 🠆 "sub" failed: (parseArgs event) <msg>` |
| `Unwrap()` | `error` | Underlying error for `errors.Is` / `errors.As` |
| `Ctx()` | `*CommandContext` | The context of the command that failed |

#### `ContextExecuteError`

Returned when a command fails during `Run` (step 5). Typically created by the caller with `ToContextExecuteError` when wrapping a `Run` error.

| Constructor / Function | Signature | Behaviour |
|------------------------|-----------|-----------|
| `NewContextExecuteError` | `(ctx *CommandContext, err error) *ContextExecuteError` | Create a new execution error |
| `ToContextExecuteError` | `(ctx *CommandContext, err error) *ContextExecuteError` | Wrap `err`; returns as-is if `err` already is a `ContextExecuteError` |

| Method | Returns | Description |
|--------|---------|-------------|
| `Error()` | `string` | Format: `execute command "root" 🠆 "sub" failed: <msg>` |
| `Unwrap()` | `error` | Underlying error for `errors.Is` / `errors.As` |
| `Ctx()` | `*CommandContext` | The context of the command that failed |

#### Typical error handling

```go
ctx, err := root.Parse(nil)
if err != nil {
    var pe *cc.ContextParserError
    if errors.As(err, &pe) {
        fmt.Fprintf(os.Stderr, "parse error in %s: %s\n", pe.Ctx().FullName(), pe.Unwrap())
    } else {
        fmt.Fprintln(os.Stderr, "ERROR:", err)
    }
    os.Exit(1)
}

if err = ctx.Run(); err != nil {
    fmt.Fprintln(os.Stderr, "ERROR:", cc.ToContextExecuteError(ctx, err))
    os.Exit(1)
}
```

---

### `Helper`

Writes formatted help text to `CommandContext.Err`. Each command receives help automatically via the built-in `help` / `--help` arguments.

| Method | Signature | Description |
|--------|-----------|-------------|
| `Help` | `(ctx *CommandContext) *Helper` | Constructor; sub-command listing is disabled by default |
| `SubCommands` | `(v bool) *Helper` | Enable or disable the sub-command list in the output (chainable) |
| `WithSubCommands` | `() *Helper` | Shorthand for `SubCommands(true)` (chainable) |
| `Execute` | `() error` | Write the full help page: usage, description, sub-commands (if enabled), then call `Command.Help` if set |

**Output order produced by `Execute`:**

1. `Usage: <path> [flags] <Usage>`
2. Command `Description`
3. Sub-command list (if enabled via `WithSubCommands()`)
4. Content from `Command.Help` callback (if set)

---

## Built-in flags

Every command automatically supports these flags (registered in step 2 of the lifecycle):

| Flag | Default | Description |
|------|---------|-------------|
| `-OUT <file>` | `-` (stdout) | Redirect `ctx.Out` to the named file |
| `-ERR <file>` | `-` (stderr) | Redirect `ctx.Err` to the named file |

## Built-in arguments

These positional arguments are intercepted during sub-command resolution (step 4):

| Argument | Effect |
|----------|--------|
| `help` or `--help` | Install a help-printing `Run`; the command's own `Run` is **not** called |
| `--` | Stop sub-command resolution — all remaining arguments become positional for the current command |

## Example

A complete working example is available in the [`example/`](example/) directory. It demonstrates:

- Defining a root command and a `passwd` sub-command
- Nested sub-commands: the `user` command has its own sub-commands (`add`, `rm`) and a `Run` fallback when the arguments do not match a sub-command
- Registering custom flags with `Command.New`
- Typed flag struct: the `user` command binds multiple flags to a local `flags` struct in `New`, stores it via `ctx.WithValue`, and retrieves it in `Run` via `ctx.Value`
- Validating argument counts with `Args.Eq`
- Secure password input via [go-tty](https://github.com/mattn/go-tty)
- Wrapping execution errors with `ToContextExecuteError`
