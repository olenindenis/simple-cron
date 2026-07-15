package runner

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"time"
)

type Fork struct {
	logger *slog.Logger
}

func NewFork() *Fork {
	return &Fork{logger: slog.Default().With("component", "fork_runner")}
}

func (fork *Fork) Exec(ctx context.Context, command string) error {
	ex, err := os.Executable()
	if err != nil {
		fork.logger.Error("failed to resolve own executable path", "error", err)

		return fmt.Errorf("executing fork: %w", err)
	}

	discard := false
	if strings.Contains(command, ">>") {
		command = strings.TrimRight(strings.Split(command, ">>")[0], " ")
		discard = true
	}

	args := strings.Split(command, " ")
	cmd := exec.CommandContext(ctx, args[0], args[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if discard {
		cmd.Stdout = io.Discard
		cmd.Stderr = io.Discard
	}

	cmd.Dir = filepath.Dir(ex)
	cmd.Env = os.Environ()
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}
	cmd.Cancel = func() error {
		fork.logger.Warn("job timed out, sending SIGTERM to process group", "command", command, "pid", cmd.Process.Pid)

		return syscall.Kill(-cmd.Process.Pid, syscall.SIGTERM)
	}

	logger := fork.logger.With("command", command, "dir", cmd.Dir, "discard_output", discard)
	logger.Info("starting job process")

	start := time.Now()

	if err = cmd.Start(); err != nil {
		logger.Error("failed to start job process", "error", err)

		return fmt.Errorf("start error: %w", err)
	}

	logger.Info("job process started", "pid", cmd.Process.Pid)

	if err = cmd.Wait(); err != nil {
		logger.Error("job process exited with error", "pid", cmd.Process.Pid, "duration", time.Since(start), "error", err)

		return fmt.Errorf("wait error: %w", err)
	}

	logger.Info("job process finished", "pid", cmd.Process.Pid, "duration", time.Since(start))

	return nil
}
