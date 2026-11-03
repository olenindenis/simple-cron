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

// Module wires the crontab parser, the scheduler and the job runner into the
// fx lifecycle. timeout bounds how long a single job invocation may run before
// it is killed (SIGTERM to the process group, see pkg/runner/fork.go); combined
// with SkipIfStillRunning, a hanging job would otherwise block all later ticks.
func Module(moduleName, crontabName string, timeout time.Duration) fx.Option {
	return fx.Module(moduleName,
		fx.Invoke(func(lifecycle fx.Lifecycle) {
			logger := slog.Default().With("application", moduleName)
			service := services.NewCrontabService(crontabName)
			cronLogger := logging.NewCronLogger(slog.Default().With("component", "robfig-cron"))
			scheduler := cron.New(cron.WithChain(
				cron.Recover(cronLogger),
				cron.SkipIfStillRunning(cronLogger),
			))
			fork := runner.NewFork()

			lifecycle.Append(fx.Hook{
				OnStart: func(_ context.Context) error {
					logger.Info("starting scheduler")

					job, err := service.Parse()
					if err != nil {
						return fmt.Errorf("parse crontab: %w", err)
					}

					logger.Info("crontab job parsed", "spec", job.Spec, "command", job.Command)

					entryID, err := scheduler.AddFunc(job.Spec, func() {
						jobLogger := logger.With("spec", job.Spec, "command", job.Command)
						jobCtx, cancel := context.WithTimeout(context.Background(), timeout)
						defer cancel()

						jobLogger.Info("job execution triggered")
						start := time.Now()

						if err := fork.Exec(jobCtx, job.Command); err != nil {
							jobLogger.Error("job execution failed", "duration", time.Since(start), "error", err)

							return
						}

						jobLogger.Info("job execution completed", "duration", time.Since(start))
					})
					if err != nil {
						return fmt.Errorf("add cron job %q: %w", job.Spec, err)
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
		}),
	)
}
