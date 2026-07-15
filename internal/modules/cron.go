package modules

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/robfig/cron/v3"
	"go.uber.org/fx"

	"cron/internal/services"
	"cron/pkg/logging"
	"cron/pkg/runner"
)

// jobTimeout bounds how long a single job invocation may run before it is
// killed (SIGTERM to the process group, see pkg/runner/fork.go). Without it,
// a hanging job command blocks its goroutine and OS thread forever, and
// combined with SkipIfStillRunning below, ticks would otherwise never
// recover from a stuck job.
type jobTimeout time.Duration

func Module(moduleName, crontabName, forkType string, timeout time.Duration) fx.Option {
	return fx.Module(moduleName,
		fx.Provide(
			func() *slog.Logger {
				return slog.Default().With("application", moduleName)
			},
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
			func() *cron.Cron {
				cronLogger := logging.NewCronLogger(slog.Default().With("component", "robfig-cron"))

				return cron.New(cron.WithChain(
					cron.Recover(cronLogger),
					cron.SkipIfStillRunning(cronLogger),
				))
			},
		),
		fx.Invoke(func(
			ctx context.Context,
			lifecycle fx.Lifecycle,
			service services.CrontabService,
			scheduler *cron.Cron,
			forkType runner.ForkType,
			timeout jobTimeout,
			logger *slog.Logger,
		) error {
			cmd := runner.NewFactory(forkType).MustMake()

			lifecycle.Append(fx.Hook{
				OnStart: func(_ context.Context) error {
					logger.Info("starting scheduler")

					job, err := service.Parse()
					if err != nil {
						logger.Error("failed to parse crontab", "error", err)

						return fmt.Errorf("parse crontab: %w", err)
					}

					logger.Info("crontab job parsed", "spec", job.Spec, "command", job.Command)

					entryID, err := scheduler.AddFunc(job.Spec, func() {
						jobLogger := logger.With("spec", job.Spec, "command", job.Command)
						jobCtx, cancel := context.WithTimeout(ctx, time.Duration(timeout))
						defer cancel()

						jobLogger.Info("job execution triggered")
						start := time.Now()

						if err := cmd.Exec(jobCtx, job.Command); err != nil {
							jobLogger.Error("job execution failed", "duration", time.Since(start), "error", err)

							return
						}

						jobLogger.Info("job execution completed", "duration", time.Since(start))
					})
					if err != nil {
						scheduler.Remove(entryID)
						logger.Error("failed to register cron job", "spec", job.Spec, "error", err)

						return fmt.Errorf("cron AddJob %w", err)
					}

					scheduler.Start()
					logger.Info("scheduler started", "entry_id", entryID)

					return nil
				},
				OnStop: func(ctx context.Context) error {
					logger.Info("stopping scheduler")

					stopCtx := scheduler.Stop()
					select {
					case <-stopCtx.Done():
						logger.Info("scheduler stopped")
					case <-ctx.Done():
						logger.Warn("shutdown timeout exceeded, forcing stop")
					}

					return nil
				},
			})

			return nil
		}),
	)
}
