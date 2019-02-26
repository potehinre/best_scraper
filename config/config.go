package config

import (
	"io/ioutil"

	"github.com/BurntSushi/toml"
	"github.com/sirupsen/logrus"
)

var Conf *Config

type AvailabilityCheck struct {
	CheckPerMinutes     int
	TimeoutMilliseconds int
	Sites               []string
}

type HTTP struct {
	Address                  string
	WriteTimeoutMilliseconds int
	ReadTimeoutMilliseconds  int
	AuthLogin                string
	AuthPassword             string
}

type Config struct {
	HTTP              HTTP
	AvailabilityCheck AvailabilityCheck
}

func Read(path string) {
	tomlData, err := ioutil.ReadFile(path)
	if err != nil {
		logrus.WithError(err).Fatal("can't read config")
	}
	if _, err := toml.Decode(string(tomlData), &Conf); err != nil {
		logrus.WithError(err).Fatal("incorrect config format")
	}
}
