package main

import (
	"fmt"
	"github.com/kparkins/yarbit/node"
	"github.com/spf13/cobra"
	"os"
	"strconv"
	"strings"
)

const flagBootstrap = "bootstrap"

func runCommand() *cobra.Command {
	command := &cobra.Command{
		Use:   "run",
		Short: "Launches the Yarbit node and its HTTP API.",
		Run: func(cmd *cobra.Command, args []string) {
			dataDir, err := cmd.Flags().GetString(flagDataDir)
			bootstrapNode, err := cmd.Flags().GetString(flagBootstrap)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
			ip, port := getBoostrapIpAndPort(bootstrapNode)
			bootstrap := node.PeerNode{
				IpAddress:   ip,
				Port:        port,
				IsBootstrap: true,
				IsActive:    true,
			}
			server := node.New(dataDir, 80, bootstrap)
			err = server.Run()
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}

		},
	}
	command.Flags().String(flagBootstrap, "", "ip:port of the bootstrap node. If empty, defaults to being the bootstrap node.")
	addDefaultRequiredFlags(command)
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
