package main

import (
	"flag"
	"time"

	"github.com/potehinre/best_scraper/api"
	"github.com/potehinre/best_scraper/availability"
	"github.com/potehinre/best_scraper/config"
)

func main() {
	var configPath = flag.String("configPath", "config.toml", "path to config")
	flag.Parse()
	config.Read(*configPath)
	availability.Check(&config.Conf.AvailabilityCheck)
	ticker := time.NewTicker(time.Duration(config.Conf.AvailabilityCheck.CheckPerMinutes) * time.Minute)
	go func() {
		for _ = range ticker.C {
			availability.Check(&config.Conf.AvailabilityCheck)
		}
	}()
	api.Init(config.Conf)
}
