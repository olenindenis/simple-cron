package services

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strings"
)

type Job struct {
	Spec    string
	Command string
}

type CrontabService struct {
	path   string
	logger *slog.Logger
}

func NewCrontabService(path string) CrontabService {
	return CrontabService{
		path:   path,
		logger: slog.Default().With("component", "crontab_service"),
	}
}

// Parse reads a single crontab line: five schedule fields followed by the command.
func (c CrontabService) Parse() (Job, error) {
	c.logger.Info("parsing crontab file", "file", c.path)

	data, err := os.ReadFile(c.path)
	if err != nil {
		return Job{}, fmt.Errorf("read crontab file: %w", err)
	}

	res := strings.Split(strings.TrimSpace(string(data)), " ")
	if len(res) < 6 {
		return Job{}, errors.New("invalid crontab line: expected 5 schedule fields and a command")
	}

	return Job{
		Spec:    strings.Join(res[:5], " "),
		Command: strings.Join(res[5:], " "),
	}, nil
}
