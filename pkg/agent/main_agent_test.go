package agent

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/travisjeffery/go-dynaport"
	contracts "github.com/w-h-a/grpc-server/contracts/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

func TestAgent(t *testing.T) {
	tests := make(map[string]func(t *testing.T, client contracts.EndpointsClient))

	tests["consume beyond range"] = testConsumeBeyondRange
	tests["produce and consume requests"] = testProduceConsume

	for situation, fn := range tests {
		t.Run(situation, func(t *testing.T) {
			client, teardown := setupTest(t)
			defer teardown()

			fn(t, client)
		})
	}
}

func setupTest(t *testing.T) (client contracts.EndpointsClient, teardown func()) {
	// setup server agent
	ports := dynaport.Get(1)
	cfg := Config{
		RPCHost: "127.0.0.1",
		RPCPort: ports[0],
	}

	agent, err := NewAgent(cfg)
	require.NoError(t, err)

	// setup client(s)
	clientConn, client := createNewClient(t, agent)

	// return
	return client, func() {
		clientConn.Close()
		err := agent.Shutdown()
		require.NoError(t, err)
	}
}

func createNewClient(t *testing.T, agent *Agent) (*grpc.ClientConn, contracts.EndpointsClient) {
	rpcAddr, err := agent.Config.RPCAddr()
	require.NoError(t, err)

	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}
	conn, err := grpc.Dial(rpcAddr, opts...)
	require.NoError(t, err)

	client := contracts.NewEndpointsClient(conn)
	return conn, client
}

func testConsumeBeyondRange(t *testing.T, client contracts.EndpointsClient) {
	ctx := context.Background()

	consumeResponse, err := client.Consume(ctx, &contracts.ConsumeRequest{Index: uint64(1)})
	require.Nil(t, consumeResponse)
	require.Error(t, err)

	got := status.Code(err)
	want := status.Code(contracts.ErrIndexOutOfRange{}.GRPCStatus().Err())
	require.Equal(t, want, got)
}

func testProduceConsume(t *testing.T, client contracts.EndpointsClient) {
	ctx := context.Background()
	want := &contracts.Record{Value: "foo"}
	produceResponse, err := client.Produce(ctx, &contracts.ProduceRequest{Record: want})
	require.NoError(t, err)

	consumeResponse, err := client.Consume(ctx, &contracts.ConsumeRequest{Index: produceResponse.Index})
	require.NoError(t, err)

	require.Equal(t, want.Value, consumeResponse.Record.Value)
	require.Equal(t, want.Index, consumeResponse.Record.Index)
}