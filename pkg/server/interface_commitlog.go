package server

import contracts "github.com/w-h-a/grpc-server/contracts/v1"

type CommitLog interface {
	Append(*contracts.Record) (uint64, error)
	Read(uint64) (*contracts.Record, error)
}
