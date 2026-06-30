package cmdctx

import (
	"testing"
)

func TestArgs_Arg(t *testing.T) {
	a := Args{"foo", "bar"}
	if a.Arg(0) != "foo" {
		t.Fatalf("Arg(0) = %q, want foo", a.Arg(0))
	}
	if a.Arg(1) != "bar" {
		t.Fatalf("Arg(1) = %q, want bar", a.Arg(1))
	}
}

func TestArgs_Arg_PanicsOutOfRange(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for out-of-range Arg")
		}
	}()
	a := Args{"only"}
	_ = a.Arg(1)
}

func TestArgs_Min(t *testing.T) {
	tests := []struct {
		args Args
		v    int
		want error
	}{
		{Args{"a", "b"}, 2, nil},
		{Args{"a", "b"}, 1, nil},
		{Args{"a", "b"}, 3, errStr("expected at least 3 arguments, got 2")},
		{Args{}, 0, nil},
		{Args{}, 1, errStr("expected at least 1 arguments, got 0")},
	}
	for _, tc := range tests {
		got := tc.args.Min(tc.v)
		if !errEq(got, tc.want) {
			t.Fatalf("Min(%d) on %v: got %v, want %v", tc.v, tc.args, got, tc.want)
		}
	}
}

func TestArgs_Eq(t *testing.T) {
	tests := []struct {
		args Args
		v    int
		want error
	}{
		{Args{"a", "b"}, 2, nil},
		{Args{"a", "b"}, 1, errStr("expected at 1 arguments, got 2")},
		{Args{"a"}, 2, errStr("expected at 2 arguments, got 1")},
		{Args{}, 0, nil},
	}
	for _, tc := range tests {
		got := tc.args.Eq(tc.v)
		if !errEq(got, tc.want) {
			t.Fatalf("Eq(%d) on %v: got %v, want %v", tc.v, tc.args, got, tc.want)
		}
	}
}

func TestArgs_Max(t *testing.T) {
	tests := []struct {
		args Args
		v    int
		want error
	}{
		{Args{"a", "b", "c"}, 5, nil},
		{Args{}, 0, nil},
		{Args{"a"}, 1, nil},
		{Args{"a", "b"}, 0, errStr("expected up to 0 arguments, got 2")},
	}
	for _, tc := range tests {
		got := tc.args.Max(tc.v)
		if !errEq(got, tc.want) {
			t.Fatalf("Max(%d) on %v: got %v, want %v", tc.v, tc.args, got, tc.want)
		}
	}
}

func TestArgs_Range(t *testing.T) {
	tests := []struct {
		args    Args
		min, max int
		want    error
	}{
		{Args{"a"}, 1, 1, nil},
		{Args{"a", "b"}, 1, 2, nil},
		{Args{}, 0, 0, nil},
		{Args{"a", "b", "c"}, 1, 2, errStr("expected up to 2 arguments, got 3")},
		{Args{}, 1, 2, errStr("expected at least 1 arguments, got 0")},
	}
	for _, tc := range tests {
		got := tc.args.Range(tc.min, tc.max)
		if !errEq(got, tc.want) {
			t.Fatalf("Range(%d,%d) on %v: got %v, want %v", tc.min, tc.max, tc.args, got, tc.want)
		}
	}
}

func TestArgs_ShiftN(t *testing.T) {
	a := Args{"a", "b", "c"}
	v, rights, err := a.ShiftN(2)
	if err != nil {
		t.Fatal(err)
	}
	if len(v) != 2 || v[0] != "a" || v[1] != "b" {
		t.Fatalf("v = %v, want [a b]", v)
	}
	if len(rights) != 1 || rights[0] != "c" {
		t.Fatalf("rights = %v, want [c]", rights)
	}
}

func TestArgs_ShiftN_NotEnough(t *testing.T) {
	a := Args{"a"}
	_, _, err := a.ShiftN(2)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestArgs_ShiftN_Exact(t *testing.T) {
	a := Args{"a", "b"}
	v, rights, err := a.ShiftN(2)
	if err != nil {
		t.Fatal(err)
	}
	if len(v) != 2 {
		t.Fatalf("v = %v, want length 2", v)
	}
	if len(rights) != 0 {
		t.Fatalf("rights = %v, want empty", rights)
	}
}

func TestArgs_ShiftN_Zero(t *testing.T) {
	a := Args{"a", "b"}
	v, rights, err := a.ShiftN(0)
	if err != nil {
		t.Fatal(err)
	}
	if len(v) != 0 {
		t.Fatalf("v = %v, want empty", v)
	}
	if len(rights) != 2 {
		t.Fatalf("rights = %v, want [a b]", rights)
	}
}

func errStr(s string) error {
	return &errorString{s}
}

type errorString struct{ s string }

func (e *errorString) Error() string { return e.s }

func errEq(a, b error) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return a.Error() == b.Error()
}
