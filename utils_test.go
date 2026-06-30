package cmdctx

import (
	"slices"
	"testing"
)

func TestSplitString_Empty(t *testing.T) {
	res := splitString("", 10)
	if res == nil || len(res) != 1 || res[0] != "" {
		t.Fatalf("got %v, want [\"\"]", res)
	}
}

func TestSplitString_ShorterThanMax(t *testing.T) {
	res := splitString("hello world", 80)
	if len(res) != 1 || res[0] != "hello world" {
		t.Fatalf("got %v, want [hello world]", res)
	}
}

func TestSplitString_WrapAtBoundary(t *testing.T) {
	res := splitString("hello world foo bar", 11)
	want := []string{"hello world", "foo bar"}
	if !slices.Equal(res, want) {
		t.Fatalf("got %v, want %v", res, want)
	}
}

func TestSplitString_ExactFit(t *testing.T) {
	res := splitString("hello world", 11)
	want := []string{"hello world"}
	if !slices.Equal(res, want) {
		t.Fatalf("got %v, want %v", res, want)
	}
}

func TestSplitString_SingleWordLongerThanMax(t *testing.T) {
	res := splitString("hello", 3)
	// the current word "hello" exceeds maxLen and is placed on its own line;
	// the empty previous line produces a leading empty string
	if len(res) != 2 || res[0] != "" || res[1] != "hello" {
		t.Fatalf("got %v, want [\"\" \"hello\"]", res)
	}
}

func TestSplitString_MultipleWraps(t *testing.T) {
	res := splitString("a bb ccc dd eeeee fff", 5)
	// "ccc dd" fits because 3+2=5 which does NOT exceed the maxLen of 5
	want := []string{"a bb", "ccc dd", "eeeee", "fff"}
	if !slices.Equal(res, want) {
		t.Fatalf("got %v, want %v", res, want)
	}
}

func TestSplitString_MultipleSpaces(t *testing.T) {
	res := splitString("hello    world", 80)
	want := []string{"hello world"}
	if !slices.Equal(res, want) {
		t.Fatalf("got %v, want %v", res, want)
	}
}
