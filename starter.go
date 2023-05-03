package starter

import (
	"context"
	"errors"
	"log"
	"os"
	"os/signal"
	"sync"

	"github.com/urfave/cli/v2"
)

type Starter struct {
	context  context.Context
	cancel   context.CancelFunc
	services []Service
	fail     bool
	err      error
}

func New() (*Starter, error) {
	ctx, cancel := context.WithCancel(context.Background())
	return &Starter{
		context: ctx,
		cancel:  cancel,
	}, nil
}

func (s *Starter) Init(ctx *cli.Context, components ...Component) *Starter {
	if s.fail {
		return s
	}
	for _, component := range components {
		select {
		case <-s.context.Done():
			log.Println("shutdown ...")
			s.fail = true
			return s
		default:
		}
		err := component.Init(ctx)
		if err != nil {
			log.Printf("init %s is error: %v", component.Name(), err)
			s.fail = true
		} else {
			log.Printf("init %s is OK", component.Name())
		}
	}
	if s.fail {
		s.err = errors.New("application initialization error")
	}
	return s
}

type Func func(ctx *cli.Context, parent context.Context) error

func (s *Starter) Run(ctx *cli.Context, f Func) *Starter {
	if s.fail {
		return s
	}
	err := f(ctx, s.context)
	if err != nil {
		s.fail = true
		s.err = err
	}
	return s
}

func (s *Starter) Error() error {
	return s.err
}

func (s *Starter) RunServices(ctx *cli.Context, services ...Service) *Starter {
	if s.fail {
		return s
	}
	s.services = append(s.services, services...)
	for _, service := range services {
		select {
		case <-s.context.Done():
			log.Println("shutdown ...")
			s.fail = true
			return s
		default:
		}
		go s.start(ctx, service)
	}
	return s
}

func (s *Starter) start(ctx *cli.Context, service Service) {
	defer s.cancel()
	err := service.Start(ctx)
	if err != nil {
		log.Printf("service %s is done: %v", service.Name(), err)
	} else {
		log.Printf("service %s is done", service.Name())
	}
}

func (s *Starter) Signals(signals ...os.Signal) *Starter {
	if s.fail {
		return s
	}
	go s.signals(signals...)
	return s
}

func (s *Starter) signals(signals ...os.Signal) {
	sigs := make(chan os.Signal, 1)
	signal.Notify(
		sigs,
		signals...,
	)
	for {
		select {
		case <-s.context.Done():
			return
		case event := <-sigs:
			for _, item := range signals {
				if item == event {
					s.cancel()
					return
				}
			}
		}
	}
}

func (s *Starter) Done() <-chan struct{} {
	return s.context.Done()
}

func (s *Starter) Wait(ctx *cli.Context) {
	if s.fail {
		log.Println(s.err.Error())
		return
	}
	<-s.context.Done()
	s.GracefulStop(ctx)
}

func (s *Starter) GracefulStop(conf *cli.Context) {
	timeout := conf.Duration(ServicesGracefulstopTimeout)
	log.Printf("Graceful shutdown ...")
	log.Printf("  - %s: %v", ServicesGracefulstopTimeout, timeout)

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	go s.gracefulStop(conf, cancel)
	<-ctx.Done()
}

func (s *Starter) gracefulStop(conf *cli.Context, cancel context.CancelFunc) {
	wg := &sync.WaitGroup{}
	for _, service := range s.services {
		wg.Add(1)
		go s.stopService(conf, wg, service)
	}
	wg.Wait()
	cancel()
}

func (s *Starter) stopService(conf *cli.Context, wg *sync.WaitGroup, service Service) {
	defer wg.Done()
	err := service.Stop(conf)
	if err != nil {
		log.Printf("service %s is stopped: %v", service.Name(), err)
	} else {
		log.Printf("service %s is stopped", service.Name())
	}
}
