package config

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/pelletier/go-toml"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
)

type Config struct {
	Appenv             string
	SbProcess          string
	SbProject          string
	Version            string
	RuntimeEnvironment string

	Environment string
	LogLevel    string
	// zerolog.Level is not unmarshalled (see https://github.com/rs/zerolog/pull/440)
	LogLevelZLog zerolog.Level // read-only

	ConfigFile string
	Port       string

	// local
	Testing bool

	// DB
	DBUrl     string
	TestDBUrl string

	// Server Timeouts
	WriteTimeOut time.Duration
	ReadTimeOut  time.Duration
	IdleTimeOut  time.Duration
}

func getEnv(name string, defaultValue string) string {
	value, found := os.LookupEnv(name)
	if !found {
		return defaultValue
	}
	return value
}

func New() (*Config, error) {
	runtimeEnvironment := getEnv("RUNTIME_ENVIRONMENT", "local")
	if runtimeEnvironment == "cloud" {
		requiredEnvs := []string{"APPENV", "CONFIG_FILE", "SB_PROCESS", "SB_PROJECT"}
		for _, requiredEnv := range requiredEnvs {
			if _, found := os.LookupEnv(requiredEnv); !found {
				return nil, errors.Errorf("required env var %s not found", requiredEnv)
			}
		}
	}
	appenv := getEnv("APPENV", "local-appenv")
	_, filePath, _, _ := runtime.Caller(0)
	configFilepath := getEnv("CONFIG_FILE", filepath.Join(filepath.Dir(filePath), "..", "config.local.toml"))
	port := getEnv("PORT", "5000")
	sbProcess := getEnv("SB_PROCESS", "main")
	sbProject := getEnv("SB_PROJECT", "simplebank")
	testingStr := getEnv("TESTING", "false")
	version := getEnv("VERSION", "0.0.0")

	configStr, err := ioutil.ReadFile(configFilepath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, errors.Errorf("config file path %s doesn't exist", configFilepath)
		}
		return nil, errors.Wrapf(err, "unable to read config file")
	}
	config := Config{}
	err = toml.Unmarshal(configStr, &config)
	if err != nil {
		return nil, errors.Wrapf(err, "unable to read toml config")
	}
	config.Appenv = appenv
	config.ConfigFile = configFilepath
	config.SbProcess = sbProcess
	config.SbProject = sbProject
	config.Port = port
	config.RuntimeEnvironment = runtimeEnvironment
	config.Version = version

	config.Testing = strings.ToLower(testingStr) == "true"

	// defaults

	if config.WriteTimeOut == 0 {
		config.WriteTimeOut = time.Second * 30
	}

	if config.ReadTimeOut == 0 {
		config.ReadTimeOut = time.Second * 30
	}

	if config.IdleTimeOut == 0 {
		config.IdleTimeOut = time.Second * 60
	}

	config.InitDefaults()

	return &config, nil
}

func (c *Config) InitDefaults() {
	c.Appenv = getEnv("APPENV", "local-appenv")
	c.Version = getEnv("VERSION", "0.0.0")
	c.SbProcess = getEnv("SB_PROCESS", "main")
	c.SbProject = getEnv("SB_PROJECT", "unknown")
	c.RuntimeEnvironment = getEnv("RUNTIME_ENVIRONMENT", "local")

	if c.LogLevel == "" {
		c.LogLevel = "INFO"
	}
	if parsed, err := zerolog.ParseLevel(c.LogLevel); err != nil {
		c.LogLevelZLog = zerolog.InfoLevel
	} else {
		c.LogLevelZLog = parsed
	}
}
