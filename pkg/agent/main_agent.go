package agent

import (
	"fmt"
	"net"
	"os"
	"sync"
	"time"

	"github.com/w-h-a/grpc-server/pkg/log"
	"github.com/w-h-a/grpc-server/pkg/server"
	"go.opencensus.io/examples/exporter"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type Agent struct {
	Config Config

	log               *log.Log
	telemetryExporter *exporter.LogExporter
	server            *grpc.Server

	shutdown     bool
	shutdownLock sync.Mutex
}

func NewAgent(config Config) (*Agent, error) {
	a := &Agent{
		Config: config,
	}

	setup := []func() error{
		a.setupLogger,
		a.setupLog,
		a.setupTelemetryExporter,
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

func (a *Agent) setupLogger() error {
	logger, err := zap.NewDevelopment()
	if err != nil {
		return err
	}

	zap.ReplaceGlobals(logger)

	return nil
}

func (a *Agent) setupLog() error {
	var err error
	a.log, err = log.NewLog()
	return err
}

func (a *Agent) setupTelemetryExporter() error {
	var err error

	metricsLogFile, err := os.CreateTemp("", "metrics-*.log")
	if err != nil {
		return err
	}

	fmt.Printf("metrics log file: %s", metricsLogFile.Name())

	tracesLogFile, err := os.CreateTemp("", "traces-*.log")
	if err != nil {
		return err
	}

	fmt.Printf("traces log file: %s", tracesLogFile.Name())

	a.telemetryExporter, err = exporter.NewLogExporter(exporter.Options{
		MetricsLogFile:    metricsLogFile.Name(),
		TracesLogFile:     tracesLogFile.Name(),
		ReportingInterval: time.Second,
	})
	if err != nil {
		return err
	}

	return a.telemetryExporter.Start()
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
		func() error {
			if a.telemetryExporter != nil {
				time.Sleep(1500 * time.Millisecond)
				a.telemetryExporter.Stop()
				a.telemetryExporter.Close()
			}
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
