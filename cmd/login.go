package cmd

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"

	"github.com/NowackiKuba/hookscope-cli/internal/auth"
	"github.com/spf13/cobra"
)

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Authenticate and save token locally",
	RunE:  runLogin,
}

func init() {
	rootCmd.AddCommand(loginCmd)
}

func runLogin(cmd *cobra.Command, args []string) error {
	token := generateFakeToken()
	if err := auth.SaveToken(token); err != nil {
		return fmt.Errorf("save token: %w", err)
	}
	fmt.Println("Logged in successfully. Token saved.")
	return nil
}

func generateFakeToken() string {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "fake-token-fallback"
	}
	return hex.EncodeToString(b)
}
