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

func main() {
	addr := flag.String("addr", "127.0.0.1:8400", "service address")
	index := flag.Uint64("index", uint64(0), "index at which to initially read from log")
	flag.Parse()
	
	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}
	conn, err := grpc.Dial(*addr, opts...)
	if err != nil {
		log.Fatal(err)
	}
	client := contracts.NewEndpointsClient(conn)

	ctx := context.Background()
	fmt.Println("values:")

	{
		consumeStream, err := client.ConsumeStream(ctx, &contracts.ConsumeRequest{Index: *index})
		if err != nil {
			log.Fatal(err)
		}

		for {
			response, err := consumeStream.Recv()
			if err != nil {
				log.Fatal(err)
			}

			fmt.Printf("\t- %v\n", response.Record.Value)
		}
	}
}