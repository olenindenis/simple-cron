package runner

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
)

type Fork struct{}

func NewFork() *Fork {
	return &Fork{}
}

func (fork *Fork) Exec(ctx context.Context, command string) error {
	ex, err := os.Executable()
	if err != nil {
		return fmt.Errorf("executing fork: %w", err)
	}

	cmd := exec.CommandContext(ctx, "sh", "-c", command)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Dir = filepath.Dir(ex)
	cmd.Env = append(os.Environ())
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}

	log.Printf("Exec command: %s in path: %s\n", command, cmd.Dir)

	if err = cmd.Start(); err != nil {
		return fmt.Errorf("start error: %w", err)
	}

	if err = cmd.Wait(); err != nil {
		return fmt.Errorf("wait error: %w", err)
	}
	return nil
}
