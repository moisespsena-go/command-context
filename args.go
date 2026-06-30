package cmdctx

import (
	"fmt"
)

// Args is a slice of positional command-line arguments remaining after flag
// parsing. It provides validation helpers that return descriptive errors when
// counts are not satisfied.
type Args []string

// Arg returns the argument at index i. Panics if i is out of range.
func (args Args) Arg(i int) string {
	return args[i]
}

// Min returns an error if the number of arguments is less than v.
func (args Args) Min(v int) (err error) {
	if len(args) < v {
		return fmt.Errorf("expected at least %d arguments, got %d", v, len(args))
	}
	return nil
}

// Eq returns an error if the number of arguments is not exactly v.
func (args Args) Eq(v int) (err error) {
	if len(args) != v {
		return fmt.Errorf("expected at %d arguments, got %d", v, len(args))
	}
	return nil
}

// Max returns an error if the number of arguments exceeds v.
func (args Args) Max(v int) (err error) {
	if len(args) > v {
		return fmt.Errorf("expected up to %d arguments, got %d", v, len(args))
	}
	return nil
}

// Range returns an error if the number of arguments is outside [min, max].
func (args Args) Range(min, max int) (err error) {
	if err := args.Min(min); err != nil {
		return err
	}
	return args.Max(max)
}

// ShiftN splits args into the first n elements and the remaining elements.
// Returns an error if fewer than n arguments are available.
func (args Args) ShiftN(n int) (v []string, rights Args, err error) {
	if err = args.Min(n); err != nil {
		return
	}
	v = args[:n]
	rights = args[n:]
	return
}
