package main

import (
	"github.com/magiconair/properties"
	log "github.com/sirupsen/logrus"
	"io"
	"math/rand"
	"mosquitoSwarm/src/config"
	"mosquitoSwarm/src/util"
	"os"
	"time"
)

func main() {
	log.Info("Program startup")
	rand.Seed(time.Now().UnixMilli())

	//load configs
	p := properties.MustLoadFile("config.properties", properties.UTF8)
	var cfg config.Config
	if err := p.Decode(&cfg); err != nil {
		log.WithError(err).Fatal("Failed to parse configs")
	}

	loc, _ := time.LoadLocation(cfg.TimeZone)

	//create a log file
	file, err := os.Create("mosquitoSwarm.log")
	if err != nil {
		log.WithError(err).Fatal("failed to even create log file, what's the point now...")
	}
	log.SetOutput(io.MultiWriter(os.Stdout, file))
	log.SetFormatter(util.LogFormatter{Formatter: &log.TextFormatter{TimestampFormat: "2006-01-02 15:04:05"}, Loc: loc})

	initializeThings(&cfg)
	scheduleJobs(&cfg, loc)

	//wait indefinitely
	select {}
}
