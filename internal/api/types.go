package api

type Endpoint struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	WebhookURL string `json:"webhookUrl"`
	Token      string `json:"token"`
}

type WebhookRequest struct {
	ID          string            `json:"id"`
	EndpointID  string            `json:"endpointId"`
	Method      string            `json:"method"`
	Headers     map[string]string `json:"headers"`
	Body        any               `json:"body"`
	Query       map[string]string `json:"query"`
	IP          string            `json:"ip"`
	ContentType string            `json:"contentType"`
	Size        int               `json:"size"`
	Overlimit   bool              `json:"overlimit"`
	ReceivedAt  string            `json:"receivedAt"`
}