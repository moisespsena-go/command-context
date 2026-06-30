package cmdctx

import (
	"errors"
	"testing"
)

func TestParse_RunWithSubcmd_UnknownArg_CallsRunWithArgs(t *testing.T) {
	var runArgs Args
	root := &Command{
		Name: "root",
		Run: func(ctx *CommandContext) error {
			runArgs = append(Args{}, ctx.Args...)
			return nil
		},
	}
	root.Sub(&Command{Name: "valid"})

	ctx, err := root.Parse(&CommandContext{
		name:      "root",
		InputArgs: []string{"unknown", "extra1", "extra2"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if err = ctx.Run(); err != nil {
		t.Fatal(err)
	}
	want := Args{"unknown", "extra1", "extra2"}
	if len(runArgs) != len(want) {
		t.Fatalf("Args = %v, want %v", runArgs, want)
	}
	for i := range want {
		if runArgs[i] != want[i] {
			t.Fatalf("Args[%d] = %q, want %q", i, runArgs[i], want[i])
		}
	}
}

func TestParse_RunWithSubcmd_UnknownArg_WithFlags(t *testing.T) {
	var runArgs Args
	root := &Command{
		Name: "root",
		Run: func(ctx *CommandContext) error {
			runArgs = append(Args{}, ctx.Args...)
			return nil
		},
	}
	root.Sub(&Command{Name: "valid"})

	ctx, err := root.Parse(&CommandContext{
		name:      "root",
		InputArgs: []string{"--OUT", "/dev/null", "unknown"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if err = ctx.Run(); err != nil {
		t.Fatal(err)
	}
	want := Args{"unknown"}
	if len(runArgs) != len(want) || runArgs[0] != want[0] {
		t.Fatalf("Args = %v, want %v", runArgs, want)
	}
}

func TestParse_RunWithSubcmd_EmptyArgs_CallsRun(t *testing.T) {
	var runCalled bool
	root := &Command{
		Name: "root",
		Run: func(ctx *CommandContext) error {
			runCalled = true
			return nil
		},
	}
	root.Sub(&Command{Name: "valid"})

	ctx, err := root.Parse(&CommandContext{
		name:      "root",
		InputArgs: nil,
	})
	if err != nil {
		t.Fatal(err)
	}
	if err = ctx.Run(); err != nil {
		t.Fatal(err)
	}
	if !runCalled {
		t.Fatal("Run not called")
	}
}

func TestParse_RunWithSubcmd_KnownArg_RunsSubcmd(t *testing.T) {
	var rootRun, subRun bool
	root := &Command{
		Name: "root",
		Run: func(ctx *CommandContext) error {
			rootRun = true
			return nil
		},
	}
	root.Sub(&Command{
		Name: "valid",
		Run: func(ctx *CommandContext) error {
			subRun = true
			return nil
		},
	})

	ctx, err := root.Parse(&CommandContext{
		name:      "root",
		InputArgs: []string{"valid"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if err = ctx.Run(); err != nil {
		t.Fatal(err)
	}
	if rootRun {
		t.Fatal("root Run should not be called when subcommand matches")
	}
	if !subRun {
		t.Fatal("sub Run not called")
	}
}

func TestParse_SubcmdWithSubcmd_UnknownArg_WithFlags(t *testing.T) {
	var runArgs Args
	var flagVal string
	sub := &Command{
		Name: "sub",
		New: func(ctx *CommandContext) error {
			ctx.Flags().StringVar(&flagVal, "name", "", "test flag")
			return nil
		},
		Run: func(ctx *CommandContext) error {
			runArgs = append(Args{}, ctx.Args...)
			return nil
		},
	}
	sub.Sub(&Command{Name: "valid-subsub"})

	root := &Command{Name: "root"}
	root.Sub(sub)

	ctx, err := root.Parse(&CommandContext{
		name:      "root",
		InputArgs: []string{"sub", "--name", "foo", "unknown"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if flagVal != "foo" {
		t.Fatalf("flag --name = %q, want %q", flagVal, "foo")
	}
	if err = ctx.Run(); err != nil {
		t.Fatal(err)
	}
	want := Args{"unknown"}
	if len(runArgs) != len(want) || runArgs[0] != want[0] {
		t.Fatalf("Args = %v, want %v", runArgs, want)
	}
}

func TestParse_NoRunWithSubcmd_UnknownArg_Error(t *testing.T) {
	root := &Command{
		Name: "root",
	}
	root.Sub(&Command{Name: "valid"})

	_, err := root.Parse(&CommandContext{
		name:      "root",
		InputArgs: []string{"unknown"},
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	var parserErr *ContextParserError
	if !errors.As(err, &parserErr) {
		t.Fatalf("expected *ContextParserError, got %T", err)
	}
}

func TestParse_HelpArg_InstallsHelpRun(t *testing.T) {
	root := &Command{
		Name: "root",
		Run: func(ctx *CommandContext) error {
			t.Fatal("Run should not be called")
			return nil
		},
	}
	root.Sub(&Command{Name: "valid"})

	ctx, err := root.Parse(&CommandContext{
		name:      "root",
		InputArgs: []string{"help"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if ctx.cmd.Run == nil {
		t.Fatal("help Run should not be nil")
	}
}

func TestParse_HelpFlag_InstallsHelpRun(t *testing.T) {
	root := &Command{
		Name: "root",
		Run: func(ctx *CommandContext) error {
			t.Fatal("Run should not be called")
			return nil
		},
	}
	root.Sub(&Command{Name: "valid"})

	ctx, err := root.Parse(&CommandContext{
		name:      "root",
		InputArgs: []string{"--help"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if ctx.cmd.Run == nil {
		t.Fatal("help Run should not be nil")
	}
}

func TestParse_RunNoSubcmd_Args(t *testing.T) {
	var runArgs Args
	root := &Command{
		Name: "root",
		Run: func(ctx *CommandContext) error {
			runArgs = append(Args{}, ctx.Args...)
			return nil
		},
	}

	ctx, err := root.Parse(&CommandContext{
		name:      "root",
		InputArgs: []string{"arg1", "arg2"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if err = ctx.Run(); err != nil {
		t.Fatal(err)
	}
	want := Args{"arg1", "arg2"}
	if len(runArgs) != len(want) || runArgs[0] != want[0] || runArgs[1] != want[1] {
		t.Fatalf("Args = %v, want %v", runArgs, want)
	}
}

func TestParse_RunNoSubcmd_EmptyArgs(t *testing.T) {
	var runCalled bool
	root := &Command{
		Name: "root",
		Run: func(ctx *CommandContext) error {
			runCalled = true
			return nil
		},
	}

	ctx, err := root.Parse(&CommandContext{
		name:      "root",
		InputArgs: nil,
	})
	if err != nil {
		t.Fatal(err)
	}
	if err = ctx.Run(); err != nil {
		t.Fatal(err)
	}
	if !runCalled {
		t.Fatal("Run not called")
	}
}

func TestParse_NoRunNoSubcmd_Nop(t *testing.T) {
	root := &Command{Name: "root"}

	ctx, err := root.Parse(&CommandContext{
		name:      "root",
		InputArgs: []string{"anything"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if err = ctx.Run(); err != nil {
		t.Fatal(err)
	}
}

func TestParse_DeepNesting_KnownPath_RunsLeaf(t *testing.T) {
	var ran []string
	makeRun := func(name string) func(*CommandContext) error {
		return func(ctx *CommandContext) error {
			ran = append(ran, name)
			return nil
		}
	}

	leaf := &Command{Name: "leaf", Run: makeRun("leaf")}
	mid := &Command{Name: "mid", Run: makeRun("mid")}
	mid.Sub(leaf)
	root := &Command{Name: "root", Run: makeRun("root")}
	root.Sub(mid)

	ctx, err := root.Parse(&CommandContext{
		name:      "root",
		InputArgs: []string{"mid", "leaf"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if err = ctx.Run(); err != nil {
		t.Fatal(err)
	}
	if len(ran) != 1 || ran[0] != "leaf" {
		t.Fatalf("ran %v, want [leaf]", ran)
	}
}

func TestParse_DeepNesting_UnknownSubsub_RunsMid(t *testing.T) {
	var ran []string
	mid := &Command{
		Name: "mid",
		Run: func(ctx *CommandContext) error {
			ran = append(ran, "mid")
			return nil
		},
	}
	mid.Sub(&Command{Name: "leaf"})

	root := &Command{Name: "root"}
	root.Sub(mid)

	ctx, err := root.Parse(&CommandContext{
		name:      "root",
		InputArgs: []string{"mid", "unknown"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if err = ctx.Run(); err != nil {
		t.Fatal(err)
	}
	if len(ran) != 1 || ran[0] != "mid" {
		t.Fatalf("ran %v, want [mid]", ran)
	}
}

func TestParse_Subcmd_ParseArgsCalled(t *testing.T) {
	var parseArgsCalled bool
	sub := &Command{
		Name: "sub",
		ParseArgs: func(ctx *CommandContext) error {
			parseArgsCalled = true
			return nil
		},
		Run: func(ctx *CommandContext) error {
			return nil
		},
	}
	root := &Command{Name: "root"}
	root.Sub(sub)

	ctx, err := root.Parse(&CommandContext{
		name:      "root",
		InputArgs: []string{"sub"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !parseArgsCalled {
		t.Fatal("sub ParseArgs not called")
	}
	if err = ctx.Run(); err != nil {
		t.Fatal(err)
	}
}

func TestParse_Subcmd_ParseArgsError(t *testing.T) {
	sub := &Command{
		Name: "sub",
		ParseArgs: func(ctx *CommandContext) error {
			return errors.New("invalid args")
		},
	}
	root := &Command{Name: "root"}
	root.Sub(sub)

	_, err := root.Parse(&CommandContext{
		name:      "root",
		InputArgs: []string{"sub", "something"},
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	var parserErr *ContextParserError
	if !errors.As(err, &parserErr) {
		t.Fatalf("expected *ContextParserError, got %T", err)
	}
}

func TestParse_NewError(t *testing.T) {
	root := &Command{
		Name: "root",
		New: func(ctx *CommandContext) error {
			return errors.New("new failed")
		},
	}

	_, err := root.Parse(&CommandContext{
		name:      "root",
		InputArgs: nil,
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	var parserErr *ContextParserError
	if !errors.As(err, &parserErr) {
		t.Fatalf("expected *ContextParserError, got %T", err)
	}
}

func TestParse_UnknownFlag_Error(t *testing.T) {
	root := &Command{
		Name: "root",
		Run: func(ctx *CommandContext) error {
			return nil
		},
	}

	_, err := root.Parse(&CommandContext{
		name:      "root",
		InputArgs: []string{"--undefined-flag"},
	})
	if err == nil {
		t.Fatal("expected error for undefined flag")
	}
}

func TestParse_UnknownFlag_ReturnsContextParserError(t *testing.T) {
	root := &Command{
		Name: "root",
		Run: func(ctx *CommandContext) error {
			return nil
		},
	}

	_, err := root.Parse(&CommandContext{
		name:      "root",
		InputArgs: []string{"--undefined-flag"},
	})
	if err == nil {
		t.Fatal("expected error")
	}
	var parserErr *ContextParserError
	if !errors.As(err, &parserErr) {
		t.Fatalf("expected *ContextParserError, got %T", err)
	}
}

func TestParse_FlagAfterSubcmdName(t *testing.T) {
	var runArgs Args
	sub := &Command{
		Name: "sub",
		New: func(ctx *CommandContext) error {
			ctx.Flags().String("name", "default", "name flag")
			return nil
		},
		Run: func(ctx *CommandContext) error {
			runArgs = append(Args{}, ctx.Args...)
			return nil
		},
	}
	sub.Sub(&Command{Name: "valid-subsub"})

	root := &Command{Name: "root"}
	root.Sub(sub)

	ctx, err := root.Parse(&CommandContext{
		name:      "root",
		InputArgs: []string{"sub", "--name", "foo", "unknown"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if err = ctx.Run(); err != nil {
		t.Fatal(err)
	}
	want := Args{"unknown"}
	if len(runArgs) != len(want) || runArgs[0] != want[0] {
		t.Fatalf("Args = %v, want %v", runArgs, want)
	}
}

func TestParse_SubcmdReceivesExtraArgs(t *testing.T) {
	var subArgs Args
	sub := &Command{
		Name: "sub",
		Run: func(ctx *CommandContext) error {
			subArgs = append(Args{}, ctx.Args...)
			return nil
		},
	}
	root := &Command{Name: "root"}
	root.Sub(sub)

	ctx, err := root.Parse(&CommandContext{
		name:      "root",
		InputArgs: []string{"sub", "extra1", "extra2"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if err = ctx.Run(); err != nil {
		t.Fatal(err)
	}
	want := Args{"extra1", "extra2"}
	if len(subArgs) != len(want) || subArgs[0] != want[0] || subArgs[1] != want[1] {
		t.Fatalf("Args = %v, want %v", subArgs, want)
	}
}

func TestParse_SubcmdMultipleChoices(t *testing.T) {
	var ran string
	root := &Command{Name: "root"}
	root.Sub(&Command{
		Name: "foo",
		Run: func(ctx *CommandContext) error {
			ran = "foo"
			return nil
		},
	})
	root.Sub(&Command{
		Name: "bar",
		Run: func(ctx *CommandContext) error {
			ran = "bar"
			return nil
		},
	})

	ctx, err := root.Parse(&CommandContext{
		name:      "root",
		InputArgs: []string{"bar"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if err = ctx.Run(); err != nil {
		t.Fatal(err)
	}
	if ran != "bar" {
		t.Fatalf("ran %q, want bar", ran)
	}
}

func TestParse_CustomContextDefaults(t *testing.T) {
	root := &Command{
		Name: "root",
		Run: func(ctx *CommandContext) error {
			return nil
		},
	}

	ctx, err := root.Parse(&CommandContext{
		name:      "root",
		InputArgs: nil,
	})
	if err != nil {
		t.Fatal(err)
	}
	if ctx.Context == nil {
		t.Fatal("Context should not be nil")
	}
	if ctx.Out == nil {
		t.Fatal("Out should not be nil")
	}
	if ctx.Err == nil {
		t.Fatal("Err should not be nil")
	}
	if ctx.name != "root" {
		t.Fatalf("name = %q, want root", ctx.name)
	}
}

func TestCommandContext_Path(t *testing.T) {
	leaf := &Command{Name: "leaf", Run: func(ctx *CommandContext) error { return nil }}
	mid := &Command{Name: "mid"}
	mid.Sub(leaf)
	root := &Command{Name: "root"}
	root.Sub(mid)

	ctx, err := root.Parse(&CommandContext{
		name:      "root",
		InputArgs: []string{"mid", "leaf"},
	})
	if err != nil {
		t.Fatal(err)
	}

	path := ctx.Path()
	want := []string{"root", "mid", "leaf"}
	if len(path) != len(want) {
		t.Fatalf("Path = %v, want %v", path, want)
	}
	for i := range want {
		if path[i] != want[i] {
			t.Fatalf("Path[%d] = %q, want %q", i, path[i], want[i])
		}
	}
}

func TestCommandContext_Parent(t *testing.T) {
	leaf := &Command{Name: "leaf", Run: func(ctx *CommandContext) error { return nil }}
	mid := &Command{Name: "mid"}
	mid.Sub(leaf)
	root := &Command{Name: "root"}
	root.Sub(mid)

	ctx, err := root.Parse(&CommandContext{
		name:      "root",
		InputArgs: []string{"mid", "leaf"},
	})
	if err != nil {
		t.Fatal(err)
	}

	if ctx.Parent() == nil {
		t.Fatal("leaf should have a parent")
	}
	if ctx.Parent().Name() != "mid" {
		t.Fatalf("leaf parent name = %q, want mid", ctx.Parent().Name())
	}
	if ctx.Parent().Parent() == nil {
		t.Fatal("mid should have a parent")
	}
	if ctx.Parent().Parent().Name() != "root" {
		t.Fatalf("mid parent name = %q, want root", ctx.Parent().Parent().Name())
	}
	if ctx.Parent().Parent().Parent() != nil {
		t.Fatal("root should have nil parent")
	}
}

func TestCommandContext_Cmd(t *testing.T) {
	sub := &Command{Name: "sub", Run: func(ctx *CommandContext) error { return nil }}
	root := &Command{Name: "root"}
	root.Sub(sub)

	ctx, err := root.Parse(&CommandContext{
		name:      "root",
		InputArgs: []string{"sub"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if ctx.Cmd() != sub {
		t.Fatal("expected sub command")
	}
}

func TestCommandContext_WithValueAndValue(t *testing.T) {
	root := &Command{Name: "root", Run: func(ctx *CommandContext) error { return nil }}

	ctx, err := root.Parse(&CommandContext{
		name:      "root",
		InputArgs: nil,
	})
	if err != nil {
		t.Fatal(err)
	}

	const key = "mykey"
	ctx.WithValue(key, "myvalue")
	v := ctx.Value(key)
	if v != "myvalue" {
		t.Fatalf("Value = %v, want myvalue", v)
	}
}

func TestCommandContext_Fork(t *testing.T) {
	parent := &CommandContext{
		name:      "parent",
		InputArgs: []string{"a", "b"},
	}

	child := parent.Fork()
	if child.Parent() != parent {
		t.Fatal("fork child parent mismatch")
	}
	if child.Name() != "parent" {
		t.Fatal("fork child should inherit name")
	}

	child.name = "child"
	if parent.Name() != "parent" {
		t.Fatal("fork should copy, not alias")
	}
}

func TestCommandContext_NilRun_ReturnsNil(t *testing.T) {
	root := &Command{Name: "root"}
	ctx := &CommandContext{cmd: root}

	if err := ctx.Run(); err != nil {
		t.Fatalf("expected nil error for nil Run, got %v", err)
	}
}


