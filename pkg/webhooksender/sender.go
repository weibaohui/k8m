package webhooksender

// SendResult holds the result of a webhook send attempt.
type SendResult struct {
	Status     string // success / failed
	StatusCode int
	RespBody   string
}

// Sender defines the webhook adapter interface.
type Sender interface {
	Name() string
	Send(event *InspectionCheckEvent, receiver *WebhookReceiver) (*SendResult, error)
}

// senderRegistry holds all registered senders.
var senderRegistry = map[string]Sender{}

// RegisterSender registers a new platform sender.
func RegisterSender(platform string, sender Sender) {
	senderRegistry[platform] = sender
}

// GetSender returns the appropriate sender.
func GetSender(platform string) Sender {
	if s, ok := senderRegistry[platform]; ok {
		return s
	}
	return &DefaultSender{}
}
