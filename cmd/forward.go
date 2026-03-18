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
	"github.com/charmbracelet/lipgloss"
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
			fmt.Println(styleZinc.Render("● Session ended"))
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

			printForwardHeader(selected, port)
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
			fmt.Println(styleZinc.Render("● Session ended"))
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

	timestampPart := styleZinc.Render("[" + time.Now().Format("15:04:05") + "]")
	methodPart := styleViolet.Render(fmt.Sprintf("%-7s", method))

	const statusWidth = 30
	const sizeWidth = 8

	var statusPart string
	if err != nil {
		msg := err.Error()
		if len(msg) > 30 {
			msg = msg[:30]
		}
		statusPart = styleRed.Render(fmt.Sprintf("%-*s", statusWidth, "ERR "+msg))
	} else {
		txt := statusText(status)
		raw := fmt.Sprintf("%d %s", status, txt)
		if txt == "" {
			raw = fmt.Sprintf("%d", status)
		}
		padded := fmt.Sprintf("%-*s", statusWidth, raw)
		switch {
		case status >= 200 && status < 300:
			statusPart = styleGreen.Render(padded)
		case status >= 400 && status < 500:
			statusPart = styleAmber.Render(padded)
		case status >= 500 && status < 600:
			statusPart = styleRed.Render(padded)
		default:
			statusPart = styleWhite.Render(padded)
		}
	}

	sizePart := styleZinc.Render(fmt.Sprintf("%*s", sizeWidth, fmt.Sprintf("%dB", req.Size)))

	fmt.Println(timestampPart + " " + methodPart + " " + statusPart + " " + sizePart)
	fmt.Println(styleZinc.Render("         └─ ") + styleWhite.Render(fmt.Sprintf("localhost:%d", port)))
}

func handleReconnect(ctx context.Context, attempts *int) error {
	*attempts++
	if *attempts > 5 {
		return fmt.Errorf("max reconnect attempts reached (5/5)")
	}
	fmt.Println(styleAmber.Render(fmt.Sprintf("⚠ Connection lost. Reconnecting (%d/5)...", *attempts)))
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

func printForwardHeader(selected api.Endpoint, port int) {
	const boxWidth = 45
	const webhookMax = 40

	webhookURL := selected.WebhookURL
	if len(webhookURL) > webhookMax {
		webhookURL = webhookURL[:webhookMax-3] + "..."
	}

	titleLeft := styleWhite.Copy().Bold(true).Render("hookscope")
	titleRight := styleZinc.Render("v" + rootCmd.Version)
	titleLine := lipgloss.NewStyle().Width(boxWidth - 2).Render(
		lipgloss.JoinHorizontal(lipgloss.Top,
			titleLeft,
			lipgloss.NewStyle().Width(boxWidth-2-lipgloss.Width(titleLeft)-lipgloss.Width(titleRight)).Render(""),
			titleRight,
		),
	)

	row := func(label, value string) string {
		lbl := styleZinc.Render(fmt.Sprintf("%-11s", label))
		val := styleWhite.Render(value)
		return lipgloss.NewStyle().Width(boxWidth - 2).Render(lbl + val)
	}

	statusRow := styleZinc.Render(fmt.Sprintf("%-11s", "Status")) + styleGreen.Render("online")
	statusLine := lipgloss.NewStyle().Width(boxWidth - 2).Render(statusRow)

	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		Padding(0, 1).
		Width(boxWidth).
		Render(
			lipgloss.JoinVertical(lipgloss.Left,
				titleLine,
				"",
				row("Endpoint", selected.Name),
				row("Webhook URL", webhookURL),
				row("Forwarding", fmt.Sprintf("http://localhost:%d", port)),
				statusLine,
			),
		)

	fmt.Println(box)
	fmt.Println(styleZinc.Render("Ctrl+C to stop"))
	fmt.Println(styleZinc.Render(strings.Repeat("─", boxWidth)))
	fmt.Println()
}

func statusText(status int) string {
	statusMap := map[int]string{
		200: "OK",
		201: "Created",
		204: "No Content",

		400: "Bad Request",
		401: "Unauthorized",
		403: "Forbidden",
		404: "Not Found",
		422: "Unprocessable",
		429: "Too Many Requests",

		500: "Internal Server Error",
		502: "Bad Gateway",
		503: "Service Unavailable",
	}
	return statusMap[status]
}
