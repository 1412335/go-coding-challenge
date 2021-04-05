package configs

import (
	"time"

	"github.com/spf13/viper"
)

const (
	ConfigName = "config"
	ConfigType = "yml"
	ConfigPath = "."
)

type ServiceConfig struct {
	Version     string
	ServiceName string
	GRPC        *GRPC
	Proxy       *Proxy
	// client services
	ClientConfig map[string]*ClientConfig
	Database     *Database
	JWT          *JWT
	// tls secure service
	EnableTLS bool
	TLSCert   *TLSCert
	// log factory
	Log *Log
	//
	AuthRequiredMethods map[string]bool
}

type ClientConfig struct {
	Version     string
	ServiceName string
	// grpc-server
	GRPC *GRPC
	// tls secure service
	EnableTLS bool
	TLSCert   *TLSCert
}

type TLSCert struct {
	CACert  string
	CertPem string
	KeyPem  string
}

// grpc-server
type GRPC struct {
	Host               string
	Port               int
	MaxCallRecvMsgSize int
	MaxCallSendMsgSize int
}

// grpc-gateway proxy
type Proxy struct {
	Port int
}

// json web token
type JWT struct {
	Issuer    string
	SecretKey string
	Duration  time.Duration
}

// Database config
type Database struct {
	Host           string
	Port           string
	User           string
	Password       string
	Scheme         string
	Debug          bool
	MaxIdleConns   int
	MaxOpenConns   int
	ConnectTimeout time.Duration
}

// Log config
type Log struct {
	Mode        string
	Level       string
	TraceLevel  string
	IsLogFile   bool
	PathLogFile string
}

func LoadConfig(cfgFile string, cfg interface{}) error {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Search config in home directory with name ".cobra" (without extension).
		viper.SetConfigName(ConfigName)
		viper.SetConfigType(ConfigType)
		viper.AddConfigPath(ConfigPath)
	}
	viper.AutomaticEnv()
	// Find and read the config file
	if err := viper.ReadInConfig(); err != nil {
		return err
	}
	if err := viper.Unmarshal(cfg); err != nil {
		return err
	}
	return nil
}
