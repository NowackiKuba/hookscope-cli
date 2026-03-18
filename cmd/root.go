package cmd

import (
	"fmt"
	"os"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "hookscope",
	Short:   "Hookscope CLI — real-time webhook inspector",
	Version: "0.1.0",
}

var (
	styleViolet = lipgloss.NewStyle().Foreground(lipgloss.Color("#7C3AED"))
	styleRed    = lipgloss.NewStyle().Foreground(lipgloss.Color("#EF4444"))
	styleZinc   = lipgloss.NewStyle().Foreground(lipgloss.Color("#71717A"))
	styleWhite  = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF"))
	styleGreen  = lipgloss.NewStyle().Foreground(lipgloss.Color("#22C55E"))
)

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}