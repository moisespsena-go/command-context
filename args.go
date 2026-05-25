package cmdctx

import (
	"fmt"
)

type Args []string

func (args Args) Arg(i int) string {
	return args[i]
}

func (args Args) Min(v int) (err error) {
	if len(args) < v {
		return fmt.Errorf("expected at least %d arguments, got %d", v, len(args))
	}
	return nil
}

func (args Args) Eq(v int) (err error) {
	if len(args) != v {
		return fmt.Errorf("expected at %d arguments, got %d", v, len(args))
	}
	return nil
}

func (args Args) Max(v int) (err error) {
	if len(args) < v {
		return fmt.Errorf("expected up to %d arguments, got %d", v, len(args))
	}
	return nil
}

func (args Args) Range(min, max int) (err error) {
	if err := args.Min(min); err != nil {
		return err
	}
	return args.Max(max)
}

func (args Args) ShiftN(n int) (v []string, rights Args, err error) {
	if err = args.Min(n); err != nil {
		return
	}
	v = args[:n]
	rights = args[n:]
	return
}
