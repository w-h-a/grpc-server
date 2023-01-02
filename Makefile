.PHONY: compilev1
compilev1:
	protoc contracts/v1/*.proto --go_out=paths=source_relative:. --go-grpc_out=paths=source_relative:. --proto_path=.

.PHONY: clean
clean:
	go clean -testcache

.PHONY: style
style:
	goimports -l -w ./pkg
	goimports -l -w ./cmd

.PHONY: tidy
tidy:
	go mod tidy

.PHONY: test
test:
	go test -v -race ./...

.PHONY: test-debug-server
test-debug:
	cd pkg/server && go test -v -race -debug=true

LOCAL_TAG ?= v0.0.1

.PHONY: build-local-image
build-local-image:
	docker build -t github.com/w-h-a/grpc-server:$(LOCAL_TAG) .

.PHONY: create-kind
create-kind:
	kind create cluster

.PHONY: load-image
load-image:
	kind load docker-image github.com/w-h-a/grpc-server:$(LOCAL_TAG)

.PHONY: deploy-local-container
deploy-local-container:
	helm install grpc-server deploy/grpc-server

.PHONY: port-forward
port-forward:
	kubectl port-forward service/grpc-server 8400:8400

.PHONY: health-probe
health-probe:
	grpc-health-probe -addr=localhost:8400

.PHONY: evans
evans:
	evans --proto ./contracts/v1/record.proto --host 127.0.0.1 --port 8400

.PHONY: start-server
start-server:
	go run ./cmd/serve

.PHONY: teardown-local-container
teardown-local-container:
	helm uninstall grpc-server

.PHONY: delete-kind
delete-kind:
	kind delete cluster
	