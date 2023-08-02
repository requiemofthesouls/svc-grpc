package grpc

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/requiemofthesouls/logger"

	"github.com/requiemofthesouls/svc-grpc/server"
)

type (
	Manager interface {
		Start(name string) error
		Stop(name string) error
		StartAll(ctx context.Context)
	}
	manager struct {
		l             logger.Wrapper
		servers       map[string]server.Server
		stopChans     map[string]chan struct{}
		stopHTTPChans map[string]chan struct{}
	}
)

func NewManager(servers map[string]server.Server, l logger.Wrapper) (Manager, error) {
	return &manager{
		l:             l,
		servers:       servers,
		stopChans:     make(map[string]chan struct{}),
		stopHTTPChans: make(map[string]chan struct{}),
	}, nil
}

func (m *manager) Start(name string) error {
	var (
		s  server.Server
		ok bool
	)
	if s, ok = m.servers[name]; !ok {
		return fmt.Errorf("unknown server '%s'", name)
	}

	m.stopChans[name] = make(chan struct{})
	go func(s server.Server, stopChan chan struct{}, l logger.Wrapper) {
		l.Info(fmt.Sprintf("Start GRPC server '%s'", name))
		if err := s.Start(); err != nil {
			l.Error("error starting GRPC server", logger.Error(err))
		}
		stopChan <- struct{}{}
	}(s, m.stopChans[name], m.l)

	if s.GetHTTPGateway() != nil {
		m.stopHTTPChans[name] = make(chan struct{})
		go func(s server.Server, stopChan chan struct{}, l logger.Wrapper) {
			l.Info(fmt.Sprintf("Start HTTP gateway server '%s'", name))
			if err := s.GetHTTPGateway().Start(); err != nil {
				l.Error("error starting HTTP gateway server", logger.Error(err))
			}
			stopChan <- struct{}{}
		}(s, m.stopHTTPChans[name], m.l)
	}

	return nil
}

func (m *manager) Stop(name string) error {
	var (
		s  server.Server
		ok bool
	)
	if s, ok = m.servers[name]; !ok {
		return fmt.Errorf("unknown server '%s'", name)
	}

	if s.GetHTTPGateway() != nil && s.GetHTTPGateway().IsStarted() {
		m.l.Info(fmt.Sprintf("Stop HTTP gateway server '%s'", name))
		if err := s.GetHTTPGateway().Stop(); err != nil {
			return err
		}

		select {
		case <-time.After(time.Second * 5):
			return errors.New("couldn't stop http gateway server within the specified timeout (5 sec)")
		case <-m.stopHTTPChans[name]:
		}
	}

	if !s.IsStarted() {
		return nil
	}

	m.l.Info(fmt.Sprintf("Stop GRPC server '%s'", name))
	s.Stop()

	select {
	case <-time.After(time.Second * 5):
		return errors.New("couldn't stop grpc server within the specified timeout (5 sec)")
	case <-m.stopChans[name]:
		return nil
	}
}

func (m *manager) StartAll(ctx context.Context) {
	for name := range m.servers {
		if err := m.Start(name); err != nil {
			m.l.Error("error starting GRPC server: "+name, logger.Error(err))
		}
	}

	<-ctx.Done()

	for name := range m.servers {
		if err := m.Stop(name); err != nil {
			m.l.Error("error stopping GRPC server: "+name, logger.Error(err))
		}
	}
}
