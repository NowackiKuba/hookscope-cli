package cmd

import (
	"fmt"

	"github.com/NowackiKuba/hookscope-cli/internal/auth"
	"github.com/NowackiKuba/hookscope-cli/internal/tunnel"
	"github.com/spf13/cobra"
)

var port int

var forwardCmd = &cobra.Command{
	Use:   "forward",
	Short: "Start a tunnel forwarding local port",
	RunE:  runForward,
}

func init() {
	rootCmd.AddCommand(forwardCmd)
	forwardCmd.Flags().IntVarP(&port, "port", "p", 3000, "Local port to forward")
}

func runForward(cmd *cobra.Command, args []string) error {
	token, err := auth.LoadToken()
	if err != nil {
		return fmt.Errorf("load token: %w", err)
	}
	if token == "" {
		fmt.Println("Please login first")
		return nil
	}
	tunnel.StartTunnel(port)
	return nil
}
