package cmd

import (
	"errors"

	"github.com/spf13/cobra"
	"go.uber.org/zap"

	grpcClient "github.com/1412335/moneyforward-go-coding-challenge/pkg/client"
	"github.com/1412335/moneyforward-go-coding-challenge/pkg/log"
	"github.com/1412335/moneyforward-go-coding-challenge/service/user/client"
)

var userClientCmd = &cobra.Command{
	Use:   "user-client",
	Short: "Start grpc client for user service",
	Long:  `Start grpc client for user service`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return UserClientService()
	},
}

func init() {
	// logger.Info("client.Init")
	rootCmd.AddCommand(userClientCmd)
}

func UserClientService() error {
	// create log factory
	logger := log.DefaultLogger.With(
		zap.String("service", cfgs.ServiceName),
		zap.String("version", cfgs.Version),
	)
	// get user client configs
	clientCfgs, ok := cfgs.ClientConfig["user"]
	if !ok {
		return logError(logger, errors.New("not found user client config"))
	}
	zapLogger := logger.With(
		zap.String("client-service", clientCfgs.ServiceName),
		zap.String("client-service-version", clientCfgs.Version),
	)

	// set default logger
	// user.DefaultLogger = zapLogger

	var opts []grpcClient.Option
	c, err := client.New(
		clientCfgs,
		opts...,
	)

	if err != nil {
		return logError(zapLogger, err)
	}
	defer c.Close()

	// login
	username, password := "string@gmail.com", "stringstring"
	if token, err := c.Login(username, password); err != nil {
		return logError(zapLogger, err)
	} else {
		zapLogger.Bg().Info("login resp", zap.String("token", token))
	}

	return nil
}
