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
		return fmt.Errorf("resolve own executable path: %w", err)
	}

	discard := false
	if strings.Contains(command, ">>") {
		command = strings.TrimRight(strings.Split(command, ">>")[0], " ")
		discard = true
	}

	args := strings.Split(command, " ")
	cmd := exec.CommandContext(ctx, args[0], args[1:]...)
	if discard {
		cmd.Stdout = io.Discard
		cmd.Stderr = io.Discard
	} else {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}

	cmd.Dir = filepath.Dir(ex)
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}
	cmd.Cancel = func() error {
		fork.logger.Warn("job timed out, sending SIGTERM to process group", "command", command, "pid", cmd.Process.Pid)

		return syscall.Kill(-cmd.Process.Pid, syscall.SIGTERM)
	}

	if err = cmd.Start(); err != nil {
		return fmt.Errorf("start error: %w", err)
	}

	fork.logger.Info("job process started", "command", command, "dir", cmd.Dir, "discard_output", discard, "pid", cmd.Process.Pid)

	if err = cmd.Wait(); err != nil {
		return fmt.Errorf("wait error (pid %d): %w", cmd.Process.Pid, err)
	}

	return nil
}
