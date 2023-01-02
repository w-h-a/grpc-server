package server

import (
	"context"

	contracts "github.com/w-h-a/grpc-server/contracts/v1"
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
