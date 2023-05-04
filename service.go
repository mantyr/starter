package starter

import (
	"github.com/urfave/cli/v2"
)

type Service interface {
	Name() string
	Start(ctx *cli.Context) error
	Stop(ctx *cli.Context) error
}
