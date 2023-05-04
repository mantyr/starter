package starter

import (
	"github.com/urfave/cli/v2"
)

type Component interface {
	Name() string
	Init(ctx *cli.Context) error
	Destroy(ctx *cli.Context) error
}

type CompositeComponent struct {
	name    string
	init    InitFunc
	destroy DestroyFunc
}

type InitFunc func(ctx *cli.Context) error
type DestroyFunc func(ctx *cli.Context) error

func NewComponent(name string) *CompositeComponent {
	return &CompositeComponent{
		name: name,
	}
}

func (c *CompositeComponent) SetInit(f InitFunc) *CompositeComponent {
	c.init = f
	return c
}

func (c *CompositeComponent) SetDestroy(f DestroyFunc) *CompositeComponent {
	c.destroy = f
	return c
}

func (c *CompositeComponent) Name() string {
	return c.name
}

func (c *CompositeComponent) Init(ctx *cli.Context) error {
	if c.init == nil {
		return nil
	}
	return c.init(ctx)
}

func (c *CompositeComponent) Destroy(ctx *cli.Context) error {
	if c.destroy == nil {
		return nil
	}
	return c.destroy(ctx)
}
