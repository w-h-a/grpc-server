package agent

import (
	"fmt"
)

type Config struct {
	RPCHost string
	RPCPort  int
}

func (c Config) RPCAddr() (string, error) {
	return fmt.Sprintf("%s:%d", c.RPCHost, c.RPCPort), nil
}
