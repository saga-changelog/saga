package main

import (
	"github.com/alecthomas/kong"
	"github.com/saga-changelog/saga/internal/cli"
)

type CLI struct {
	Dir string `help:"Path to .saga directory (bypasses git repo lookup)." type:"existingdir"`

	Init      cli.InitCmd      `cmd:"" help:"Initialize Saga in the current git repository."`
	Validate  cli.ValidateCmd  `cmd:"" help:"Validate all configuration and feat files."`
	Pending   cli.PendingCmd   `cmd:"" help:"List pending feats."`
	Chapters  cli.ChaptersCmd  `cmd:"" help:"List chapters or manage them."`
	Tell      cli.TellCmd      `cmd:"" help:"Dispatch couriers for a chapter."`
	Audiences cli.AudiencesCmd `cmd:"" help:"List configured audiences and routes."`
	Couriers  cli.CouriersCmd  `cmd:"" help:"List available courier plugins."`
	ThemeTest cli.ThemeTestCmd `cmd:"" hidden:"" help:"Display themed output for visual validation."`
}

func main() {
	var c CLI
	ctx := kong.Parse(&c,
		kong.Name("saga"),
		kong.Description("Communicate software changes to the right people through the right channels."),
		kong.UsageOnError(),
	)
	cli.SetGlobalDir(c.Dir)
	err := ctx.Run()
	ctx.FatalIfErrorf(err)
}
