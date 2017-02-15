package main

import (
	"log"
	"os"
	"time"

	"github.com/quipo/statsd"
	gracefully "github.com/tj/go-gracefully"
	poller "github.com/travisjeffery/burrow-stats/poller"
	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

var (
	burrowAddr  = kingpin.Flag("burrow-addr", "burrow address").Required().String()
	statsAddr   = kingpin.Flag("stats-addr", "stats address").Required().String()
	statsPrefix = kingpin.Flag("stats-prefix", "stats prefix").Required().String()
)

func main() {
	kingpin.Parse()

	stats := statsd.NewStatsdClient(*statsAddr, *statsPrefix)
	if err := stats.CreateSocket(); err != nil {
		panic(err)
	}

	logger := log.New(os.Stdout, "burrow-stat", log.LstdFlags)

	b, err := poller.New(*burrowAddr, logger, stats)
	if err != nil {
		panic(err)
	}

	go b.Start()

	gracefully.Timeout = time.Second * 5
	gracefully.Shutdown()

	b.Stop()
}
