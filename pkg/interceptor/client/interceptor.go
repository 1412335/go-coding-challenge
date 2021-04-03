package interceptor

import (
	"context"

	"github.com/1412335/moneyforward-go-coding-challenge/pkg/log"
	"google.golang.org/grpc"
)

var (
	DefaultLogger = log.DefaultLogger
)

type ClientInterceptor interface {
	// set logger
	// SetLogger(logger log.Factory)

	// client interceptor for unary request
	Unary() grpc.UnaryClientInterceptor
	unaryClientInterceptor(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error

	// client streaming interceptor
	Stream() grpc.StreamClientInterceptor
	streamClientInterceptor(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error)
}
