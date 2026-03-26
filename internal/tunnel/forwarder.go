package tunnel

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/NowackiKuba/hookscope-cli/internal/api"
)

// cli_ebaedf96ca3fda5d548cbf3051db17dfa94b409348997d2b92496332af5c9813

func Forward(localPort int, req api.WebhookRequest, endpointPath string) (status int, err error) {
	 fmt.Printf("DEBUG rawBody len: %d\n", len(req.RawBody))
    fmt.Printf("DEBUG body nil: %v\n", req.Body == nil)
	path := req.Path
	if path == "" {
		path = endpointPath
	}
	if path == "" {
		path = "/"
	}
	if path[0] != '/' {
		path = "/" + path
	}
	url := fmt.Sprintf("http://localhost:%d%s", localPort, path)

	var bodyReader *bytes.Reader
	if len(req.RawBody) > 0 {
		bodyReader = bytes.NewReader(req.RawBody)
	} else if req.Body != nil {
		b, mErr := json.Marshal(req.Body)
		if mErr != nil {
			return 0, fmt.Errorf("marshal body: %w", mErr)
		}
		bodyReader = bytes.NewReader(b)
	} else {
		bodyReader = bytes.NewReader(nil)
	}

	httpReq, err := http.NewRequest(req.Method, url, bodyReader)
	if err != nil {
		return 0, fmt.Errorf("create request: %w", err)
	}

	for k, v := range req.Headers {
		if k == "" {
			continue
		}
		if isHopByHopHeader(k) {
			continue
		}
		httpReq.Header.Set(k, v)
	}
	if req.ContentType != "" && httpReq.Header.Get("Content-Type") == "" {
		httpReq.Header.Set("Content-Type", req.ContentType)
	}

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	return resp.StatusCode, nil
}

func isHopByHopHeader(k string) bool {
	switch strings.ToLower(strings.TrimSpace(k)) {
	case "host", "content-length", "transfer-encoding", "connection":
		return true
	default:
		return false
	}
}
