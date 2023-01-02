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

If you want to run evans while seeing the server's logs, currently you need to run the server outside of the k8s cluster:

```bash
make start-server
```

To teardown:

```bash
make teardown-local-container delete-kind
```
