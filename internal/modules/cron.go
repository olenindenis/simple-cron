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

// jobTimeout bounds how long a single job invocation may run before it is
// killed (SIGTERM to the process group, see pkg/runner/fork.go). Without it,
// a hanging job command blocks its goroutine and OS thread forever, and
// combined with SkipIfStillRunning below, ticks would otherwise never
// recover from a stuck job.
type jobTimeout time.Duration

func newScheduler() *cron.Cron {
	return cron.New(cron.WithChain(
		cron.Recover(cron.DefaultLogger),
		cron.SkipIfStillRunning(cron.DefaultLogger),
	))
}

func Module(moduleName, crontabName, forkType string, timeout time.Duration) fx.Option {
	return fx.Module(moduleName,
		fx.Provide(
			func() services.CrontabFileName {
				return services.CrontabFileName(crontabName)
			},
			func() runner.ForkType {
				return runner.ForkType(forkType)
			},
			func() jobTimeout {
				return jobTimeout(timeout)
			},

			fx.Annotate(context.Background, fx.As(new(context.Context))),
			services.NewCrontabService,
			newScheduler,
		),
		fx.Invoke(func(
			ctx context.Context,
			lifecycle fx.Lifecycle,
			service services.CrontabService,
			scheduler *cron.Cron,
			forkType runner.ForkType,
			timeout jobTimeout,
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
						jobCtx, cancel := context.WithTimeout(ctx, time.Duration(timeout))
						defer cancel()

						if err = cmd.Exec(jobCtx, job.Command); err != nil {
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

					stopCtx := scheduler.Stop()
					select {
					case <-stopCtx.Done():
					case <-ctx.Done():
						log.Println("Shutdown timeout exceeded, forcing stop")
					}

					return nil
				},
			})

			return nil
		}),
	)
}
