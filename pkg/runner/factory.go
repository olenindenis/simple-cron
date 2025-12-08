package runner

import "log"

type Factory struct {
	forkType ForkType
}

func NewFactory(forkType ForkType) *Factory {
	return &Factory{
		forkType: forkType,
	}
}

func (f *Factory) MustMake() Runner {
	switch f.forkType {
	case SystemFork:
		log.Println("make runner with type", SystemFork)
		return NewSystem()
	case OwnFork:
		log.Println("make runner with type", OwnFork)
		return NewFork()
	default:
		panic("unknown fork type")
	}
}
