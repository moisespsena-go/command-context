package cmdctx

import "fmt"

// ContextParserError is returned when a command fails during the parse phase.
// It records which command was being parsed and which lifecycle event triggered
// the error, enabling callers to distinguish parse failures from execution failures.
type ContextParserError struct {
	ctx   *CommandContext
	event string
	err   error
}

// NewContextParserError creates a ContextParserError for the given context,
// lifecycle event name, and underlying error.
func NewContextParserError(ctx *CommandContext, event string, err error) *ContextParserError {
	return &ContextParserError{ctx: ctx, event: event, err: err}
}

// Ctx returns the CommandContext in which the parse error occurred.
func (c *ContextParserError) Ctx() *CommandContext {
	return c.ctx
}

// Unwrap returns the underlying error, enabling [errors.Is] and [errors.As] unwrapping.
func (c *ContextParserError) Unwrap() error {
	return c.err
}

// Error returns a formatted message that includes the full command path, the
// lifecycle event name, and the underlying error.
func (c *ContextParserError) Error() string {
	var event string
	if c.event != "" {
		event = "(" + c.event + " event) "
	}
	return fmt.Sprintf("parse command %v failed: %v%v", c.ctx.FullName(), event, c.err.Error())
}

// ToContextParserError wraps err in a ContextParserError unless err already is one,
// in which case it is returned unchanged.
func ToContextParserError(ctx *CommandContext, event string, err error) (e *ContextParserError) {
	if e, _ = err.(*ContextParserError); e == nil {
		e = NewContextParserError(ctx, event, err)
	}
	return
}

// ContextExecuteError is returned when a command fails during execution (the Run phase).
// Use [ToContextExecuteError] to wrap errors from [CommandContext.Run].
type ContextExecuteError struct {
	ctx *CommandContext
	err error
}

// NewContextExecuteError creates a ContextExecuteError for the given context and underlying error.
func NewContextExecuteError(ctx *CommandContext, err error) *ContextExecuteError {
	return &ContextExecuteError{ctx: ctx, err: err}
}

// Ctx returns the CommandContext in which execution failed.
func (c *ContextExecuteError) Ctx() *CommandContext {
	return c.ctx
}

// Unwrap returns the underlying error, enabling [errors.Is] and [errors.As] unwrapping.
func (c *ContextExecuteError) Unwrap() error {
	return c.err
}

// Error returns a formatted message that includes the full command path and
// the underlying error.
func (c *ContextExecuteError) Error() string {
	return fmt.Sprintf("execute command %v failed: %v", c.ctx.FullName(), c.err.Error())
}

// ToContextExecuteError wraps err in a ContextExecuteError unless err already is one,
// in which case it is returned unchanged.
func ToContextExecuteError(ctx *CommandContext, err error) (e *ContextExecuteError) {
	if e, _ = err.(*ContextExecuteError); e == nil {
		e = NewContextExecuteError(ctx, err)
	}
	return
}
