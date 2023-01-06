# GRPC Server

To run:

```bash
make build-local-image create-kind load-image deploy-local-container port-forward
```

To check if server is running:

```bash
make k8s-server-logs
make health-probe
```

To generate a grpc client:

```bash
make evans
```

For now, metrics and traces are exported to the `/tmp` directory in the container:

```bash
make exec-telemetry
```

If you want to run evans while seeing the server's logs and you don't want to run the above `k8s-server-logs` cmd:

```bash
make start-server
```

To teardown:

```bash
make teardown-local-container delete-kind
```

fix

feat
