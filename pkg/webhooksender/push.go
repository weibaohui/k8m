package webhooksender

// PushEvent sends the event to a list of receivers.
func PushEvent(event *InspectionCheckEvent, receivers []*WebhookReceiver) []SendResult {
	results := make([]SendResult, 0, len(receivers))
	for _, r := range receivers {
		sender := GetSender(r.Platform)
		res, err := sender.Send(event, r)
		if err != nil {
			results = append(results, SendResult{Status: "failed", RespBody: err.Error()})
		} else {
			results = append(results, *res)
		}
	}
	return results
}
