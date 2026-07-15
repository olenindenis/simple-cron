package services

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"

	"cron/internal/domain"
)

type CrontabFileName string

type CrontabService struct {
	crontabFile CrontabFileName
	logger      *slog.Logger
}

func NewCrontabService(crontabFile CrontabFileName) CrontabService {
	return CrontabService{
		crontabFile: crontabFile,
		logger:      slog.Default().With("component", "crontab_service"),
	}
}

func (c CrontabService) Parse() (domain.Job, error) {
	c.logger.Info("parsing crontab file", "file", string(c.crontabFile))

	f, err := os.Open(string(c.crontabFile))
	if err != nil {
		c.logger.Error("failed to open crontab file", "file", string(c.crontabFile), "error", err)

		return domain.Job{}, fmt.Errorf("os.Open %w", err)
	}
	defer func(f *os.File) {
		if err = f.Close(); err != nil {
			panic(fmt.Errorf("f.Close %w", err))
		}
	}(f)

	data, err := io.ReadAll(f)
	if err != nil {
		c.logger.Error("failed to read crontab file", "file", string(c.crontabFile), "error", err)

		return domain.Job{}, fmt.Errorf("io.ReadAll %w", err)
	}

	res := strings.Split(string(data), " ")

	job := domain.Job{
		Spec:    strings.Join(res[:5], " "),
		Command: strings.Join(res[5:], " "),
	}

	c.logger.Info("crontab file parsed", "spec", job.Spec, "command", job.Command)

	return job, nil
}
