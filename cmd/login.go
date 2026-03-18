package cmd

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/NowackiKuba/hookscope-cli/internal/auth"
	"github.com/NowackiKuba/hookscope-cli/internal/config"
	"github.com/spf13/cobra"
)

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Authenticate with Hookscope",
	RunE:  runLogin,
}

func init() {
	rootCmd.AddCommand(loginCmd)
}

func openBrowser(url string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", url)
	default:
		cmd = exec.Command("xdg-open", url)
	}
	return cmd.Start()
}

func runLogin(cmd *cobra.Command, args []string) error {
	url := "https://app.hookscope.dev/settings/#cli-token"
	_ = openBrowser(url)
	fmt.Println(styleZinc.Render("Opening Hookscope in your browser..."))
	fmt.Println(styleZinc.Render("If browser didn't open, visit: ") + styleWhite.Render(url))
	fmt.Println()

	fmt.Println(styleWhite.Render("Paste your CLI token:"))

	scanner := bufio.NewScanner(os.Stdin)
	if !scanner.Scan() {
		if err := scanner.Err(); err != nil {
			return err
		}
		return fmt.Errorf("no token provided")
	}
	token := strings.TrimSpace(scanner.Text())
	if !strings.HasPrefix(token, "cli_") || len(token) <= 10 {
		return fmt.Errorf("%s", styleRed.Render("Invalid token. Expected something like cli_abc123..."))
	}

	if err := auth.Save(auth.Credentials{Token: token, APIURL: config.DefaultAPIURL}); err != nil {
		return fmt.Errorf("save credentials: %w", err)
	}

	fmt.Println(styleGreen.Render("✓ Logged in successfully"))
	return nil
}
