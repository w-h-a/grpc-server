# GRPC Server

To run:

```bash
make build-local-image create-kind load-image deploy-local-container port-forward
```

To check if server is running:

```bash
make health-probe
```

To run consume stream:

```bash
go run ./cmd/consume_stream
```

To produce/produce stream/consume:

```bash
go run ./cmd/produce
go run ./cmd/produce_stream -value "foo" -value "bar" -value <string>
go run ./cmd/consume
```

For help, run any of the above `go run` commands with `-h` appended to the end.

To teardown:

```bash
make teardown-local-container delete-kind
```
