package cli

import (
	"flag"
	"io"
	"os"
)

func NewApp(out, errOut io.Writer) *App {
	if out == nil {
		out = os.Stdout
	}
	if errOut == nil {
		errOut = os.Stderr
	}
	return &App{out: out, errOut: errOut}
}

type App struct {
	out    io.Writer
	errOut io.Writer
}

func (c *App) Run(argv []string) error {
	fs := flag.NewFlagSet(argv[0], flag.ContinueOnError)
	fs.SetOutput(c.errOut)
	err := fs.Parse(argv[1:])
	if err == flag.ErrHelp {
		return nil
	}
	if err != nil {
		return err
	}
	return nil
}
