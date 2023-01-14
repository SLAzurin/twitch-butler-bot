package data

type EventSubMessage struct {
	Metadata struct {
		MessageID        string `json:"message_id"`
		MessageType      string `json:"message_type"`
		MessageTimestamp string `json:"message_timestamp"`
	} `json:"metadata"`
	Payload struct {
		Session struct {
			ID                      string  `json:"id"`
			Status                  string  `json:"status"`
			ConnectedAt             string  `json:"connected_at"`
			KeepaliveTimeoutSeconds int     `json:"keepalive_timeout_seconds"`
			ReconnectURL            *string `json:"reconnect_url"`
		} `json:"session"`
	} `json:"payload"`
}
