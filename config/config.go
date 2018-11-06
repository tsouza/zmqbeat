// Config is put into a different package to prevent cyclic imports in case
// it is needed in several locations

package config

import "time"

// Config is a zmqbeat configuration struct
type Config struct {
	Period time.Duration `config:"period"`
	Bind   string        `config:"bind"`
}

// DefaultConfig is a default zmqbeat Config
var DefaultConfig = Config{
	Period: 1 * time.Second,
	Bind:   "tcp://*:5555",
}
