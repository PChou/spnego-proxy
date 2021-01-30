package main

import (
	"gopkg.in/yaml.v2"
	"io/ioutil"
)

type SpnegoConfig struct {
	Krb5 string `yaml:"krb5"`
	Client struct {
		Principal string `yaml:"principal"`
		Keytab string `yaml:"keytab"`
	}
	Server struct {
		Principal string `yaml:"principal"`
		Upstream string `yaml:"upstream"`
		Listen string `yaml:"listen"`
	}
}

func loadConfig(path string) SpnegoConfig {
	config := SpnegoConfig{}
	content, err := ioutil.ReadFile(path)
	if err != nil {
		panic(err)
	}
	err = yaml.Unmarshal(content, &config)
	if err != nil {
		panic(err)
	}
	return config
}

func (sc SpnegoConfig) checkValid() {
	if sc.Krb5 == "" || sc.Client.Principal == "" || sc.Client.Keytab == "" ||
		sc.Server.Upstream == "" || sc.Server.Listen == ""{
		panic("Invalid config")
	}
}