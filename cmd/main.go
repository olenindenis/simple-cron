package main

import (
	"errors"
	"log"
	"os"

	"github.com/urfave/cli/v2"
	"go.uber.org/fx"

	"cron/internal/modules"
)

const moduleName = "cron"

func main() {
	app := &cli.App{
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "file",
				Usage: "crontab config file",
			},
		},
		Action: func(cCtx *cli.Context) error {
			if cCtx.String("file") == "" {
				return errors.New("file is required")
			}

			log.Printf("Run with crontab name: %q\n", cCtx.String("file"))

			fx.New(
				modules.Module(
					moduleName,
					cCtx.String("file"),
				),
			).Run()

			return nil
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
