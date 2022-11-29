package starter

import (
	"github.com/urfave/cli/v2"
)

type Component interface {
	Name() string
	Init(ctx *cli.Context) error
}

type component struct {
	name string
	f    func(ctx *cli.Context) error
}

func NewComponent(name string, f func(ctx *cli.Context) error) Component {
	return &component{
		name: name,
		f:    f,
	}
}

func (c *component) Name() string {
	return c.name
}

func (c *component) Init(ctx *cli.Context) error {
	return c.f(ctx)
}
