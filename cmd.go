// Package cmdctx provides a hierarchical command framework with context propagation.
// Commands are organized in a tree where each node may have sub-commands.
// A [Command.Parse] call resolves the target command from the argument list and
// returns a fully-initialized [CommandContext] that is then passed to every
// lifecycle callback: New, ParseArgs, Run, and Help.
package cmdctx

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
)

// CommandContext carries state through the lifecycle of a command invocation.
// It holds I/O writers, the parsed argument list, named arguments, and an
// embedded [context.Context] for value propagation and cancellation.
// A CommandContext is created by [Command.Parse] and threaded through every
// callback (New, ParseArgs, Run, Help).
type CommandContext struct {
	// Out is the writer for standard output. Defaults to [os.Stdout].
	// Can be redirected to a file with the -OUT flag.
	Out,
	// Err is the writer for standard error. Defaults to [os.Stderr].
	// Can be redirected to a file with the -ERR flag.
	Err io.Writer
	parent    *CommandContext
	cmd       *Command
	flags     *flag.FlagSet
	name      string
	// Context is the standard-library context for cancellation and value propagation.
	Context   context.Context
	// InputArgs is the raw argument slice received by this command before flag parsing.
	InputArgs []string
	// Args holds positional arguments remaining after flag parsing.
	Args      Args
	// NamedArgs is an optional structured value populated by [Command.ParseArgs]; its type is caller-defined.
	NamedArgs any
}

// Name returns the name of the current command as it was invoked.
func (ctx *CommandContext) Name() string {
	return ctx.name
}

// Path returns the full invocation path from the root command to the current
// command as a slice of names (e.g. ["myapp", "users", "passwd"]).
func (ctx *CommandContext) Path() []string {
	names := []string{ctx.name}
	p := ctx.parent
	for p != nil {
		names = append(names, p.Name())
		p = p.parent
	}

	slices.Reverse(names)
	return names
}

// FullName returns the full invocation path with each segment quoted and
// separated by arrows (🠆), suitable for display in error messages.
func (ctx *CommandContext) FullName() string {
	pth := ctx.Path()
	for i, s := range pth {
		pth[i] = strconv.Quote(s)
	}
	return strings.Join(pth, " 🠆 ")
}

// Parent returns the parent CommandContext, or nil for the root command.
func (ctx *CommandContext) Parent() *CommandContext {
	return ctx.parent
}

// Cmd returns the [Command] associated with this context.
func (ctx *CommandContext) Cmd() *Command {
	return ctx.cmd
}

// Flags returns the [flag.FlagSet] used to parse this command's arguments.
// Register custom flags in [Command.New] before Parse calls flags.Parse.
func (ctx *CommandContext) Flags() *flag.FlagSet {
	return ctx.flags
}

// WithValue stores key/value in the embedded [context.Context] and returns ctx
// for chaining. Retrieve values with [CommandContext.Value].
func (ctx *CommandContext) WithValue(name, value any) *CommandContext {
	ctx.Context = context.WithValue(ctx.Context, name, value)
	return ctx
}

// Value retrieves a value from the embedded [context.Context] by key.
func (ctx *CommandContext) Value(name any) any {
	return ctx.Context.Value(name)
}

// Fork creates and returns a child CommandContext whose parent is ctx.
// The child inherits all fields; changes to slice/map fields after the fork
// are not reflected in the parent.
func (ctx *CommandContext) Fork() *CommandContext {
	child := *ctx
	child.parent = ctx
	return &child
}

// Run invokes [Command.Run] with this context.
// It returns nil without error when Run is not set.
func (ctx *CommandContext) Run() error {
	if ctx.cmd.Run == nil {
		return nil
	}

	return ctx.cmd.Run(ctx)
}

// Help writes help text for the current command to [CommandContext.Err].
func (ctx *CommandContext) Help() (err error) {
	return Help(ctx).Execute()
}

// Command describes a single command or sub-command in the hierarchy.
// Fields New, ParseArgs, Run, and Help are optional lifecycle callbacks
// invoked by [Command.Parse] and [CommandContext.Run].
type Command struct {
	// Name is the command name used to match this command from its parent's argument list.
	Name        string
	// Usage is a short one-line argument synopsis appended after the command path in help output.
	Usage       string
	// Description is the long-form description printed in help output.
	Description string
	sub         map[string]*Command
	// New is called before flag parsing. Register flags on [CommandContext.Flags] here.
	New         func(ctx *CommandContext) (err error)
	// Help is called by [Helper.Execute] after the auto-generated usage block.
	// Use it to print additional help text.
	Help        func(ctx *CommandContext) (err error)
	// Run is called by [CommandContext.Run] to execute the command's logic.
	Run         func(ctx *CommandContext) (err error)
	// ParseArgs is called after flag parsing to validate or transform positional arguments.
	// Populate [CommandContext.NamedArgs] here for typed access in Run.
	ParseArgs   func(ctx *CommandContext) (err error)
}

