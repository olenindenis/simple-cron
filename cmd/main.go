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
				Name:  "file",
				Usage: "crontab config file",
			},
		},
		Action: func(cCtx context.Context, cmd *cli.Command) error {
			if cmd.String("file") == "" {
				return errors.New("file is required")
			}

			log.Printf("Run with crontab name: %q\n", cmd.String("file"))

			fx.New(
				modules.Module(
					moduleName,
					cmd.String("file"),
				),
			).Run()

			return nil
		},
	}

	if err := app.Run(ctx, os.Args); err != nil {
		log.Fatal(err)
	}
}
