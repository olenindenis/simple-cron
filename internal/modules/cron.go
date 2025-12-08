package modules

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/robfig/cron/v3"
	"go.uber.org/fx"

	"cron/internal/services"
	"cron/pkg/runner"
)

func Module(moduleName, crontabName, forkType string) fx.Option {
	return fx.Module(moduleName,
		fx.Provide(
			func() services.CrontabFileName {
				return services.CrontabFileName(crontabName)
			},
			func() runner.ForkType {
				return runner.ForkType(forkType)
			},

			fx.Annotate(context.Background, fx.As(new(context.Context))),
			services.NewCrontabService,
			cron.New,
		),
		fx.Invoke(func(
			ctx context.Context,
			lifecycle fx.Lifecycle,
			service services.CrontabService,
			scheduler *cron.Cron,
			forkType runner.ForkType,
		) error {
			cmd := runner.NewFactory(forkType).MustMake()

			lifecycle.Append(fx.Hook{
				OnStart: func(_ context.Context) error {
					log.Println("Start jobs")

					job, err := service.Parse()
					if err != nil {
						return fmt.Errorf("parse crontab: %w", err)
					}

					if entryID, err := scheduler.AddFunc(job.Spec, func() {
						if err = cmd.Exec(ctx, job.Command); err != nil {
							log.Println(err)
						}
					}); err != nil {
						scheduler.Remove(entryID)

						return fmt.Errorf("cron AddJob %w", err)
					}

					scheduler.Start()

					return nil
				},
				OnStop: func(ctx context.Context) error {
					log.Println("Stop all jobs")

					_, shutdownRelease := context.WithTimeout(ctx, 30*time.Second)
					defer shutdownRelease()

					scheduler.Stop()

					return nil
				},
			})

			return nil
		}),
	)
}
