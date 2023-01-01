package agent

import (
	"net"
	"sync"

	"github.com/w-h-a/grpc-server/pkg/log"
	"github.com/w-h-a/grpc-server/pkg/server"
	"google.golang.org/grpc"
)

type Agent struct {
	Config Config

	log    *log.Log
	server *grpc.Server

	shutdown     bool
	shutdownLock sync.Mutex
}

func NewAgent(config Config) (*Agent, error) {
	a := &Agent{
		Config: config,
	}

	setup := []func() error{
		a.setupLog,
		a.setupServer,
	}

	for _, fn := range setup {
		err := fn()
		if err != nil {
			return nil, err
		}
	}

	return a, nil
}

func (a *Agent) setupLog() error {
	var err error
	a.log, err = log.NewLog()
	return err
}

func (a *Agent) setupServer() error {
	var err error

	serverConfig := &server.Config{
		CommitLog: a.log,
	}

	a.server, err = server.NewGRPCServer(serverConfig)
	if err != nil {
		return err
	}

	rpcAddr, err := a.Config.RPCAddr()
	if err != nil {
		return err
	}

	listener, err := net.Listen("tcp", rpcAddr)
	if err != nil {
		return err
	}

	go func() {
		err := a.server.Serve(listener)
		if err != nil {
			_ = a.Shutdown()
		}
	}()

	return err
}

func (a *Agent) Shutdown() error {
	a.shutdownLock.Lock()
	defer a.shutdownLock.Unlock()

	if a.shutdown {
		return nil
	}

	a.shutdown = true

	shutdowns := []func() error{
		func() error {
			a.server.GracefulStop()
			return nil
		},
	}

	for _, fn := range shutdowns {
		err := fn()
		if err != nil {
			return err
		}
	}

	return nil
}
