package main

import (
	"fmt"
	"os"

	socket2stdout "github.com/bakins/socket2stdout"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "socket2stdout",
	Short: "copy lines from a socket to stdout",
	Run:   runServer,
}

var (
	tcpAddr  *string
	unixAddr *string
	auxAddr  *string
)

func main() {
	tcpAddr = rootCmd.PersistentFlags().StringP("addr", "", "127.0.0.1:4444", "tcp address")
	unixAddr = rootCmd.PersistentFlags().StringP("unix", "", "", "unix address. takes precedence if both unix and tcp are set")
	auxAddr = rootCmd.PersistentFlags().StringP("aux-addr", "", ":9090", "listen address for aux handler. metrics, healthchecks, etc")

	if err := rootCmd.Execute(); err != nil {
		panic(err)
	}
}

func runServer(cmd *cobra.Command, args []string) {
	s, err := socket2stdout.New(
		socket2stdout.SetAuxAddress(*auxAddr),
		socket2stdout.SetTCPAddress(*tcpAddr),
		socket2stdout.SetUnixAddress(*unixAddr),
	)

	if err != nil {
		errorLog("failed to create server: %s\n", err)
	}
	if err := s.Run(); err != nil {
		errorLog("unable to run server: %s\n", err)
	}
}

func errorLog(format string, a ...interface{}) {
	fmt.Fprintf(os.Stderr, format, a...)
}