// Sub registers sub as a sub-command of b and returns b for chaining.
func (b *Command) Sub(sub *Command) *Command {
	if b.sub == nil {
		b.sub = make(map[string]*Command)
	}
	b.sub[sub.Name] = sub
	return b
}

// SubCb registers sub as a sub-command of b, calls cb with sub, then returns b
// for chaining. cb is a convenience hook for inline configuration of the sub-command.
func (b *Command) SubCb(sub *Command, cb func(sub *Command)) *Command {
	b.Sub(sub)
	cb(sub)
	return b
}

// Parse resolves the command to run from ctx.InputArgs and returns a populated
// CommandContext ready for [CommandContext.Run].
//
// If ctx is nil, Parse bootstraps a root context from [os.Args].
// The built-in flags -OUT and -ERR redirect [CommandContext.Out] and
// [CommandContext.Err] to the named files.
// The special arguments "help" and "--help" install a help-printing Run function.
// The sentinel "--" stops sub-command resolution; remaining args become positional.
//
// The returned context may point to a different *Command than b when a
// sub-command was matched.
func (b *Command) Parse(ctx *CommandContext) (_ *CommandContext, err error) {
	if ctx == nil {
		ctx = &CommandContext{
			name:      filepath.Base(os.Args[0]),
			InputArgs: os.Args[1:],
		}
	}

	if ctx.Context == nil {
		ctx.Context = context.Background()
	}

	if ctx.Out == nil {
		ctx.Out = os.Stdout
	}

	if ctx.Err == nil {
		ctx.Err = os.Stderr
	}

	dot := b

parse:
	ctx.cmd = dot
	ctx.flags = flag.NewFlagSet(ctx.name, flag.ContinueOnError)

	var defaultFlags struct {
		Stdout string
		Stderr string
	}

	ctx.flags.StringVar(&defaultFlags.Stdout, "OUT", "-", "program stdout file name")
	ctx.flags.StringVar(&defaultFlags.Stderr, "ERR", "-", "program stderr file name")

	if dot.New != nil {
		if err = dot.New(ctx); err != nil {
			err = ToContextParserError(ctx, "new", err)
			return
		}
	}

	if err = ctx.flags.Parse(ctx.InputArgs); err != nil {
		if err == flag.ErrHelp {
			helpCmd := *dot
			helpCmd.Run = func(ctx *CommandContext) (err error) {
				return Help(ctx).WithSubCommands().Execute()
			}
			ctx.cmd = &helpCmd
			return ctx, nil
		}
		err = ToContextParserError(ctx, "parse", err)
		return
	}

	ctx.Args = ctx.flags.Args()

	if dot.ParseArgs != nil {
		if err = dot.ParseArgs(ctx); err != nil {
			err = ToContextParserError(ctx, "parseArgs", err)
			return
		}
	}

	if ctx.Out, err = openWriterOrDefault(defaultFlags.Stdout, os.Stdout); err != nil {
		return
	}

	if ctx.Err, err = openWriterOrDefault(defaultFlags.Stderr, os.Stderr); err != nil {
		return
	}

	if len(ctx.Args) > 0 {
		subName := ctx.Args[0]
		switch subName {
		case "--":
			ctx.Args = ctx.Args[1:]
		case "help", "--help":
			helpCmd := *dot
			helpCmd.Run = func(ctx *CommandContext) (err error) {
				return Help(ctx).WithSubCommands().Execute()
			}
			ctx.cmd = &helpCmd
			return ctx, nil
		default:
			if len(dot.sub) > 0 {
				sub := dot.sub[subName]
				if sub == nil {
					err = ToContextParserError(ctx, "sub", fmt.Errorf("unknown command: %s", subName))
					return
				}
				ctx.Args = ctx.Args[1:]
				ctx = ctx.Fork()
				ctx.name = subName
				ctx.InputArgs = ctx.Args
				ctx.Args = nil
				dot = sub
				goto parse
			}
		}
	}

	return ctx, nil
}

func openWriter(name string) (f *os.File, err error) {
	if _, err = os.Stat(name); os.IsNotExist(err) {
		if f, err = os.Create(name); err != nil {
			return
		}
	} else if err == nil {
		return os.OpenFile(name, os.O_WRONLY, 0666)
	}
	return
}

func openWriterOrDefault(name string, defaultf *os.File) (f *os.File, err error) {
	switch name {
	case "", "-":
		return defaultf, nil
	default:
		return openWriter(name)
	}
}
