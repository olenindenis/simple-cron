package runner

import (
	"context"
	"fmt"
	"log"
	"os"
	"syscall"
)

type System struct {
}

func NewSystem() *System {
	return &System{}
}

func (r *System) Exec(_ context.Context, command string) error {
	log.Println(command)

	env := os.Environ()
	execErr := syscall.Exec("sh", []string{"-c", command}, env)
	if execErr != nil {
		return fmt.Errorf("system exec error: %s", execErr)
	}
	return nil
}
