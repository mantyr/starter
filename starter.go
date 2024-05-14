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

type Starter interface {
	// Logger set customer logger
	Logger(l Logger) Starter

	// Signals
	Signals(signals ...os.Signal) Starter

	// Init ...
	Init(ctx *cli.Context, components ...Component) Starter

	// Run ...
	Run(ctx *cli.Context, f Func) Starter

	// RunServices ...
	RunServices(ctx *cli.Context, services ...Service) Starter

	// Done
	Done() <-chan struct{}

	// Stop ...
	Stop() Starter

	// Wait ...
	Wait(ctx *cli.Context) Starter

	// Error return pipeline error
	Error() error
}

type Func func(ctx *cli.Context, parent context.Context) error

type starter struct {
	context    context.Context
	cancel     context.CancelFunc
	components []Component
	services   []Service
	fail       bool
	err        error
	logger     Logger
}

func New() (Starter, error) {
	ctx, cancel := context.WithCancel(context.Background())
	return &starter{
		context: ctx,
		cancel:  cancel,
		logger:  log.Default(),
	}, nil
}

func (s *starter) Logger(l Logger) Starter {
	if s.fail {
		return s
	}
	if l == nil {
		s.err = errors.New("empty logger")
		s.fail = true
		return s
	}
	s.logger = l
	return s
}

func (s *starter) Init(ctx *cli.Context, components ...Component) Starter {
	if s.fail {
		return s
	}
	s.components = append(s.components, components...)
	for _, component := range components {
		select {
		case <-s.context.Done():
			s.logger.Println("shutdown ...")
			s.fail = true
			return s
		default:
		}
		err := component.Init(ctx)
		if err != nil {
			s.logger.Printf("init %s is error: %v", component.Name(), err)
			s.fail = true
		} else {
			s.logger.Printf("init %s is OK", component.Name())
		}
	}
	if s.fail {
		s.err = errors.New("application initialization error")
	}
	return s
}

func (s *starter) stopComponents(ctx *cli.Context) {
	for i := len(s.components); i > 0; i-- {
		component := s.components[i-1]
		if component == nil {
			continue
		}
		err := component.Destroy(ctx)
		if err != nil {
			s.logger.Printf("destroy %s is error: %v", component.Name(), err)
		} else {
			s.logger.Printf("destroy %s is OK", component.Name())
		}
	}
}

func (s *starter) Run(ctx *cli.Context, f Func) Starter {
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

func (s *starter) Error() error {
	return s.err
}

func (s *starter) RunServices(ctx *cli.Context, services ...Service) Starter {
	if s.fail {
		return s
	}
	s.services = append(s.services, services...)
	for _, service := range services {
		select {
		case <-s.context.Done():
			s.logger.Println("shutdown ...")
			s.fail = true
			return s
		default:
		}
		go s.start(ctx, service)
	}
	return s
}

func (s *starter) start(ctx *cli.Context, service Service) {
	defer s.cancel()
	err := service.Start(ctx)
	if err != nil {
		s.logger.Printf("service %s is done: %v", service.Name(), err)
	} else {
		s.logger.Printf("service %s is done", service.Name())
	}
}

func (s *starter) Signals(signals ...os.Signal) Starter {
	if s.fail {
		return s
	}
	go s.signals(signals...)
	return s
}

func (s *starter) signals(signals ...os.Signal) {
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

func (s *starter) Done() <-chan struct{} {
	return s.context.Done()
}

func (s *starter) Stop() Starter {
	s.cancel()
	return s
}

func (s *starter) Wait(ctx *cli.Context) Starter {
	if s.fail {
		s.logger.Println(s.err.Error())
		return s
	}
	<-s.context.Done()
	s.GracefulStop(ctx)
	s.stopComponents(ctx)
	return s
}

func (s *starter) GracefulStop(conf *cli.Context) {
	timeout := conf.Duration(ServicesGracefulstopTimeout)
	s.logger.Printf("Graceful shutdown ...")
	s.logger.Printf("  - %s: %v", ServicesGracefulstopTimeout, timeout)

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	go s.gracefulStop(conf, cancel)
	<-ctx.Done()
}

func (s *starter) gracefulStop(conf *cli.Context, cancel context.CancelFunc) {
	wg := &sync.WaitGroup{}
	for _, service := range s.services {
		wg.Add(1)
		go s.stopService(conf, wg, service)
	}
	wg.Wait()
	cancel()
}

func (s *starter) stopService(conf *cli.Context, wg *sync.WaitGroup, service Service) {
	defer wg.Done()
	err := service.Stop(conf)
	if err != nil {
		s.logger.Printf("service %s is stopped: %v", service.Name(), err)
	} else {
		s.logger.Printf("service %s is stopped", service.Name())
	}
}
