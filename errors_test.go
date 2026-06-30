package cmdctx

import (
	"errors"
	"testing"
)

func TestContextParserError_New(t *testing.T) {
	ctx := &CommandContext{name: "root"}
	underlying := errors.New("something broke")
	e := NewContextParserError(ctx, "parse", underlying)

	if e.Ctx() != ctx {
		t.Fatal("Ctx() mismatch")
	}
	if !errors.Is(e, underlying) {
		t.Fatal("errors.Is should unwrap to underlying error")
	}

	got := e.Error()
	want := `parse command "root" failed: (parse event) something broke`
	if got != want {
		t.Fatalf("Error() = %q, want %q", got, want)
	}
}

func TestContextParserError_EmptyEvent(t *testing.T) {
	ctx := &CommandContext{name: "root"}
	e := NewContextParserError(ctx, "", errors.New("fail"))

	got := e.Error()
	want := `parse command "root" failed: fail`
	if got != want {
		t.Fatalf("Error() = %q, want %q", got, want)
	}
}

func TestContextParserError_ToContextParserError_Wraps(t *testing.T) {
	ctx := &CommandContext{name: "root"}
	original := errors.New("original")
	e := ToContextParserError(ctx, "test", original)

	if e == nil {
		t.Fatal("ToContextParserError returned nil")
	}
	if !errors.Is(e, original) {
		t.Fatal("should wrap original error")
	}
}

func TestContextParserError_ToContextParserError_Passthrough(t *testing.T) {
	ctx := &CommandContext{name: "root"}
	original := NewContextParserError(ctx, "first", errors.New("inner"))
	e := ToContextParserError(ctx, "second", original)

	if e != original {
		t.Fatal("ToContextParserError should return original when already a ContextParserError")
	}
}

func TestContextExecuteError_New(t *testing.T) {
	ctx := &CommandContext{name: "root"}
	underlying := errors.New("exec failed")
	e := NewContextExecuteError(ctx, underlying)

	if e.Ctx() != ctx {
		t.Fatal("Ctx() mismatch")
	}
	if !errors.Is(e, underlying) {
		t.Fatal("errors.Is should unwrap to underlying error")
	}

	got := e.Error()
	want := `execute command "root" failed: exec failed`
	if got != want {
		t.Fatalf("Error() = %q, want %q", got, want)
	}
}

func TestContextExecuteError_ToContextExecuteError_Wraps(t *testing.T) {
	ctx := &CommandContext{name: "root"}
	original := errors.New("original")
	e := ToContextExecuteError(ctx, original)

	if e == nil {
		t.Fatal("ToContextExecuteError returned nil")
	}
	if !errors.Is(e, original) {
		t.Fatal("should wrap original error")
	}
}

func TestContextExecuteError_ToContextExecuteError_Passthrough(t *testing.T) {
	ctx := &CommandContext{name: "root"}
	original := NewContextExecuteError(ctx, errors.New("inner"))
	e := ToContextExecuteError(ctx, original)

	if e != original {
		t.Fatal("ToContextExecuteError should return original when already a ContextExecuteError")
	}
}

func TestContextParserError_FullNameWithPath(t *testing.T) {
	leaf := &Command{Name: "leaf"}
	mid := &Command{Name: "mid"}
	mid.Sub(leaf)
	root := &Command{Name: "root"}
	root.Sub(mid)

	leafCtx, _ := root.Parse(&CommandContext{
		name:      "root",
		InputArgs: []string{"mid", "leaf"},
	})

	e := NewContextParserError(leafCtx, "test", errors.New("err"))
	got := e.Error()
	want := `parse command "root" 🠆 "mid" 🠆 "leaf" failed: (test event) err`
	if got != want {
		t.Fatalf("Error() for nested command = %q, want %q", got, want)
	}
}

func TestContextExecuteError_FullNameWithPath(t *testing.T) {
	leaf := &Command{Name: "leaf", Run: func(ctx *CommandContext) error { return nil }}
	mid := &Command{Name: "mid"}
	mid.Sub(leaf)
	root := &Command{Name: "root"}
	root.Sub(mid)

	leafCtx, _ := root.Parse(&CommandContext{
		name:      "root",
		InputArgs: []string{"mid", "leaf"},
	})

	e := NewContextExecuteError(leafCtx, errors.New("err"))
	got := e.Error()
	want := `execute command "root" 🠆 "mid" 🠆 "leaf" failed: err`
	if got != want {
		t.Fatalf("Error() for nested command = %q, want %q", got, want)
	}
}
