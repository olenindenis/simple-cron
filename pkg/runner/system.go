package runner

import (
	"context"
	"fmt"
	"log"
	"os"
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

	args := strings.Split(command, " ")

	env := os.Environ()
	execErr := syscall.Exec(args[0], []string{strings.Join(args[1:], " ")}, env)
	if execErr != nil {
		return fmt.Errorf("system exec error: %s", execErr)
	}
	return nil
}
