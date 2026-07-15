package runner

import "log/slog"

type Factory struct {
	forkType ForkType
}

func NewFactory(forkType ForkType) *Factory {
	return &Factory{
		forkType: forkType,
	}
}

func (f *Factory) MustMake() Runner {
	logger := slog.Default().With("component", "runner_factory")

	switch f.forkType {
	case SystemFork:
		logger.Info("creating runner", "fork_type", SystemFork)
		return NewSystem()
	case OwnFork:
		logger.Info("creating runner", "fork_type", OwnFork)
		return NewFork()
	default:
		logger.Error("unknown fork type requested", "fork_type", f.forkType)
		panic("unknown fork type")
	}
}
