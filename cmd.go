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

type CommandContext struct {
	Out,
	Err io.Writer
	parent    *CommandContext
	cmd       *Command
	flags     *flag.FlagSet
	name      string
	Context   context.Context
	InputArgs []string
	Args      Args
	NamedArgs any
}

func (ctx *CommandContext) Name() string {
	return ctx.name
}

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

func (ctx *CommandContext) FullName() string {
	pth := ctx.Path()
	for i, s := range pth {
		pth[i] = strconv.Quote(s)
	}
	return strings.Join(pth, " 🠆 ")
}

func (ctx *CommandContext) Parent() *CommandContext {
	return ctx.parent
}

func (ctx *CommandContext) Cmd() *Command {
	return ctx.cmd
}

func (ctx *CommandContext) Flags() *flag.FlagSet {
	return ctx.flags
}

func (ctx *CommandContext) WithValue(name, value any) *CommandContext {
	ctx.Context = context.WithValue(ctx.Context, name, value)
	return ctx
}

func (ctx *CommandContext) Value(name any) any {
	return ctx.Context.Value(name)
}

func (ctx *CommandContext) Fork() *CommandContext {
	child := *ctx
	child.parent = ctx
	return &child
}

func (ctx *CommandContext) Run() error {
	if ctx.cmd.Run == nil {
		return nil
	}

	return ctx.cmd.Run(ctx)
}

func (ctx *CommandContext) Help() (err error) {
	return Help(ctx).Execute()
}

type Command struct {
	Name        string
	Usage       string
	Description string
	sub         map[string]*Command
	New         func(ctx *CommandContext) (err error)
	Help        func(ctx *CommandContext) (err error)
	Run         func(ctx *CommandContext) (err error)
	ParseArgs   func(ctx *CommandContext) (err error)
}

func (b *Command) Sub(sub *Command) *Command {
	if b.sub == nil {
		b.sub = make(map[string]*Command)
	}
	b.sub[sub.Name] = sub
	return b
}

func (b *Command) SubCb(sub *Command, cb func(sub *Command)) *Command {
	b.Sub(sub)
	cb(sub)
	return b
}

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
