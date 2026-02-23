package runner

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"syscall"
)

type System struct {
}

func NewSystem() *System {
	return &System{}
}

func (r *System) Exec(_ context.Context, command string) error {
	log.Println(command)

	cmd := strings.Split(command, " ")

	binary, lookErr := exec.LookPath(cmd[0])
	if lookErr != nil {
		return fmt.Errorf("exec.LookPath: %v", lookErr)
	}

	env := os.Environ()
	execErr := syscall.Exec(binary, cmd, env)
	if execErr != nil {
		return fmt.Errorf("system exec error: %s", execErr)
	}
	return nil
}
