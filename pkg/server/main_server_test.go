package server

import (
	"context"
	"net"
	"testing"

	"github.com/stretchr/testify/require"
	contracts "github.com/w-h-a/grpc-server/contracts/v1"
	"github.com/w-h-a/grpc-server/pkg/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

func TestServer(t *testing.T) {
	tests := make(map[string]func(t *testing.T, client contracts.EndpointsClient))

	tests["consume beyond range"] = testConsumeBeyondRange
	tests["produce and consume requests"] = testProduceConsume
	tests["produce and consume stream"] = testProduceConsumeStream

	for situation, fn := range tests {
		t.Run(situation, func(t *testing.T) {
			client, teardown := setupTest(t)
			defer teardown()

			fn(t, client)
		})
	}
}

func setupTest(t *testing.T) (client contracts.EndpointsClient, teardown func()) {
	// setup log
	log, err := log.NewLog()
	require.NoError(t, err)

	// setup server
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)

	cfg := &Config{CommitLog: log}

	server, err := NewGRPCServer(cfg)
	require.NoError(t, err)

	go func() {
		server.Serve(listener)
	}()

	// setup client(s)
	createNewClient := func() (*grpc.ClientConn, contracts.EndpointsClient) {
		opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}
		conn, err := grpc.Dial(listener.Addr().String(), opts...)
		require.NoError(t, err)
		client := contracts.NewEndpointsClient(conn)
		return conn, client
	}

	clientConn, client := createNewClient()

	// return
	return client, func() {
		clientConn.Close()
		server.Stop()
		listener.Close()
	}
}

func testConsumeBeyondRange(t *testing.T, client contracts.EndpointsClient) {
	ctx := context.Background()
	produceResponse, err := client.Produce(ctx, &contracts.ProduceRequest{Record: &contracts.Record{Value: "hello world"}})
	require.NoError(t, err)

	consumeResponse, err := client.Consume(ctx, &contracts.ConsumeRequest{Index: produceResponse.Index + 1})
	if consumeResponse != nil {
		t.Fatal("consume is not nil")
	}

	got := status.Code(err)
	want := status.Code(contracts.ErrIndexOutOfRange{}.GRPCStatus().Err())
	if got != want {
		t.Fatalf("got err: %v, want: %v", got, want)
	}
}

func testProduceConsume(t *testing.T, client contracts.EndpointsClient) {
	ctx := context.Background()
	want := &contracts.Record{Value: "hello world"}
	produceResponse, err := client.Produce(ctx, &contracts.ProduceRequest{Record: want})
	require.NoError(t, err)

	consumeResponse, err := client.Consume(ctx, &contracts.ConsumeRequest{Index: produceResponse.Index})
	require.NoError(t, err)

	require.Equal(t, want.Value, consumeResponse.Record.Value)
	require.Equal(t, want.Index, consumeResponse.Record.Index)
}

func testProduceConsumeStream(t *testing.T, client contracts.EndpointsClient) {
	ctx := context.Background()
	records := []*contracts.Record{
		{
			Value: "first message",
			Index: 0,
		},
		{
			Value: "second message",
			Index: 1,
		},
	}

	{
		produceStream, err := client.ProduceStream(ctx)
		require.NoError(t, err)

		for idx, record := range records {
			err = produceStream.Send(&contracts.ProduceRequest{Record: record})
			require.NoError(t, err)

			response, err := produceStream.Recv()
			require.NoError(t, err)

			if response.Index != uint64(idx) {
				t.Fatalf("got offset: %d, want: %d", response.Index, idx)
			}
		}
	}

	{
		consumeStream, err := client.ConsumeStream(ctx, &contracts.ConsumeRequest{Index: 0})
		require.NoError(t, err)

		for idx, record := range records {
			response, err := consumeStream.Recv()
			require.NoError(t, err)

			require.Equal(t, &contracts.Record{Value: record.Value, Index: uint64(idx)}, response.Record)
		}
	}
}
