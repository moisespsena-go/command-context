package cmdctx

import (
	"fmt"
	"slices"
	"sort"
	"strings"
)

type Helper struct {
	subCommands bool
	ctx         *CommandContext
}

func Help(ctx *CommandContext) *Helper {
	return &Helper{ctx: ctx, subCommands: false}
}

func (h *Helper) SubCommands(subCommands bool) *Helper {
	h.subCommands = subCommands
	return h
}

func (h *Helper) WithSubCommands() *Helper {
	return h.SubCommands(true)
}

func (h *Helper) Execute() (err error) {
	ctx := h.ctx
	help := ctx.cmd.Help

	var (
		names  = []string{ctx.name}
		parent = ctx.parent
	)

	for parent != nil {
		names = append(names, parent.Name())
		parent = parent.parent
	}

	slices.Reverse(names)

	fmt.Fprintf(ctx.Err, "Usage: %s %s\n", strings.Join(names, " "), ctx.cmd.Usage)

	if ctx.cmd.Description != "" {
		fmt.Fprintln(ctx.Err, ctx.cmd.Description)
	}

	if h.subCommands && len(ctx.cmd.sub) > 0 {
		fmt.Fprintln(ctx.Err)
		fmt.Fprintln(ctx.Err, "SUB COMMANDS:")

		var names []string
		for name := range ctx.cmd.sub {
			names = append(names, name)
		}

		sort.Strings(names)

		w, _ := consoleSize()
		if w == 0 {
			w = 100
		}

		for _, name := range names {
			sub := ctx.cmd.sub[name]

			help := []string{name + ": "}
			if sub.Description != "" {
				help = append(help, sub.Description)
			}

			lines := splitString(strings.Join(help, " "), w)
			fmt.Fprintln(ctx.Err, "\t"+lines[0])
			lines = lines[1:]
			for _, line := range lines {
				fmt.Fprintln(ctx.Err, "\t\t"+line)
			}

			fmt.Fprintln(ctx.Err)
		}
	}

	if help != nil {
		return help(ctx)
	}
	return nil
}
