package models

type Response struct {
	MessagingType    string  `json:"messaging_type,omitempty"`
	Recipient        User    `json:"recipient,omitempty"`
	Message          Message `json:"message,omitempty"`
	Tag              string  `json:"tag,omitempty"`
	NotificationType string  `json:"notification_type,omitempty"`
}
