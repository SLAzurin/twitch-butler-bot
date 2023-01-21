package data

type ApiStreams struct {
	Data []struct {
		StartedAt string `json:"started_at"`
		UserName  string `json:"user_name"`
	}
}
