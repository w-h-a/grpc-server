package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/w-h-a/grpc-server/pkg/agent"
)

type cli struct {
	cfg cfg
}

type cfg struct {
	agent agent.Config
}

func main() {
	cli := &cli{}

	cmd := &cobra.Command{
		Use:     "serve",
		PreRunE: cli.setupConfig,
		RunE:    cli.run,
	}

	err := setupFlags(cmd)
	if err != nil {
		log.Fatal(err)
	}

	err = cmd.Execute()
	if err != nil {
		log.Fatal(err)
	}
}

func setupFlags(cmd *cobra.Command) error {
	cmd.Flags().String("rpc-host", "127.0.0.1", "Host for RPC client connections.")

	cmd.Flags().Int("rpc-port", 8400, "Port for RPC client connections.")

	return viper.BindPFlags(cmd.Flags())
}

func (c *cli) setupConfig(cmd *cobra.Command, args []string) error {
	c.cfg.agent.RPCHost = viper.GetString("rpc-host")

	c.cfg.agent.RPCPort = viper.GetInt("rpc-port")

	return nil
}

func (c *cli) run(cmd *cobra.Command, args []string) error {
	agent, err := agent.NewAgent(c.cfg.agent)
	if err != nil {
		return err
	}

	fmt.Println(agent.Config.RPCAddr())

	sigChan := make(chan os.Signal, 1)

	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	<- sigChan

	return agent.Shutdown()
}