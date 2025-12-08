package main

import (
	"context"
	"errors"
	"log"
	"os"

	"github.com/urfave/cli/v3"
	"go.uber.org/fx"

	"cron/internal/modules"
)

const moduleName = "cron"

func main() {
	ctx := context.Background()

	app := &cli.Command{
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "crontab",
				Aliases: []string{"c"},
				Usage:   "crontab config file",
			},
			&cli.StringFlag{
				Name:        "fork",
				Aliases:     []string{"f"},
				Usage:       "process fork type (one of [\"system\", \"own\")",
				Value:       "system",
				DefaultText: "random",
			},
		},
		Action: func(cCtx context.Context, cmd *cli.Command) error {
			if cmd.String("crontab") == "" {
				return errors.New("file is required")
			}

			log.Printf("Run with crontab name: %q\n", cmd.String("crontab"))

			fx.New(
				modules.Module(
					moduleName,
					cmd.String("crontab"),
					cmd.String("fork"),
				),
			).Run()

			return nil
		},
	}

	if err := app.Run(ctx, os.Args); err != nil {
		log.Fatal(err)
	}
}
