package domain

import "time"

// Attachment points from a chat message to a file already stored by uploads.
type Attachment struct {
	Name  string `json:"name"`
	Bytes int64  `json:"bytes"`
	URL   string `json:"url"`
}

// Message is one entry in the shared LAN chat.
type Message struct {
	ID          string       `json:"id"`
	TS          time.Time    `json:"ts"`
	DeviceID    string       `json:"deviceId"`
	DisplayName string       `json:"displayName"`
	Text        string       `json:"text"`
	Attachments []Attachment `json:"attachments,omitempty"`
}

// Participant describes a device recently seen by the HTTP chat hub.
type Participant struct {
	DeviceID    string    `json:"deviceId"`
	DisplayName string    `json:"displayName"`
	LastSeen    time.Time `json:"lastSeen"`
}
