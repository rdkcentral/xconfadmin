package common

import (
	"fmt"
	"os"

	"github.com/go-akka/configuration"
	log "github.com/sirupsen/logrus"
)

type ServerConfig struct {
	*configuration.Config
	configBytes []byte
}

func NewServerConfig(configFile string) (*ServerConfig, error) {
	configBytes, err := os.ReadFile(configFile)
	if err != nil {
		return nil, err
	}
	conf := configuration.ParseString(string(configBytes))
	return &ServerConfig{
		Config:      conf,
		configBytes: configBytes,
	}, nil
}

func (c *ServerConfig) ConfigBytes() []byte {
	return c.configBytes
}

func ServerOriginId() string {
	if hostname, err := os.Hostname(); err != nil {
		log.Errorf("ERROR getting host name: %v", err)
		return fmt.Sprintf("%d", os.Getpid())
	} else {
		return fmt.Sprintf("%s:%d", hostname, os.Getpid())
	}
}
