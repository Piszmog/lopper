package main

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/urfave/cli/v2"
	"lopper/ui"
	"os"
)

func main() {
	app := &cli.App{
		Name:  "lopper",
		Usage: "removes dead local Git branches",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "Path",
				Aliases:  []string{"p"},
				Usage:    "Path to the repository or root directory containing Git repositories",
				Required: true,
			},
			&cli.StringSliceFlag{
				Name:    "protected-branch",
				Aliases: []string{"b"},
				Usage:   "branches that are protected from deletion (e.g. -b foo -b bar -b baz)",
			},
			&cli.IntFlag{
				Name:    "concurrency",
				Aliases: []string{"c"},
				Usage:   "determines how many repositories to process concurrently",
				Value:   1,
			},
			&cli.BoolFlag{
				Name:  "dry-run",
				Usage: "runs thru the process without actually removing branches",
			},
		},
		Action: func(ctx *cli.Context) error {
			m := ui.NewModel(
				ui.Path(ctx.String("Path")),
				ui.ProtectedBranches(ctx.StringSlice("protected-branch")),
				ui.Concurrency(ctx.Int("concurrency")),
				ui.DryRun(ctx.Bool("dry-run")),
			)
			if err := tea.NewProgram(m, tea.WithAltScreen()).Start(); err != nil {
				return err
			}
			if m.Error() != nil {
				return m.Error()
			}
			return nil
		},
	}
	// start lopping some branches
	if err := app.Run(os.Args); err != nil {
		fmt.Println(err)
	}
}
