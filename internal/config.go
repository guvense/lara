package lara

import (
	"io/ioutil"
	"log"
	"os"

	"gopkg.in/yaml.v2"
)

type Config struct {
	MocksPath      string               `yaml:"mocks_path"`
	ServerConfig   ServerConfig         `yaml:"server"`
	TokenGenerator TokenGenerator 		`yaml:"token-generator"`
	Watcher bool                  		`yaml:"watcher"`
	RegexExpression map[string]string           `yaml:"regex"`
}

type ServerConfig struct {
	Port int    `yaml:"port"`
	Host string `yaml:"host"`
}

type TokenServers map[string]TokenServerDetail

type TokenGenerator struct {
	TokenServers TokenServers
}

func (b *TokenGenerator) UnmarshalYAML(unmarshal func(interface{}) error) error {
	return unmarshal(&b.TokenServers)
}

type TokenServerDetail struct {
	Type         string `yaml:"type"`
	TokenUrl     string `yaml:"token-url"`
	ClientID     string `yaml:"client-id"`
	ClientSecret string `yaml:"client-secret"`
	Username     string `yaml:"username"`
	Password     string `yaml:"password"`
	Scope        string `yaml:"scope"`
}

func PrepareConfig(host string, port int, watcher bool, mocksPath string, configFilePath string) (Config, error) {

	var config Config

	if configFilePath != "" {
		config, _ = parseConfig(configFilePath)
	} else {
		config = Config{}
		serverConfig := ServerConfig{}
		config.ServerConfig = serverConfig

	}

	if config.ServerConfig.Host == "" || host != "localhost" {
		config.ServerConfig.Host = host
	}

	if config.ServerConfig.Port == 0 ||  port != 8898 {
		config.ServerConfig.Port = port
	}

	if config.MocksPath == "" ||  mocksPath != "/mocs" {
		config.MocksPath = mocksPath
	}

	if  watcher {
		config.Watcher = watcher
	}


	return config, nil

}

func parseConfig(configFilePath string) (Config, error) {

	var config Config

	configFile, err := os.Open(configFilePath)

	if err != nil {
		log.Printf("%v: error occurred with path: %s", err, configFilePath)
		return Config{}, err
	}

	defer configFile.Close()

	bytes, _ := ioutil.ReadAll(configFile)
	if err := yaml.Unmarshal(bytes, &config); err != nil {
		log.Printf("%v: error while unmarshall configFile file %s, using default configuration instead", err, configFilePath)
	}
	return config, nil

}
