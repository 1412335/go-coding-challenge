package cmd

import (
	"os"

	"github.com/1412335/moneyforward-go-coding-challenge/pkg/configs"
	"github.com/1412335/moneyforward-go-coding-challenge/pkg/log"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

var (
	// Used for flags.
	cfgFile string
	version string
	// service config
	cfgs *configs.ServiceConfig
	// log
	// logger log.Logger
	// cmd
	rootCmd = &cobra.Command{
		Use:   "moneyforward-go-coding-challenge",
		Short: "moneyforward-go-coding-challenge",
		Long:  `moneyforward-go-coding-challenge`,
	}
)

func logError(logger log.Factory, err error) error {
	if err != nil {
		logger.Bg().Error("Error running cmd", zap.Error(err))
	}
	return err
}

func initConfig() {
	// load config from file
	cfgs = &configs.ServiceConfig{}
	if err := configs.LoadConfig(cfgFile, cfgs); err != nil {
		log.Fatal("Load config failed", zap.Error(err))
	}
	log.Info("Load config success", zap.String("file", viper.ConfigFileUsed()), zap.Any("config", cfgs))

	if cfgs.Log != nil {
		// set default logger
		log.DefaultLogger = log.NewFactory(log.WithLevel(cfgs.Log.Level))
	}
	// // add serviceName + version to log
	// log.With(zap.String("service", cfgs.ServiceName), zap.String("version", cfgs.Version))
}

func init() {
	cobra.OnInitialize(initConfig)

	// cobra cmd bind args
	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "config file (default $HOME/config.yml)")
	rootCmd.PersistentFlags().StringVarP(&version, "version", "v", "v1", "version")

	// bind pflag
	if err := viper.BindPFlag("version", rootCmd.PersistentFlags().Lookup("version")); err != nil {
		log.Error("Bind pflag version error", zap.Error(err))
	}

	// set logger
	// logger = log.DefaultLogger.Bg()
	log.Info("Root.Init")
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatal("Execute cmd failed", zap.Error(err))
		os.Exit(-1)
	}
}
