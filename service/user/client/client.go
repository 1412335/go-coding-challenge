package client

import (
	"context"
	"time"

	"github.com/1412335/moneyforward-go-coding-challenge/pkg/api/user"
	grpcClient "github.com/1412335/moneyforward-go-coding-challenge/pkg/client"
	"github.com/1412335/moneyforward-go-coding-challenge/pkg/configs"
	interceptor "github.com/1412335/moneyforward-go-coding-challenge/pkg/interceptor/client"
	"github.com/1412335/moneyforward-go-coding-challenge/pkg/log"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type Client interface {
	Login(email, password string) (string, error)
	Close() error
}

type clientImpl struct {
	ctx           context.Context
	logger        log.Factory
	client        *grpcClient.Client
	userSrvClient user.UserServiceClient
}

func New(cfgs *configs.ClientConfig, opt ...grpcClient.Option) (Client, error) {
	opt = append(opt,
		grpcClient.WithInterceptors(interceptor.NewSimpleClientInterceptor()),
	)
	client, err := grpcClient.New(cfgs, opt...)
	if err != nil {
		return nil, err
	}

	ctx := context.Background()
	return &clientImpl{
		ctx:           ctx,
		logger:        client.GetLogger(),
		client:        client,
		userSrvClient: user.NewUserServiceClient(client.ClientConn),
	}, nil
}

func (c *clientImpl) Close() error {
	return c.client.Close()
}

func (c *clientImpl) setHeader(m map[string]string) context.Context {
	md := metadata.New(m)
	ctx := metadata.NewOutgoingContext(c.ctx, md)
	return ctx
}

// login & get token
func (c *clientImpl) Login(email, password string) (string, error) {
	ctx := c.setHeader(map[string]string{"custom-req-header": "login"})
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	// prepare request
	msg := &user.LoginRequest{
		Email:    email,
		Password: password,
	}
	// fetch response headers
	var header metadata.MD
	// call service
	reply, err := c.userSrvClient.Login(ctx, msg, grpc.Header(&header))
	if err != nil {
		c.logger.For(ctx).Error("login failed", zap.Error(err))
		return "", err
	}
	// get response headers
	token := header.Get("token")
	if len(token) > 0 {
		c.logger.For(ctx).Info("response headers", zap.Strings("token", token))
	}
	// return
	return reply.GetToken(), nil
}
