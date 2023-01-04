FROM golang:1.19-alpine AS build
WORKDIR /go/src/grpc-server
COPY . .
RUN CGO_ENABLED=0 go build -o /go/bin/grpc-server ./cmd/serve
RUN GRPC_HEALTH_PROBE_VERSION=v0.4.13 && wget -qO/go/bin/grpc-health-probe https://github.com/grpc-ecosystem/grpc-health-probe/releases/download/${GRPC_HEALTH_PROBE_VERSION}/grpc_health_probe-linux-amd64 && chmod +x /go/bin/grpc-health-probe

FROM alpine
COPY --from=build /go/bin/grpc-server /bin/grpc-server
COPY --from=build /go/bin/grpc-health-probe /bin/grpc-health-probe
ENTRYPOINT [ "/bin/grpc-server" ]