package beater

import (
	"errors"
	"fmt"
	"time"

	"github.com/elastic/beats/libbeat/beat"
	"github.com/elastic/beats/libbeat/common"

	"github.com/tsouza/zmqbeat/config"

	zmq "github.com/pebbe/zmq4"
)

// Zmqbeat configuration.
type Zmqbeat struct {
	done      chan struct{}
	config    config.Config
	client    beat.Client
	closing   bool
	receivers []*zmq.Socket
}

// New creates an instance of zmqbeat.
func New(b *beat.Beat, cfg *common.Config) (beat.Beater, error) {
	c := config.DefaultConfig
	if err := cfg.Unpack(&c); err != nil {
		return nil, fmt.Errorf("Error reading config file: %v", err)
	}

	bt := &Zmqbeat{
		done:      make(chan struct{}),
		config:    c,
		closing:   false,
		receivers: make([]*zmq.Socket, len(c.Pull)),
	}
	return bt, nil
}

// Run starts zmqbeat.
func (bt *Zmqbeat) Run(b *beat.Beat) error {
	var err error
	bt.client, err = b.Publisher.Connect()
	if err != nil {
		return err
	}

	events := make(chan beat.Event)

	for idx, pullConfig := range bt.config.Pull {
		var receiver *zmq.Socket
		if (pullConfig.Connect == "" && pullConfig.Bind == "") ||
			(pullConfig.Connect != "" && pullConfig.Bind != "") {
			return errors.New("Either 'connect' or 'bind' must be configured")
		}
		receiver, err = zmq.NewSocket(zmq.PULL)
		if err != nil {
			return err
		}
		if pullConfig.Connect != "" {
			receiver.Connect(pullConfig.Connect)
		} else {
			receiver.Bind(pullConfig.Bind)
		}
		tags := pullConfig.Tags
		bt.receivers[idx] = receiver
		go func() {
			for {
				data, err := receiver.Recv(0)
				if err != nil {
					return
				}
				event := beat.Event{
					Timestamp: time.Now(),
					Fields: common.MapStr{
						"data": data,
						"tags": tags,
					},
				}
				events <- event
			}
		}()
	}

	for {
		select {
		case event := <-events:
			bt.client.Publish(event)
		case <-bt.done:
			return nil
		}
	}
}

// Stop stops zmqbeat.
func (bt *Zmqbeat) Stop() {
	bt.closing = true
	for _, receiver := range bt.receivers {
		receiver.Close()
	}
	close(bt.done)
}
