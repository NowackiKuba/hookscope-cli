package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/charmbracelet/huh"
	"github.com/NowackiKuba/hookscope-cli/internal/auth"
	"github.com/NowackiKuba/hookscope-cli/internal/api"
	"github.com/NowackiKuba/hookscope-cli/internal/tunnel"
	"github.com/spf13/cobra"
)

var endpointIDFlag string

var forwardCmd = &cobra.Command{
	Use:   "forward [port]",
	Short: "Forward webhooks to localhost",
	RunE:  runForward,
}

func init() {
	rootCmd.AddCommand(forwardCmd)
	forwardCmd.Flags().StringVarP(&endpointIDFlag, "endpoint", "e", "", "Endpoint ID to forward (skip selection)")
}

func runForward(cmd *cobra.Command, args []string) error {
	port := 3000
	if len(args) >= 1 && stringsTrim(args[0]) != "" {
		p, err := strconv.Atoi(args[0])
		if err != nil || p <= 0 || p > 65535 {
			return fmt.Errorf("invalid port: %q", args[0])
		}
		port = p
	}

	creds, err := auth.Load()
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("run hookscope login first")
		}
		return fmt.Errorf("load credentials: %w", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	var selected api.Endpoint
	var selectedID string
	reconnectAttempts := 0

	for {
		if ctx.Err() != nil {
			fmt.Println(styleZinc.Render("Disconnected"))
			return nil
		}

		t := tunnel.New(creds.APIURL, creds.Token, port)
		if err := t.Connect(ctx); err != nil {
			if err := handleReconnect(ctx, &reconnectAttempts); err != nil {
				return err
			}
			continue
		}

		endpoints, err := t.Authenticate(ctx)
		if err != nil {
			t.Close()
			return fmt.Errorf("auth failed: %w", err)
		}

		if selectedID == "" {
			ep, id, err := chooseEndpoint(endpoints, endpointIDFlag)
			if err != nil {
				t.Close()
				return err
			}
			selected = ep
			selectedID = id

			fmt.Println(styleViolet.Render("● Forwarding webhooks"))
			fmt.Println(styleZinc.Render("  Endpoint: ") + styleWhite.Render(selected.Name))
			fmt.Println(styleZinc.Render("  Webhook URL: ") + styleWhite.Render(selected.WebhookURL))
			fmt.Println(styleZinc.Render("  Local: ") + styleWhite.Render(fmt.Sprintf("http://localhost:%d", port)))
			fmt.Println(styleZinc.Render("  Press Ctrl+C to stop"))
		}

		if err := t.Subscribe(ctx, selectedID); err != nil {
			t.Close()
			return fmt.Errorf("subscribe failed: %w", err)
		}

		reconnectAttempts = 0
		listenErr := t.Listen(ctx, port, func(req api.WebhookRequest, status int, err error) {
			printForwardResult(port, req, status, err)
		})
		t.Close()

		if ctx.Err() != nil {
			fmt.Println(styleZinc.Render("Disconnected"))
			return nil
		}
		_ = listenErr

		if err := handleReconnect(ctx, &reconnectAttempts); err != nil {
			return err
		}
	}
}

func chooseEndpoint(endpoints []api.Endpoint, endpointIDFlag string) (api.Endpoint, string, error) {
	if endpointIDFlag != "" {
		for _, ep := range endpoints {
			if ep.ID == endpointIDFlag {
				return ep, ep.ID, nil
			}
		}
		return api.Endpoint{}, "", fmt.Errorf("endpoint not found: %s", endpointIDFlag)
	}
	if len(endpoints) == 0 {
		return api.Endpoint{}, "", fmt.Errorf("no endpoints available for this token")
	}
	if len(endpoints) == 1 {
		return endpoints[0], endpoints[0].ID, nil
	}

	opts := make([]huh.Option[string], 0, len(endpoints))
	for _, ep := range endpoints {
		label := ep.Name
		if label == "" {
			label = ep.ID
		}
		opts = append(opts, huh.NewOption(label, ep.ID))
	}

	var pickedID string
	sel := huh.NewSelect[string]().
		Title("Select an endpoint to forward").
		Options(opts...).
		Value(&pickedID)

	if err := sel.Run(); err != nil {
		return api.Endpoint{}, "", err
	}
	for _, ep := range endpoints {
		if ep.ID == pickedID {
			return ep, ep.ID, nil
		}
	}
	return api.Endpoint{}, "", fmt.Errorf("endpoint not found: %s", pickedID)
}

func printForwardResult(port int, req api.WebhookRequest, status int, err error) {
	method := req.Method
	if method == "" {
		method = "REQUEST"
	}
	methodPart := styleViolet.Render(method)
	destPart := styleZinc.Render(fmt.Sprintf(" → localhost:%d ", port))

	var statusPart string
	if err != nil {
		statusPart = styleRed.Render("[ERR]")
	} else if status >= 200 && status < 400 {
		statusPart = styleGreen.Render(fmt.Sprintf("[%d]", status))
	} else {
		statusPart = styleRed.Render(fmt.Sprintf("[%d]", status))
	}

	sizePart := styleZinc.Render(fmt.Sprintf(" %dB", req.Size))
	fmt.Println(methodPart + destPart + statusPart + sizePart)
}

func handleReconnect(ctx context.Context, attempts *int) error {
	*attempts++
	if *attempts > 5 {
		return fmt.Errorf("max reconnect attempts reached (5/5)")
	}
	fmt.Println(styleZinc.Render(fmt.Sprintf("Reconnecting... (attempt %d/5)", *attempts)))
	timer := time.NewTimer(3 * time.Second)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return nil
	case <-timer.C:
		return nil
	}
}

func stringsTrim(s string) string {
	return strings.TrimSpace(s)
}
