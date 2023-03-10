package server

import (
	"context"
	"time"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_zap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap"
	grpc_ctxtags "github.com/grpc-ecosystem/go-grpc-middleware/tags"
	contracts "github.com/w-h-a/grpc-server/contracts/v1"
	"go.opencensus.io/plugin/ocgrpc"
	"go.opencensus.io/stats/view"
	"go.opencensus.io/trace"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	healthsrv "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
)

type grpcServer struct {
	contracts.UnimplementedEndpointsServer
	Config *Config
}

func NewGRPCServer(config *Config, opts ...grpc.ServerOption) (*grpc.Server, error) {
	logger := zap.L().Named("server")
	zapOpts := []grpc_zap.Option{
		grpc_zap.WithDurationField(
			func(duration time.Duration) zapcore.Field {
				return zap.Int64("grpc.time_ns", duration.Nanoseconds())
			},
		),
	}

	err := view.Register(ocgrpc.DefaultServerViews...)
	if err != nil {
		return nil, err
	}

	trace.ApplyConfig(trace.Config{DefaultSampler: trace.AlwaysSample()})

	opts = append(opts,
		grpc.StreamInterceptor(
			grpc_middleware.ChainStreamServer(
				grpc_ctxtags.StreamServerInterceptor(),
				grpc_zap.StreamServerInterceptor(logger, zapOpts...),
			),
		),
		grpc.UnaryInterceptor(
			grpc_middleware.ChainUnaryServer(
				grpc_ctxtags.UnaryServerInterceptor(),
				grpc_zap.UnaryServerInterceptor(logger, zapOpts...),
			),
		),
		grpc.StatsHandler(&ocgrpc.ServerHandler{}),
	)

	gsrv := grpc.NewServer(opts...)

	hsrv := health.NewServer()
	hsrv.SetServingStatus("", healthsrv.HealthCheckResponse_SERVING)
	healthsrv.RegisterHealthServer(gsrv, hsrv)

	srv := &grpcServer{Config: config}
	contracts.RegisterEndpointsServer(gsrv, srv)

	reflection.Register(gsrv)

	return gsrv, nil
}

func (g *grpcServer) Produce(ctx context.Context, req *contracts.ProduceRequest) (*contracts.ProduceResponse, error) {
	index, err := g.Config.CommitLog.Append(req.Record)
	if err != nil {
		return nil, err
	}
	return &contracts.ProduceResponse{Index: index}, nil
}

func (g *grpcServer) Consume(ctx context.Context, req *contracts.ConsumeRequest) (*contracts.ConsumeResponse, error) {
	record, err := g.Config.CommitLog.Read(req.Index)
	if err != nil {
		return nil, err
	}
	return &contracts.ConsumeResponse{Record: record}, nil
}

func (g *grpcServer) ProduceStream(stream contracts.Endpoints_ProduceStreamServer) error {
	for {
		req, err := stream.Recv()
		if err != nil {
			return err
		}

		res, err := g.Produce(stream.Context(), req)
		if err != nil {
			return err
		}

		err = stream.Send(res)
		if err != nil {
			return err
		}
	}
}

func (g *grpcServer) ConsumeStream(req *contracts.ConsumeRequest, stream contracts.Endpoints_ConsumeStreamServer) error {
	for {
		select {
		case <-stream.Context().Done():
			return nil
		default:
			res, err := g.Consume(stream.Context(), req)
			switch err.(type) {
			case nil:
			case contracts.ErrIndexOutOfRange:
				continue
			default:
				return err
			}

			err = stream.Send(res)
			if err != nil {
				return err
			}

			req.Index++
		}
	}
}
