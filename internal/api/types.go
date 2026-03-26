package api

type Endpoint struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	WebhookURL string `json:"webhookUrl"`
	Token      string `json:"token"`
	TargetUrl string `json:"targetUrl"`
}

type WebhookRequest struct {
	ID          string            `json:"id"`
	EndpointID  string            `json:"endpointId"`
	Method      string            `json:"method"`
	Path        string            `json:"path"`
	Headers     map[string]string `json:"headers"`
	Body        any               `json:"body"`
	Query       map[string]string `json:"query"`
	IP          string            `json:"ip"`
	ContentType string            `json:"contentType"`
	Size        int               `json:"size"`
	Overlimit   bool              `json:"overlimit"`
	ReceivedAt  string            `json:"receivedAt"`
}