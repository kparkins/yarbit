package main

import (
	"fmt"
	"github.com/kparkins/yarbit/node"
	"github.com/spf13/cobra"
	"os"
	"strconv"
	"strings"
)

const flagIp = "ip"
const flagPort = "port"
const flagBootstrap = "bootstrap"

func runCommand() *cobra.Command {
	command := &cobra.Command{
		Use:   "run",
		Short: "Launches the Yarbit node and its HTTP API.",
		Run: func(cmd *cobra.Command, args []string) {
			dataDir, _ := cmd.Flags().GetString(flagDataDir)
			bootstrapNode, _ := cmd.Flags().GetString(flagBootstrap)
			ip, _ := cmd.Flags().GetString(flagIp)
			port, _ := cmd.Flags().GetUint64(flagPort)

			bootstrapIp, bootstrapPort := getBoostrapIpAndPort(bootstrapNode)
			bootstrap := node.PeerNode{
				IpAddress:   bootstrapIp,
				Port:        bootstrapPort,
				IsBootstrap: true,
				IsActive:    true,
			}
			config := node.Config{
				DataDir:      dataDir,
				IpAddress:    ip,
				Port:         port,
				Protocol:     "http",
				Bootstrap:    bootstrap,
				MinerAccount: "kyle",
			}
			server := node.New(config)
			err := server.Run()
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}

		},
	}
	command.Flags().String(flagBootstrap, "", "ip:port of the bootstrap node. If empty, defaults to being the bootstrap node.")
	addDefaultRequiredFlags(command)
	command.Flags().String(flagIp, "127.0.0.1", "the ip of the node")
	command.Flags().Uint64(flagPort, uint64(80), "the port of the node")
	return command
}

func getBoostrapIpAndPort(node string) (string, uint64) {
	parts := strings.Split(node, ":")
	if len(parts) != 2 {
		return "", 0
	}
	port, err := strconv.ParseUint(parts[1], 10, 64)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return "", 0
	}
	return parts[0], port
}
