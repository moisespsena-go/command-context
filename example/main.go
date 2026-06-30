package main

import (
	"fmt"
	"os"
	"path/filepath"
	"net/http"

	"github.com/mattn/go-tty"
	cc "github.com/moisespsena-go/command-context"
)

func main() {
	const bindKey = "bindAddr"

	// the default command.
	main := &cc.Command{
		Name:        filepath.Base(os.Args[0]),
		Description: "runs the http server",
		New: func(ctx *cc.CommandContext) (err error) {
			var bind string
			ctx.WithValue(bindKey, &bind)
			ctx.Flags().StringVar(&bind, "bind", "0.0.0.0:8000", "address to listen on")
			return nil
		},
		Run: func(ctx *cc.CommandContext) (err error) {
			bind := *ctx.Value(bindKey).(*string)
			fmt.Fprintf(ctx.Out, "listening on %s\n", bind)

			mux := http.NewServeMux()
			mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
				fmt.Fprint(w, "Welcome to the github.com/moisespsena-go/command-context example page!")
			})
			return http.ListenAndServe(bind, mux)
		},
	}

	// the password manager sub command
	passwd := &cc.Command{
		Name:        "passwd",
		Description: "changes user password",
		Usage:       "USER_NAME",
		ParseArgs: func(ctx *cc.CommandContext) (err error) {
			// specify require exact one arg (USER_NAME)
			return ctx.Args.Eq(1)
		},
		Run: func(ctx *cc.CommandContext) (err error) {
			userName := ctx.Args[0]
			fmt.Fprintf(os.Stdin, "Enter a new %q password: ", userName)

			var TTY *tty.TTY
			if TTY, err = tty.Open(); err != nil {
				return
			}

			defer TTY.Close()

			var pwd string
			if pwd, err = TTY.ReadPassword(); err != nil {
				return
			}

			if len(pwd) < 6 {
				err = fmt.Errorf("expected at least %d chars, got %d", 6, len(pwd))
				return
			}

			fmt.Fprint(os.Stdin, "Confirm password: ")

			var pwd2 string
			if pwd2, err = TTY.ReadPassword(); err != nil {
				return
			}

			if pwd2 != pwd {
				err = fmt.Errorf("Passwords not equal")
				return
			}

			fmt.Fprintf(ctx.Out, "password for user %q changed.\n", userName)
			return nil
		},
	}

	// typed struct holding parsed flags, stored in ctx and retrieved in Run
	type flags struct {
		Role       string
		Department string
	}

	// the user management sub command with its own sub-commands
	user := &cc.Command{
		Name:        "user",
		Description: "manage system users",
		Usage:       "[--role ROLE] [--dept DEPT] [add|rm]",
		New: func(ctx *cc.CommandContext) (err error) {
			f := &flags{}
			flgs := ctx.Flags()
			flgs.StringVar(&f.Role, "role", "viewer", "default role for new users")
			flgs.StringVar(&f.Department, "dept", "engineering", "department name")
			ctx.WithValue("flags", f)
			return nil
		},
		Run: func(ctx *cc.CommandContext) (err error) {
			f := ctx.Value("flags").(*flags)
			fmt.Fprintf(ctx.Out, "manage users (role=%s, dept=%s): args=%v\n", f.Role, f.Department, ctx.Args)
			return nil
		},
	}

	userAdd := &cc.Command{
		Name:        "add",
		Description: "add a new user",
		Usage:       "USER_NAME",
		ParseArgs: func(ctx *cc.CommandContext) (err error) {
			return ctx.Args.Eq(1)
		},
		Run: func(ctx *cc.CommandContext) (err error) {
			fmt.Fprintf(ctx.Out, "adding user %q\n", ctx.Args[0])
			return nil
		},
	}

	userRm := &cc.Command{
		Name:        "rm",
		Description: "remove an existing user",
		Usage:       "USER_NAME",
		ParseArgs: func(ctx *cc.CommandContext) (err error) {
			return ctx.Args.Eq(1)
		},
		Run: func(ctx *cc.CommandContext) (err error) {
			fmt.Fprintf(ctx.Out, "removing user %q\n", ctx.Args[0])
			return nil
		},
	}

	user.Sub(userAdd).Sub(userRm)

	// take passwd and user as subcommands of main
	main.Sub(passwd)
	main.Sub(user)

	ctx, err := main.Parse(nil)

	if err != nil {
		fmt.Fprintln(os.Stderr, "ERROR:", err)
		os.Exit(1)
		return
	}

	if err = ctx.Run(); err != nil {
		fmt.Fprintln(os.Stderr, "ERROR:", cc.ToContextExecuteError(ctx, err))
		os.Exit(1)
	}
}
