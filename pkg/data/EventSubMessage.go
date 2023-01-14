package data

type EventSubMessage struct {
	Metadata struct {
		MessageID        string `json:"message_id"`
		MessageType      string `json:"message_type"`
		MessageTimestamp string `json:"message_timestamp"`
		SubscriptionType string `json:"subscription_type"`
	} `json:"metadata"`
	Payload struct {
		Session struct {
			ID                      string  `json:"id"`
			Status                  string  `json:"status"`
			ConnectedAt             string  `json:"connected_at"`
			KeepaliveTimeoutSeconds int     `json:"keepalive_timeout_seconds"`
			ReconnectURL            *string `json:"reconnect_url"`
		} `json:"session"`
		Event struct {
			UserLogin          string `json:"user_login"`
			ModeratorUserLogin string `json:"moderator_user_login"`
			IsPermanant        bool   `json:"is_permanent"`
		} `json:"event"`
	} `json:"payload"`
}
