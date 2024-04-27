# Starter

[![Build Status](https://travis-ci.org/mantyr/starter.svg?branch=master)](https://travis-ci.org/mantyr/starter)
[![GoDoc](https://godoc.org/github.com/mantyr/starter?status.png)](http://godoc.org/github.com/mantyr/starter)
[![Go Report Card](https://goreportcard.com/badge/github.com/mantyr/starter?v=1)][goreport]
[![Software License](https://img.shields.io/badge/license-MIT-brightgreen.svg)](LICENSE.md)

This stable version.

## Description

This is a package for organizing the launch of services inside the application with convenient initialization of components.

- The component can be initialized
- The services can be started and stopped

```go
import (
	"github.com/urfave/cli/v2"
)

type Component interface {
	Name() string
	Init(ctx *cli.Context) error
	Destroy(ctx *cli.Context) error
}

type Service interface {
	Name() string
	Start(ctx *cli.Context) error
	Stop(ctx *cli.Context) error
}
```

### Supports

- [x] Signals for graceful shutdown
- [x] Init functions
- [x] Components
- [x] Service startup
- [x] Customer logger


## Installation

    $ go get github.com/mantyr/starter

## Example

```go
package main

import (
	"syscall"

	"github.com/urfave/cli/v2"
	"github.com/mantyr/starter"

	"service1"
	"service2"
)

func main() {
	var ctx cli.Context
	ctx.Set("server.gracefulstop.duration", "30m")

	s1, err := service1.New()
	if err != nil {
		panic(err)
	}
	s2, err := service2.New()
	if err != nil {
		panic(err)
	}

	s := starter.New()
	s.Signals(
		syscall.SIGINT,
		syscall.SIGTERM,
	).Init(
		ctx,
		db,
		s1,
		s2,
		starter.NewComponent("service3").SetInit(func() error{return nil}).SetDestroy(func() error{return nil}),
	).RunServices(
		ctx,
		s1,
		s2,
	).Wait()
```

## Author

[Oleg Shevelev][mantyr]

[mantyr]: https://github.com/mantyr

[build_status]: https://travis-ci.org/mantyr/starter
[godoc]:        http://godoc.org/github.com/mantyr/starter
[goreport]:     https://goreportcard.com/report/github.com/mantyr/starter
