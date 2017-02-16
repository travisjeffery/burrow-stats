package poller

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/quipo/statsd"
	burrow "github.com/segmentio/go-burrow"
)

type Poller struct {
	tick   *time.Ticker
	wg     sync.WaitGroup
	stats  *statsd.StatsdClient
	logger *log.Logger
	*burrow.Client
}

func New(url string, logger *log.Logger, stats *statsd.StatsdClient) (*Poller, error) {
	c, err := burrow.New(url)
	if err != nil {
		return nil, err
	}
	return &Poller{
		tick:   time.NewTicker(10 * time.Second),
		stats:  stats,
		logger: logger,
		Client: c,
	}, nil
}

func (p *Poller) Start() {
	for range p.tick.C {
		if err := p.fetch(); err != nil {
			p.stats.Incr("burrow.error", 1)
			log.Printf("[ERROR] check: %v", err)
			continue
		}
		log.Printf("[INFO] fetched")
	}
}

func (p *Poller) Stop() {
	p.tick.Stop()
	p.wg.Wait()
}

func (p *Poller) fetch() error {
	p.wg.Add(1)
	defer p.wg.Done()

	cluster, err := p.Clusters()
	if err != nil {
		return err
	}

	for _, cluster := range cluster.Names {
		consumers, err := p.Consumers(cluster)
		if err != nil {
			return err
		}

		for _, group := range consumers.Names {
			lag, err := p.ConsumerLag(cluster, group)
			if err != nil {
				return err
			}

			p.stats.Gauge(fmt.Sprintf("consumer.lag,cluster=%s,consumer_group=%s,topic=%s", lag.Cluster, lag.Group, lag.MaxLag.Topic), int64(lag.TotalLag))

			for _, partition := range lag.Partitions {
				p.stats.Gauge(fmt.Sprintf("consumer.lag,cluster=%s,consumer_group=%s,topic=%s,partition=%d", lag.Cluster, lag.Group, partition.Topic, partition.Partition), int64(partition.Start.Lag))
			}

			p.logger.Printf("[INFO] lag: %v", lag)
		}
	}

	return nil
}
