package Helper

import (
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"os"
)

type Config struct {
	Parameters struct {
		Namespace           string `yaml:"app.namespace"`
		DfProductApiUri     string `yaml:"df.product.api.uri"`
		ConfigPath          string `yaml:"config.path"`
		TokenUri            string `yaml:"google.token.uri"`
		Type                string `yaml:"google.type"`
		AuthProviderCertUrl string `yaml:"google.auth.provider.x509.cert.url"`
		AuthUri             string `yaml:"google.auth.uri"`
		UniverseDomain      string `yaml:"google.universe.domain"`
		ProjectId           string `yaml:"google.project.id"`
		ClientEmail         string `yaml:"google.client.email"`
		ClientCertUrl       string `yaml:"google.client.x509.cert.url"`
		RabbitMqClient      string `yaml:"rabbitmq.client"`
		RabbitMqUserName    string `yaml:"rabbitmq.username"`
		RabbitMqVhost       string `yaml:"rabbitmq.vhost"`
		RedisAddress        string `yaml:"redis.address"`
		DfSessionPrefix     string `yaml:"df.session.prefix"`
		GinServerPort       string `yaml:"gin.server.port"`
		GinMode             string `yaml:"gin.mode"`
		IsCloudStorage      string `yaml:"is.cloud.storage"`
		CoreMarketApiUri    string `yaml:"core.market.api.uri"`
	} `yaml:"parameters"`
}

const DEFAULT_PARAMETER = "config/parameters/all.yaml"
const PARAMETER_PATH = "config/parameters/"

func GetParameter() (Config, error) {
	var config Config

	data, err := ioutil.ReadFile(DEFAULT_PARAMETER)
	if err != nil {
		return config, err
	}

	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return config, err
	}

	appEnv := os.Getenv("APP_ENV")
	filePath := PARAMETER_PATH + appEnv + ".yaml"
	data, err = ioutil.ReadFile(filePath)
	if err != nil {
		return config, err
	}

	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return config, err
	}

	return config, nil
}
