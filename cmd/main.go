package main

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"time"

	"github.com/urfave/cli/v3"
	"go.uber.org/fx"

	"cron/internal/modules"
)

const moduleName = "cron"

func main() {
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})))

	ctx := context.Background()

	app := &cli.Command{
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "crontab",
				Aliases: []string{"c"},
				Usage:   "crontab config file",
			},
			&cli.DurationFlag{
				Name:    "timeout",
				Aliases: []string{"t"},
				Usage:   "max duration a single job run may take before it is killed",
				Value:   5 * time.Minute,
			},
		},
		Action: func(cCtx context.Context, cmd *cli.Command) error {
			if cmd.String("crontab") == "" {
				return errors.New("file is required")
			}

			slog.Info("starting cron daemon",
				"crontab", cmd.String("crontab"),
				"timeout", cmd.Duration("timeout"),
			)

			fx.New(
				modules.Module(moduleName, cmd.String("crontab"), cmd.Duration("timeout")),
			).Run()

			return nil
		},
	}

	if err := app.Run(ctx, os.Args); err != nil {
		slog.Error("cron daemon exited with error", "error", err)
		os.Exit(1)
	}
}
