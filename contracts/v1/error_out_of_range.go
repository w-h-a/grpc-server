package record_v1

import (
	"fmt"

	"google.golang.org/genproto/googleapis/rpc/errdetails"
	status "google.golang.org/grpc/status"
)

type ErrIndexOutOfRange struct {
	Index uint64
}

func (e ErrIndexOutOfRange) Error() string {
	return e.GRPCStatus().Err().Error()
}

func (e ErrIndexOutOfRange) GRPCStatus() *status.Status {
	status := status.New(404, fmt.Sprintf("index is out of range: %d", e.Index))
	
	msg := fmt.Sprintf("The requested index is outside of the log's range: %d", e.Index)

	details := &errdetails.LocalizedMessage{
		Locale: "en-US",
		Message: msg,
	}

	detailedStatus, err := status.WithDetails(details)
	if err != nil {
		return status
	}

	return detailedStatus
}