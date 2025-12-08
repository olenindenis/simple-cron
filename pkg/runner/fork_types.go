package runner

import "context"

const (
	SystemFork ForkType = "system"
	OwnFork    ForkType = "own"
)

type Runner interface {
	Exec(ctx context.Context, cmd string) error
}

type ForkType string
