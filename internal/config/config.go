package config

import (
	"fmt"
	"strconv"

	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

var ErrConfigRead = errors.New("unable to read config file")

type Config struct {
	Logger LoggerConf
	HTTP   HTTPConf
	AWS    AWSConf
}

type LoggerConf struct {
	Level   string
	File    string
	Size    int
	Backups int
	Age     int
}

type HTTPConf struct {
	Host                   string
	Port                   string
	Secret                 string
	FaceComparisonRouteTpl string
}

type AWSConf struct {
	AccessKeyId         string
	SecretAccessKey     string
	Region              string
	SimilarityThreshold float64
}

func New(path string) (*Config, error) {
	viper.SetConfigFile(path)

	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("%w: %s", ErrConfigRead, path)
	}

	st, err := strconv.Atoi(viper.GetString("aws.similarity_threshold"))
	if err != nil {
		st = 80
	}

	return &Config{
		LoggerConf{
			viper.GetString("logger.level"),
			viper.GetString("logger.file"),
			viper.GetInt("logger.size"),
			viper.GetInt("logger.backups"),
			viper.GetInt("logger.age"),
		},
		HTTPConf{
			viper.GetString("http.host"),
			viper.GetString("http.port"),
			viper.GetString("http.secret"),
			viper.GetString("http.face_comparison_route_tpl"),
		},
		AWSConf{
			viper.GetString("aws.access_key_id"),
			viper.GetString("aws.secret_access_key"),
			viper.GetString("aws.region"),
			float64(st),
		},
	}, nil
}

func (c *Config) GetLoggerLevel() string {
	return c.Logger.Level
}

func (c *Config) GetLoggerFile() string {
	return c.Logger.File
}

func (c *Config) GetHTTPHost() string {
	return c.HTTP.Host
}

func (c *Config) GetHTTPPort() string {
	return c.HTTP.Port
}

func (c *Config) GetSecret() string {
	return c.HTTP.Secret
}

func (c *Config) GetFaceComparisonRouteTpl() string {
	return c.HTTP.FaceComparisonRouteTpl
}

func (c *Config) GetAccessKeyId() string {
	return c.AWS.AccessKeyId
}

func (c *Config) GetSecretAccessKey() string {
	return c.AWS.SecretAccessKey
}

func (c *Config) GetRegion() string {
	return c.AWS.Region
}

func (c *Config) GetSimilarityThreshold() float64 {
	return c.AWS.SimilarityThreshold
}
