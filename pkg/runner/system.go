package runner

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"strings"
	"syscall"
)

type System struct {
	logger *slog.Logger
}

func NewSystem() *System {
	return &System{logger: slog.Default().With("component", "system_runner")}
}

func (r *System) Exec(_ context.Context, command string) error {
	logger := r.logger.With("command", command)

	cmd := strings.Split(command, " ")

	binary, lookErr := exec.LookPath(cmd[0])
	if lookErr != nil {
		logger.Error("failed to resolve binary path", "error", lookErr)

		return fmt.Errorf("exec.LookPath: %v", lookErr)
	}

	logger.Info("replacing process image", "binary", binary)

	env := os.Environ()
	execErr := syscall.Exec(binary, cmd, env)
	if execErr != nil {
		logger.Error("syscall.Exec failed", "binary", binary, "error", execErr)

		return fmt.Errorf("system exec error: %s", execErr)
	}
	return nil
}
