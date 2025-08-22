package modules

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"time"

	"cron/internal/services"
	"github.com/robfig/cron/v3"
	"go.uber.org/fx"
)

func Module(moduleName, crontabName string) fx.Option {
	return fx.Module(moduleName,
		fx.Supply(crontabName),
		fx.Provide(
			fx.Annotate(context.Background, fx.As(new(context.Context))),
			services.NewCrontabService,
			cron.New,
		),
		fx.Invoke(func(
			ctx context.Context,
			lifecycle fx.Lifecycle,
			service services.CrontabService,
			scheduler *cron.Cron,
		) error {
			lifecycle.Append(fx.Hook{
				OnStart: func(_ context.Context) error {
					log.Println("Start jobs")

					job, err := service.Parse()
					if err != nil {
						return err
					}
					if _, err = scheduler.AddFunc(job.Spec, func() {
						log.Println(job.Command)

						cmd := exec.Command("sh", "-c", fmt.Sprintf("\"%s\"", job.Command))
						cmd.Stdout = os.Stdout
						cmd.Stderr = os.Stderr
						if err = cmd.Run(); err != nil {
							log.Println(fmt.Errorf("run: %w", err))
						}
					}); err != nil {
						return fmt.Errorf("cron AddFunc %w", err)
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
