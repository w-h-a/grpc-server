# GRPC Server

To run:

```bash
make build-local-image create-kind load-image deploy-local-container port-forward
```

To check if server is running:

```bash
make health-probe
```

To generate a grpc client:

```bash
make evans
```

To teardown:

```bash
make teardown-local-container delete-kind
```
