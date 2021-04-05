package errors

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/status"
)

type errorBody struct {
	Err  string            `json:"error,omitempty"`
	Data map[string]string `json:"data,omitempty"`
}

func CustomHTTPError(ctx context.Context, _ *runtime.ServeMux, marshaler runtime.Marshaler, w http.ResponseWriter, _ *http.Request, err error) {
	const fallback = `{"error": "failed to marshal error message"}`

	w.Header().Set("Content-type", marshaler.ContentType())
	w.WriteHeader(runtime.HTTPStatusFromCode(status.Code(err)))

	st := status.Convert(err)

	errBd := errorBody{
		Err: st.Message(),
	}

	for _, detail := range st.Details() {
		switch t := detail.(type) {
		case *errdetails.BadRequest:
			errBd.Data = make(map[string]string, len(t.GetFieldViolations()))
			for _, violation := range t.GetFieldViolations() {
				errBd.Data[violation.GetField()] = violation.GetDescription()
			}
		}
	}

	jErr := json.NewEncoder(w).Encode(errBd)
	if jErr != nil {
		w.Write([]byte(fallback))
	}
}
