package main

import (
	"os"

	"github.com/tsouza/zmqbeat/cmd"

	_ "github.com/tsouza/zmqbeat/include"
)

func main() {
	if err := cmd.RootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
