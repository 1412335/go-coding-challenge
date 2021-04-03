package interceptor

import (
	"context"
	"time"

	"github.com/1412335/moneyforward-go-coding-challenge/pkg/log"
	uuid "github.com/satori/go.uuid"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type SimpleClientInterceptor struct{}

var _ ClientInterceptor = (*SimpleClientInterceptor)(nil)

func NewSimpleClientInterceptor() *SimpleClientInterceptor {
	return &SimpleClientInterceptor{}
}

func (i *SimpleClientInterceptor) Log() log.Factory {
	return DefaultLogger.With(zap.String("interceptor-name", "simple"))
}

func (i *SimpleClientInterceptor) Unary() grpc.UnaryClientInterceptor {
	return i.unaryClientInterceptor
}

func (i *SimpleClientInterceptor) Stream() grpc.StreamClientInterceptor {
	return i.streamClientInterceptor
}

func (i *SimpleClientInterceptor) unaryClientInterceptor(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) (err error) {
	defer func() {
		if r := recover(); r != nil {
			i.Log().For(ctx).Error("unary req", zap.Any("panic", r))
			err = status.Error(codes.Unknown, "server error")
		}
	}()

	start := time.Now()

	// send x-request-id header
	xrid := uuid.NewV4().String()
	// header := metadata.New(map[string]string{"x-request-id": xrid})
	// APPEND HEADER RESQUEST
	ctx = metadata.AppendToOutgoingContext(ctx, []string{"x-request-id", xrid}...)

	// fetch response header
	var header metadata.MD
	opts = append(opts, grpc.Header(&header))

	// invoke request
	err = invoker(ctx, method, req, reply, cc, opts...)

	// get x-response-id header
	xrespid := header.Get("x-response-id")
	customHeader := header.Get("custom-resp-header")

	i.Log().For(ctx).Info("unary client request",
		zap.String("method", method),
		zap.String("x-request-id", xrid),
		zap.Strings("x-response-id", xrespid),
		zap.Strings("custom-resp-header", customHeader),
		zap.Duration("duration", time.Since(start)),
		zap.Any("resp", reply),
		zap.Error(err),
	)
	return err
}

func (i *SimpleClientInterceptor) streamClientInterceptor(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (clientStream grpc.ClientStream, err error) {
	defer func() {
		if r := recover(); r != nil {
			i.Log().For(ctx).Error("stream req", zap.Any("panic", r))
			err = status.Error(codes.Unknown, "server error")
		}
	}()
	start := time.Now()

	// send x-request-id header
	xrid := uuid.NewV4().String()
	// header := metadata.New(map[string]string{"x-request-id": xrid})
	// APPEND HEADER RESQUEST
	ctx = metadata.AppendToOutgoingContext(ctx, []string{"x-request-id", xrid}...)

	clientStream, err = streamer(ctx, desc, cc, method, opts...)

	// get x-response-id header
	// NOT WORK: not using stream context
	// md, ok := metadata.FromIncomingContext(clientStream.Context())
	// if !ok {
	// 	return nil, status.Errorf(codes.DataLoss, "failed to get metadata")
	// }
	// xrespid := md.Get("x-response-id")

	var xrespid []string
	var customHeader []string
	header, ok := clientStream.Header()
	if ok == nil {
		xrespid = header.Get("x-response-id")
		customHeader = header.Get("custom-resp-header")
	}

	i.Log().For(ctx).Info("stream client request",
		zap.String("method", method),
		zap.String("x-request-id", xrid),
		zap.Strings("x-response-id", xrespid),
		zap.Strings("custom-resp-header", customHeader),
		zap.Duration("duration", time.Since(start)),
		zap.Error(err),
	)
	return clientStream, err
}
