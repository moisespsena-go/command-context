package cmdctx

import "fmt"

type ContextParserError struct {
	ctx   *CommandContext
	event string
	err   error
}

func NewContextParserError(ctx *CommandContext, event string, err error) *ContextParserError {
	return &ContextParserError{ctx: ctx, event: event, err: err}
}

func (c *ContextParserError) Ctx() *CommandContext {
	return c.ctx
}

func (c *ContextParserError) Unwrap() error {
	return c.err
}

func (c *ContextParserError) Error() string {
	var event string
	if c.event != "" {
		event = "(" + c.event + " event) "
	}
	return fmt.Sprintf("parse command %v failed: %v%v", c.ctx.FullName(), event, c.err.Error())
}

func ToContextParserError(ctx *CommandContext, event string, err error) (e *ContextParserError) {
	if e, _ = err.(*ContextParserError); e == nil {
		e = NewContextParserError(ctx, event, err)
	}
	return
}

type ContextExecuteError struct {
	ctx *CommandContext
	err error
}

func NewContextExecuteError(ctx *CommandContext, err error) *ContextExecuteError {
	return &ContextExecuteError{ctx: ctx, err: err}
}

func (c *ContextExecuteError) Ctx() *CommandContext {
	return c.ctx
}

func (c *ContextExecuteError) Unwrap() error {
	return c.err
}

func (c *ContextExecuteError) Error() string {
	return fmt.Sprintf("execute command %v failed: %v", c.ctx.FullName(), c.err.Error())
}

func ToContextExecuteError(ctx *CommandContext, err error) (e *ContextExecuteError) {
	if e, _ = err.(*ContextExecuteError); e == nil {
		e = NewContextExecuteError(ctx, err)
	}
	return
}
