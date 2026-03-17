package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var rootCmd = &cobra.Command{
	Use:   "hookscope",
	Short: "Hookscope CLI",
	Long:  "CLI tool for exposing local servers via Hookscope",
}

func init() {
	viper.SetConfigName("config")
	viper.AddConfigPath("$HOME/.hookscope")
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}