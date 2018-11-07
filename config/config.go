// Config is put into a different package to prevent cyclic imports in case
// it is needed in several locations

package config

// Config is a zmqbeat configuration
type Config struct {
	Pull []PullConfig `config:"pull"`
}

// PullConfig is a zmq pull configuration
type PullConfig struct {
	Bind    string `config:"bind"`
	Connect string `config:"connect"`
	Tags    []string
}

// DefaultConfig is a default zmqbeat Config
var DefaultConfig = Config{
	Pull: []PullConfig{
		PullConfig{
			Connect: "tcp://localhost:5556",
			Tags:    []string{},
		},
	},
}
