package main

import (
	"context"
	"flag"
	"fmt"
	"log"

	contracts "github.com/w-h-a/grpc-server/contracts/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type arrayFlags []string

func (a *arrayFlags) String() string {
	return "my string representation"
}

func (a *arrayFlags) Set(value string) error {
	*a = append(*a, value)
	return nil
}

func main() {
	addr := flag.String("addr", "127.0.0.1:8400", "service address")
	var myFlags arrayFlags
	flag.Var(&myFlags, "value", "add multiple values to the log")
	flag.Parse()

	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}
	conn, err := grpc.Dial(*addr, opts...)
	if err != nil {
		log.Fatal(err)
	}
	client := contracts.NewEndpointsClient(conn)

	ctx := context.Background()
	fmt.Println("indices:")

	produceStream, err := client.ProduceStream(ctx)
	if err != nil {
		log.Fatal(err)
	}

	for _, value := range myFlags {
		err = produceStream.Send(&contracts.ProduceRequest{Record: &contracts.Record{Value: value}})
		if err != nil {
			log.Fatal(err)
		}

		response, err := produceStream.Recv()
		if err != nil {
			log.Fatal(err)
		}

		fmt.Printf("\t- %v\n", response.Index)
	}

}
