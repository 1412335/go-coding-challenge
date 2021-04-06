package user

import (
	"context"
	"strings"

	pb "github.com/1412335/moneyforward-go-coding-challenge/pkg/api/user"
	interceptor "github.com/1412335/moneyforward-go-coding-challenge/pkg/interceptor/server"
	"github.com/1412335/moneyforward-go-coding-challenge/pkg/log"
	"go.uber.org/zap"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// Auth interceptor with JWT
type AuthServerInterceptor struct {
	jwtManager          *TokenService
	authRequiredMethods map[string]bool
}

var _ interceptor.ServerInterceptor = (*AuthServerInterceptor)(nil)

func NewAuthServerInterceptor(jwtManager *TokenService, authRequiredMethods map[string]bool) interceptor.ServerInterceptor {
	return &AuthServerInterceptor{
		jwtManager:          jwtManager,
		authRequiredMethods: authRequiredMethods,
	}
}

func (a *AuthServerInterceptor) Log() log.Factory {
	return interceptor.DefaultLogger.With(zap.String("interceptor-name", "auth"))
}

func (a *AuthServerInterceptor) Unary() grpc.UnaryServerInterceptor {
	return a.UnaryInterceptor
}
func (a *AuthServerInterceptor) Stream() grpc.StreamServerInterceptor {
	return a.StreamInterceptor
}

// check accessiable method with user role got from header authorization
func (a *AuthServerInterceptor) authorize(ctx context.Context, method string, req interface{}) error {
	// check accessiable method with user role got from header authorization
	authReq, ok := a.authRequiredMethods[method]
	if !authReq || !ok {
		return nil
	}
	a.Log().For(ctx).Info("authorize", zap.Any("req", req))

	// fetch authorization header
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return status.Errorf(codes.DataLoss, "failed to get metadata")
	}
	accessToken := md.Get("authorization")
	if len(accessToken) == 0 {
		return status.Errorf(codes.InvalidArgument, "missing 'authorization' header")
	}
	if strings.Trim(accessToken[0], " ") == "" {
		return status.Errorf(codes.InvalidArgument, "empty 'authorization' header")
	}
	a.Log().For(ctx).Info("authorize", zap.String("token", accessToken[0]))

	// verify token
	userClaims, err := a.jwtManager.Verify(accessToken[0])
	if err != nil {
		return status.Errorf(codes.Unauthenticated, "verify token failed: %v", err)
	}
	ErrAccessDined := status.Errorf(codes.Unauthenticated, "access denied")
	switch method {
	case "/user.UserService/UpdateUser":
		if userClaims.ID != req.(*pb.UpdateUserRequest).GetUser().GetId() {
			return ErrAccessDined
		}
	case "/user.UserService/DeleteUser":
		if userClaims.ID != req.(*pb.DeleteUserRequest).GetId() {
			return ErrAccessDined
		}
	case "/user.UserService/CreateAccount":
		if userClaims.ID != req.(*pb.CreateAccountRequest).GetUserId() {
			return ErrAccessDined
		}
	case "/user.UserService/ListAccounts":
		if userClaims.ID != req.(*pb.ListAccountsRequest).GetUserId().Value {
			return ErrAccessDined
		}
	case "/user.UserService/CreateTransaction":
		if userClaims.ID != req.(*pb.CreateTransactionRequest).GetUserId() {
			return ErrAccessDined
		}
	case "/user.UserService/ListTransactions":
		if userClaims.ID != req.(*pb.ListTransactionsRequest).GetUserId() {
			return ErrAccessDined
		}
	}
	return nil
}

// unary request to grpc server
func (a *AuthServerInterceptor) UnaryInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
	defer func() {
		if r := recover(); r != nil {
			a.Log().For(ctx).Error("unary req", zap.Any("panic", r))
			err = status.Error(codes.Unknown, "Internal server error")
		}
	}()
	a.Log().For(ctx).Info("unary req", zap.String("method", info.FullMethod))

	// authorize request
	err = a.authorize(ctx, info.FullMethod, req)
	if err != nil {
		return nil, err
	}

	// check request timeout or canceled by the client
	if ctx.Err() == context.Canceled {
		return nil, status.Error(codes.Canceled, "request is canceled")
	}
	if ctx.Err() == context.DeadlineExceeded {
		return nil, status.Error(codes.DeadlineExceeded, "deadline is exceeded")
	}

	//
	return handler(ctx, req)
}

// stream request interceptor
func (a *AuthServerInterceptor) StreamInterceptor(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) (err error) {
	defer func() {
		if r := recover(); r != nil {
			a.Log().For(ss.Context()).Error("stream req", zap.Any("panic", r))
			err = status.Error(codes.Unknown, "Internal server error")
		}
	}()
	a.Log().For(ss.Context()).Info("stream req", zap.String("method", info.FullMethod), zap.Any("serverStream", info.IsServerStream))

	// authorize request
	err = a.authorize(ss.Context(), info.FullMethod, nil)
	if err != nil {
		return err
	}
	//
	return handler(srv, ss)
}
