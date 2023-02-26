package utils

import "strings"

func LinkToID(s string) string {
	TrackID := ""
	if strings.HasPrefix(s, "https://open.spotify.com/track/") {
		TrackID = strings.TrimPrefix(s, "https://open.spotify.com/track/")
		if strings.Contains(TrackID, "?") {
			TrackID = TrackID[:strings.Index(TrackID, "?")]
		}
	}
	return TrackID
}
