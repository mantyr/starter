package starter

import (
	"github.com/urfave/cli/v2"
	"github.com/urfave/cli/v2/altsrc"
)

var Flags = []cli.Flag{
	altsrc.NewDurationFlag(&cli.DurationFlag{
		Name:    ServicesGracefulstopTimeout,
		Usage:   "Maximum time to gracefully stop services",
		EnvVars: []string{"SERVICES_GRACEFULSTOP_TIMEOUT"},
	}),
}
