package beater

import (
	"bytes"
	"fmt"
	"time"

	"github.com/elastic/beats/libbeat/beat"
	"github.com/elastic/beats/libbeat/common"
	"github.com/elastic/beats/libbeat/logp"

	"github.com/tsouza/zmqbeat/config"

	"github.com/zeromq/goczmq"
)

// Zmqbeat configuration.
type Zmqbeat struct {
	done   chan struct{}
	config config.Config
	client beat.Client

	pull *goczmq.Channeler
}

// New creates an instance of zmqbeat.
func New(b *beat.Beat, cfg *common.Config) (beat.Beater, error) {
	c := config.DefaultConfig
	if err := cfg.Unpack(&c); err != nil {
		return nil, fmt.Errorf("Error reading config file: %v", err)
	}

	bt := &Zmqbeat{
		done:   make(chan struct{}),
		config: c,
	}
	return bt, nil
}

// Run starts zmqbeat.
func (bt *Zmqbeat) Run(b *beat.Beat) error {
	logp.Info("zmqbeat is running! Hit CTRL-C to stop it.")

	var err error
	bt.client, err = b.Publisher.Connect()
	if err != nil {
		return err
	}

	bt.pull = goczmq.NewPullChanneler(bt.config.Bind)

	for {
		select {
		case msg := <-bt.pull.RecvChan:
			var body bytes.Buffer
			for _, part := range msg {
				body.WriteString(string(part))
			}
			event := beat.Event{
				Timestamp: time.Now(),
				Fields: common.MapStr{
					"test": body.String(),
				},
			}
			bt.client.Publish(event)
		}
	}

	/*bt.ctx, err = zmq.NewContext()
	if err != nil {
		panic(err)
	}

	bt.sock, err = bt.ctx.Socket(zmq.Pull)
	if err != nil {
		panic(err)
	}

	bt.chans = bt.sock.Channels()

	if err = bt.sock.Connect(bt.config.Connect); err != nil {
		panic(err)
	}

	for {
		select {
		case msg := <-bt.chans.In():
			var body bytes.Buffer
			for _, part := range msg {
				body.WriteString(string(part))
			}
			event := beat.Event{
				Timestamp: time.Now(),
				Fields: common.MapStr{
					"test": body.String(),
				},
			}
			bt.client.Publish(event)
		case err := <-bt.chans.Errors():
			panic(err)
		}
	}*/

	/*ticker := time.NewTicker(bt.config.Period)
	counter := 1
	for {
		select {
		case <-bt.done:
			return nil
		case <-ticker.C:
		}

		event := beat.Event{
			Timestamp: time.Now(),
			Fields: common.MapStr{
				"type":    b.Info.Name,
				"counter": counter,
			},
		}
		bt.client.Publish(event)
		logp.Info("Event sent")
		counter++
	}*/
}

// Stop stops zmqbeat.
func (bt *Zmqbeat) Stop() {
	if bt.pull != nil {
		bt.pull.Destroy()
	}
	/*if bt.chans != nil {
		bt.chans.Close()
	}
	if bt.sock != nil {
		bt.sock.Close()
	}
	if bt.ctx != nil {
		bt.ctx.Close()
	}
	bt.client.Close()*/
	close(bt.done)
}
