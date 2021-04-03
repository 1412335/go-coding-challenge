package interceptor

import (
	"context"

	"github.com/1412335/moneyforward-go-coding-challenge/pkg/log"
	"google.golang.org/grpc"
)

var (
	DefaultLogger = log.DefaultLogger
)

type ServerInterceptor interface {
	// unary request to grpc server
	Unary() grpc.UnaryServerInterceptor
	UnaryInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error)

	// stream request interceptor
	Stream() grpc.StreamServerInterceptor
	StreamInterceptor(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) (err error)
}
