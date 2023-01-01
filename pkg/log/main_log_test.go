package log

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	contracts "github.com/w-h-a/grpc-server/contracts/v1"
)

const (
	numOfWrites uint64 = 3
)

func TestLog(t *testing.T) {
	tests := make(map[string]func(t *testing.T, log *Log))
	tests["append and read"] = testAppendRead
	tests["index out of range"] = testIndexOutOfRange

	for situation, fn := range tests {
		t.Run(situation, func(t *testing.T) {
			log, err := NewLog()
			require.NoError(t, err)
			fn(t, log)
		})
	}
}

func testAppendRead(t *testing.T, log *Log) {
	for i := uint64(0); i < numOfWrites; i++ {
		value := fmt.Sprintf("hello world %v", i)

		current, err := log.Append(&contracts.Record{Value: value})
		require.NoError(t, err)
		require.Equal(t, i, current)

		record, err := log.Read(current)
		require.NoError(t, err)
		require.Equal(t, record.Value, value)
	}
}

func testIndexOutOfRange(t *testing.T, log *Log) {
	record, err := log.Read(numOfWrites)
	require.Nil(t, record)
	require.Error(t, err)
}
