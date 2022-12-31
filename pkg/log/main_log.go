package log

import (
	"sync"

	contracts "github.com/w-h-a/grpc-server/contracts/v1"
)

type Log struct {
	mu      sync.Mutex
	records []*contracts.Record
}

func NewLog() (*Log, error) {
	return &Log{}, nil
}

func (l *Log) Append(record *contracts.Record) (uint64, error) {
	l.mu.Lock()
	defer l.mu.Unlock()

	record.Index = uint64(len(l.records))

	l.records = append(l.records, record)

	return record.Index, nil
}

func (l *Log) Read(index uint64) (*contracts.Record, error) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if index >= uint64(len(l.records)) {
		return nil, contracts.ErrIndexOutOfRange{Index: index}
	}

	return l.records[index], nil
}
