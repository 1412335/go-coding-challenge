package cmd

import (
	grpcClient "github.com/1412335/moneyforward-go-coding-challenge/pkg/client"
	"github.com/1412335/moneyforward-go-coding-challenge/pkg/configs"
	"github.com/1412335/moneyforward-go-coding-challenge/pkg/log"
	"github.com/1412335/moneyforward-go-coding-challenge/service/user"
	"github.com/1412335/moneyforward-go-coding-challenge/service/user/client"

	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

var userCmd = &cobra.Command{
	Use:   "user-service",
	Short: "Start User Service v1",
	Long:  `Start User Service v1`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return userService()
	},
}

func init() {
	log.Info("service.user.Init")
	rootCmd.AddCommand(userCmd)
}

func userService() error {
	// create log factory
	zapLogger := log.With(zap.String("service", cfgs.ServiceName), zap.String("version", cfgs.Version))

	// server
	server := user.NewServer(
		cfgs,
	)

	// run grpc server
	// return logError(zapLogger, server.Run())
	go func() {
		logError(zapLogger, server.Run())
	}()

	go func() {
		err := testGrpcClient(cfgs.ClientConfig["user"])
		if err != nil {
			logError(zapLogger, err)
		}
	}()

	// run grpc-gateway
	handler := user.NewHandler(cfgs)
	err := handler.Run()
	if err != nil {
		zapLogger.Error("Starting gRPC-gateway error", zap.Error(err))
	}
	return err
}

func testGrpcClient(cfgs *configs.ClientConfig) error {
	var opts []grpcClient.Option
	c, err := client.New(
		cfgs,
		opts...,
	)
	if err != nil {
		return err
	}
	defer c.Close()

	// login
	username, password := "abc@gmail.com", "stringstring"
	token, err := c.Login(username, password)
	if err != nil {
		return err
	}
	log.Info("login resp", zap.String("token", token))
	return nil
}
